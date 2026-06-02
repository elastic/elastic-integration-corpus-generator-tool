// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License 2.0;
// you may not use this file except in compliance with the Elastic License 2.0.

package genlib

// Regression test for fbxj-vtfy-yrbl-mlyw:
// NewGenerator panics with "interface conversion: interface {} is nil, not
// genlib.emitFNotReturn" when a flattened-type field without object_type is
// present and multiple goroutines share the same fields slice whose backing
// array has excess capacity (the common output of normaliseFields when wildcard
// fields are removed).
//
// Root cause: in newGeneratorWithCustomTemplate, `fields = append(fields,
// objectKeysField...)` could write into the shared backing array when cap >
// len, creating a data race. Concurrently running goroutines overwrite each
// other's objectKeysField slots, so a goroutine may bind a field name that
// differs from the name it wrote into the template, leaving fieldMap without
// an entry for that template token and causing the nil type-assertion panic.
//
// Trigger: auditd_manager/auditd v1.20.0, field auditd.paths (type=flattened)
// combined with concrete auditd.paths.* sub-fields (type=keyword), called from
// multiple goroutines sharing the same fields slice.

import (
	"bytes"
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/elastic/elastic-integration-corpus-generator-tool/pkg/genlib/fields"
)

// auditdPathsFields mirrors the relevant subset of the auditd_manager/auditd
// v1.20.0 field definitions: the flattened parent plus all concrete sub-fields.
// Having both present is what triggers the collision that leads to the panic.
var auditdPathsFields = Fields{
	{Name: "auditd.paths", Type: FieldTypeFlattened},
	{Name: "auditd.paths.dev", Type: FieldTypeKeyword},
	{Name: "auditd.paths.inode", Type: FieldTypeKeyword},
	{Name: "auditd.paths.item", Type: FieldTypeKeyword},
	{Name: "auditd.paths.mode", Type: FieldTypeKeyword},
	{Name: "auditd.paths.name", Type: FieldTypeKeyword},
	{Name: "auditd.paths.nametype", Type: FieldTypeKeyword},
	{Name: "auditd.paths.obj_domain", Type: FieldTypeKeyword},
	{Name: "auditd.paths.obj_level", Type: FieldTypeKeyword},
	{Name: "auditd.paths.obj_role", Type: FieldTypeKeyword},
	{Name: "auditd.paths.obj_type", Type: FieldTypeKeyword},
	{Name: "auditd.paths.obj_user", Type: FieldTypeKeyword},
	{Name: "auditd.paths.ogid", Type: FieldTypeKeyword},
	{Name: "auditd.paths.ouid", Type: FieldTypeKeyword},
	{Name: "auditd.paths.rdev", Type: FieldTypeKeyword},
}

// TestNewGenerator_FlattenedWithSubFields_NoPanic is a regression test for
// fbxj-vtfy-yrbl-mlyw. It uses the exact auditd.paths field configuration from
// the EPR package (flattened parent + concrete sub-fields) and sweeps seeds to
// expose the collision.
func TestNewGenerator_FlattenedWithSubFields_NoPanic(t *testing.T) {
	cfg := Config{}
	const seeds = 10000
	for seed := int64(0); seed < seeds; seed++ {
		seed := seed
		t.Run(fmt.Sprintf("seed=%d", seed), func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("NewGenerator panicked (seed %d): %v", seed, r)
				}
			}()
			g, err := NewGenerator(cfg, auditdPathsFields, 0, WithRandSeed(seed))
			if err != nil {
				t.Errorf("NewGenerator returned error (seed %d): %v", seed, err)
				return
			}
			_ = g.Close()
		})
	}
}

// TestNewGenerator_FlattenedWithSubFields_EmitNoPanic checks that Emit also
// works correctly after construction.
func TestNewGenerator_FlattenedWithSubFields_EmitNoPanic(t *testing.T) {
	cfg := Config{}
	const seeds = 1000
	for seed := int64(0); seed < seeds; seed++ {
		seed := seed
		t.Run(fmt.Sprintf("seed=%d", seed), func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("panicked (seed %d): %v", seed, r)
				}
			}()
			g, err := NewGenerator(cfg, auditdPathsFields, 5, WithRandSeed(seed))
			if err != nil {
				t.Errorf("NewGenerator error (seed %d): %v", seed, err)
				return
			}
			defer func() { _ = g.Close() }()
			var buf bytes.Buffer
			for i := 0; i < 5; i++ {
				buf.Reset()
				if err := g.Emit(&buf); err != nil {
					t.Errorf("Emit error (seed %d, event %d): %v", seed, i, err)
					return
				}
			}
		})
	}
}

// TestNewGenerator_RealEPRFields_NoPanic loads the actual auditd_manager/auditd
// v1.20.0 field definitions from the testdata directory and runs many seeds.
// This is the most faithful reproduction of the production trigger.
func TestNewGenerator_RealEPRFields_NoPanic(t *testing.T) {
	ctx := context.Background()
	flds, err := fields.LoadFieldsWithTemplate(ctx, "testdata/auditd_manager_auditd_v1.20.0_fields.yml")
	if err != nil {
		t.Fatalf("failed to load testdata fields: %v", err)
	}

	cfg := Config{}
	const seeds = 10000
	for seed := int64(0); seed < seeds; seed++ {
		seed := seed
		t.Run(fmt.Sprintf("seed=%d", seed), func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("NewGenerator panicked (seed %d): %v", seed, r)
				}
			}()
			g, err := NewGenerator(cfg, flds, 0, WithRandSeed(seed))
			if err != nil {
				t.Errorf("NewGenerator returned error (seed %d): %v", seed, err)
				return
			}
			_ = g.Close()
		})
	}
}

// TestNewGenerator_ConcurrentSharedSlice_NoPanic is the primary race-condition
// reproducer for fbxj-vtfy-yrbl-mlyw.
//
// It creates a fields slice with excess backing-array capacity (simulating the
// output of normaliseFields when wildcard fields are filtered out) and calls
// NewGenerator from many goroutines concurrently, all sharing the same slice.
//
// Before the fix, append(fields, objectKeysField...) in
// newGeneratorWithCustomTemplate would write into the shared backing array,
// causing goroutines to overwrite each other's appended entries. The goroutine
// that "won" the write would bind the wrong field name into fieldMap while the
// template still referenced the original name, leading to a nil fieldMap entry
// and a type-assertion panic.
//
// Run with -race to additionally detect the data race itself.
func TestNewGenerator_ConcurrentSharedSlice_NoPanic(t *testing.T) {
	cfg := Config{}

	// Create a fields slice with EXCESS CAPACITY, just as normaliseFields
	// produces when wildcard fields are removed. This is the precondition for
	// the data race: all goroutines share the same backing array and the in-
	// bounds append slot.
	flds := make(Fields, 0, len(auditdPathsFields)+10)
	flds = append(flds, auditdPathsFields...)
	// Invariant: len(flds) < cap(flds); the suffix is the race-prone zone.

	const goroutines = 32
	const callsPerGoroutine = 200

	var wg sync.WaitGroup
	failures := make(chan string, goroutines*callsPerGoroutine)

	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < callsPerGoroutine; j++ {
				seed := int64(id*callsPerGoroutine + j)
				func() {
					defer func() {
						if r := recover(); r != nil {
							failures <- fmt.Sprintf("goroutine %d seed %d: panic: %v", id, seed, r)
						}
					}()
					gen, err := NewGenerator(cfg, flds, 0, WithRandSeed(seed))
					if err != nil {
						failures <- fmt.Sprintf("goroutine %d seed %d: error: %v", id, seed, err)
						return
					}
					_ = gen.Close()
				}()
			}
		}(g)
	}

	wg.Wait()
	close(failures)

	for msg := range failures {
		t.Errorf("%s", msg)
	}
}
