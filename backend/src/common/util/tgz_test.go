// Copyright 2018 The Kubeflow Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package util

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSaturatingAdd(t *testing.T) {
	tests := []struct {
		name     string
		first    int64
		second   int64
		expected int64
	}{
		{"small values", 2, 3, 5},
		{"adding zero", math.MaxInt64, 0, math.MaxInt64},
		{"one below the limit", math.MaxInt64 - 1, 1, math.MaxInt64},
		{"overflow saturates", math.MaxInt64, 1, math.MaxInt64},
		{"large overflow saturates", math.MaxInt64 - 5, 100, math.MaxInt64},
		{"negative addend is not saturated", 10, -3, 7},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, SaturatingAdd(test.first, test.second))
		})
	}
}

func TestArchiveTraversalBudget(t *testing.T) {
	const overhead = int64(1 << 20)

	assert.Equal(t, overhead, ArchiveTraversalBudget(0))
	assert.Equal(t, int64(32<<20)+overhead, ArchiveTraversalBudget(32<<20))

	// Near the int64 limit the budget saturates one below MaxInt64, so callers
	// can still add a sentinel byte without wrapping negative.
	for _, size := range []int64{math.MaxInt64 - overhead, math.MaxInt64 - 1, math.MaxInt64} {
		budget := ArchiveTraversalBudget(size)
		assert.Equal(t, int64(math.MaxInt64-1), budget)
		assert.Equal(t, int64(math.MaxInt64), SaturatingAdd(budget, 1))
	}
}

func TestArchiveTgzAndExtractTgz_Roundtrip(t *testing.T) {
	expectedFiles := map[string]string{
		"README":         "Hello, World",
		"hello_world.sh": "echo 'Hello, World'",
	}

	tgzContent, err := ArchiveTgz(expectedFiles)
	assert.Nil(t, err)
	assert.NotEmpty(t, tgzContent)

	actualFiles, err := ExtractTgz(tgzContent)
	assert.Nil(t, err)
	assert.Equal(t, expectedFiles, actualFiles)
}

func TestArchiveTgz_EmptyFiles(t *testing.T) {
	tgzContent, err := ArchiveTgz(map[string]string{})
	assert.Nil(t, err)
	assert.NotEmpty(t, tgzContent)

	files, err := ExtractTgz(tgzContent)
	assert.Nil(t, err)
	assert.Empty(t, files)
}

func TestExtractTgz_InvalidContent(t *testing.T) {
	files, err := ExtractTgz("not a valid tgz")
	assert.NotNil(t, err)
	assert.Nil(t, files)
}

func TestArchiveTgz_SingleFile(t *testing.T) {
	tgzContent, err := ArchiveTgz(map[string]string{"file.txt": "content"})
	assert.Nil(t, err)

	files, err := ExtractTgz(tgzContent)
	assert.Nil(t, err)
	assert.Equal(t, "content", files["file.txt"])
}
