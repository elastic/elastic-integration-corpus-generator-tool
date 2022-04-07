// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License 2.0;
// you may not use this file except in compliance with the Elastic License 2.0.

package corpus

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilename(t *testing.T) {
	fc := TestNewGenerator()

	expected := "1647345675-integration-data_stream-0.0.1.ndjson"
	got := fc.bulkPayloadFilename("integration", "data_stream", "0.0.1")
	assert.Equal(t, expected, got)
}

func TestSanitizeFilename(t *testing.T) {
	type test struct {
		input string
		want  string
	}

	tests := []test{
		{input: "foo bar", want: "foo-bar"},
		{input: "foo bar foobar", want: "foo-bar-foobar"},
		{input: "foo/bar", want: "foo-bar"},
		{input: "foo\\bar", want: "foo-bar"},
		{input: "foo bar/foobar\\", want: "foo-bar-foobar-"},
	}

	for _, tc := range tests {
		got := sanitizeFilename(tc.input)
		if !reflect.DeepEqual(tc.want, got) {
			t.Fatalf("expected: %v, got: %v", tc.want, got)
		}
	}
}
