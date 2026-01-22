// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License 2.0;
// you may not use this file except in compliance with the Elastic License 2.0.

package genlib

import (
	"math/rand"
	"time"
)

// options holds the configuration options for generators.
type options struct {
	randSeed  int64
	startTime time.Time
	template  []byte
	make      func(Config, Fields, uint64, options) (Generator, error)
}

// Option defines a functional option for configuring generators.
type Option func(*options)

// WithStartTime sets the start time for the generator.
func WithStartTime(t time.Time) Option {
	return func(o *options) {
		o.startTime = t
	}
}

// WithRandSeed sets the random seed for the generator.
func WithRandSeed(seed int64) Option {
	return func(o *options) {
		o.randSeed = seed
	}
}

// WithTextTemplate sets a Go text template for the generator.
func WithTextTemplate(template []byte) Option {
	return func(o *options) {
		o.template = template
		o.make = newGeneratorWithTextTemplate
	}
}

// WithCustomTemplate sets a custom placeholder template for the generator.
func WithCustomTemplate(template []byte) Option {
	return func(o *options) {
		o.template = template
		o.make = newGeneratorWithCustomTemplate
	}
}

// applyOptions applies the given options and returns the final configuration.
func applyOptions(opts []Option) options {
	o := options{
		make:      newGeneratorWithCustomTemplate,
		randSeed:  rand.Int63(),
		startTime: time.Now(),
	}
	for _, opt := range opts {
		opt(&o)
	}
	return o
}
