// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License 2.0;
// you may not use this file except in compliance with the Elastic License 2.0.

package cmd_test

import (
	"bytes"
	"testing"

	"github.com/elastic/elastic-integration-corpus-generator-tool/cmd"
	"github.com/elastic/elastic-integration-corpus-generator-tool/internal/version"
	"github.com/stretchr/testify/require"
)

// saveVersionState saves the current version state and returns a cleanup function
// that restores it. Use with t.Cleanup() to ensure isolation between tests.
func saveVersionState(t *testing.T) {
	t.Helper()
	origTag := version.Tag
	origSourceDateEpoch := version.SourceDateEpoch
	origCommitHash := version.CommitHash

	t.Cleanup(func() {
		version.Tag = origTag
		version.SourceDateEpoch = origSourceDateEpoch
		version.CommitHash = origCommitHash
	})
}

func TestVersionCmd_default(t *testing.T) {
	saveVersionState(t)

	cmd := cmd.VersionCmd()

	b := new(bytes.Buffer)
	cmd.SetOut(b)

	version.Tag = ""
	version.SourceDateEpoch = ""
	version.CommitHash = "undefined"

	err := cmd.Execute()
	require.Nil(t, err)

	const expected = "elastic-integration-corpus-generator-tool devel version-hash undefined (source date: unknown)"
	require.Equal(t, expected, b.String())
}

func TestVersionCmd_withValues(t *testing.T) {
	saveVersionState(t)

	cmd := cmd.VersionCmd()

	b := new(bytes.Buffer)
	cmd.SetOut(b)

	version.Tag = "v0.1.0"
	version.SourceDateEpoch = "1648570012"
	version.CommitHash = "5561aef"

	err := cmd.Execute()
	require.Nil(t, err)

	const expected = "elastic-integration-corpus-generator-tool v0.1.0 version-hash 5561aef (source date: 2022-03-29T16:06:52Z)"
	require.Equal(t, expected, b.String())
}
