// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License 2.0;
// you may not use this file except in compliance with the Elastic License 2.0.

package genlib

import "math/rand"

// options holds the configuration options for generators.
type options struct {
	randSeed int64
	template []byte
	make     func(Config, Fields, uint64, options) (Generator, error)
}

// Option defines a functional option for configuring generators.
type Option func(*options)

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
		randSeed: rand.Int63(),
		make:     newGeneratorWithCustomTemplate,
	}
	for _, opt := range opts {
		opt(&o)
	}
	return o
}
