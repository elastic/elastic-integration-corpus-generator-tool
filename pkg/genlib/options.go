// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License 2.0;
// you may not use this file except in compliance with the Elastic License 2.0.

package genlib

import "time"

// options holds the configuration options for generators.
type options struct {
	randSeed  int64
	startTime time.Time
	timeSpeed float64
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

// WithTimeSpeed controls how fast simulated time advances relative to wall
// clock time elapsed since the generator's start time.
//
//	0   (default) → time accumulates randomly per Emit call (legacy behaviour)
//	1.0           → simulated clock matches wall clock
//	2.0           → twice real time
//	0.5           → half speed (e.g. historical replay)
//
// Negative values are invalid and will cause NewGenerator to return an error.
//
// Note: WithTimeSpeed only takes effect when no Period is configured on the
// timestamp field. When Period is set and totEvents > 0, the period-based
// distribution takes precedence regardless of this setting.
func WithTimeSpeed(factor float64) Option {
	return func(o *options) {
		o.timeSpeed = factor
	}
}

// WithRandSeed sets the random seed for the generator.
func WithRandSeed(seed int64) Option {
	return func(o *options) {
		o.randSeed = seed
	}
}

// WithTextTemplate sets a Go text template for the generator.
// The template is compiled on each NewGenerator call. To compile once and share
// across multiple generators, use NewTextTemplate instead.
func WithTextTemplate(template []byte) Option {
	return func(o *options) {
		o.make = newGeneratorWithTextTemplate(template)
	}
}

// WithCustomTemplate sets a custom placeholder template for the generator.
// The template is compiled on each NewGenerator call. To compile once and share
// across multiple generators, use NewCustomTemplate instead.
func WithCustomTemplate(template []byte) Option {
	return func(o *options) {
		o.make = newGeneratorWithCustomTemplate(template)
	}
}

// applyOptions applies the given options and returns the final configuration.
func applyOptions(opts []Option) options {
	// This initialization is executed in a concurrent context, any access
	// to non thread-safe resources must be properly synchronized.
	o := options{
		make:      newGeneratorWithCustomTemplate(nil),
		randSeed:  time.Now().UnixNano(),
		startTime: time.Now(),
	}
	for _, opt := range opts {
		opt(&o)
	}
	return o
}
