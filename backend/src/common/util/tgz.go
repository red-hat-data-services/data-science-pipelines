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
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"math"
	"strings"
)

// Utils to archive and extract tgz (tar.gz) file.

// SaturatingAdd adds two int64s and returns math.MaxInt64 on overflow.
func SaturatingAdd(a, b int64) int64 {
	if b > 0 && a > math.MaxInt64-b {
		return math.MaxInt64
	}
	return a + b
}

// ArchiveTraversalBudget returns an overflow-safe decompressed-byte budget for
// scanning archive headers and metadata (e.g. PAX/GNU) surrounding an entry.
// The budget is maxFileSize plus a 1 MiB overhead for framing, saturating
// at math.MaxInt64 - 1 to avoid wrapping negative when callers add a sentinel byte.
func ArchiveTraversalBudget(maxFileSize int64) int64 {
	const overhead = 1 << 20 // 1 MiB for archive headers and metadata
	if maxFileSize > math.MaxInt64-overhead-1 {
		return math.MaxInt64 - 1
	}
	return maxFileSize + overhead
}

// ArchiveTgz takes a map of files with name as key and content as value and
// tar and gzip it to a tgz content string. Nested files and directories are
// not supported.
func ArchiveTgz(files map[string]string) (string, error) {
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)
	for name, content := range files {
		hdr := &tar.Header{
			Typeflag: tar.TypeReg,
			Name:     name,
			Size:     int64(len(content)),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			return "", err
		}
		if _, err := tw.Write([]byte(content)); err != nil {
			return "", err
		}
	}
	if err := tw.Close(); err != nil {
		return "", err
	}
	if err := gw.Close(); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// ExtractTgz extracts a list of files from a tgz content. The output is a map
// with file name as key and content as value. Nested files and directories are
// not supported.
func ExtractTgz(tgzContent string) (map[string]string, error) {
	sr := strings.NewReader(tgzContent)
	gr, err := gzip.NewReader(sr)
	if err != nil {
		return nil, err
	}
	tr := tar.NewReader(gr)

	files := make(map[string]string)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if hdr == nil {
			continue
		}
		fileContent, err := io.ReadAll(tr)
		if err != nil {
			return nil, err
		}
		files[hdr.Name] = string(fileContent)
	}
	return files, nil
}
