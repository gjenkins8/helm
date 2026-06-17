/*
Copyright The Helm Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVersionSet(t *testing.T) {
	vs := VersionSet{"v1", "apps/v1"}
	if d := len(vs); d != 2 {
		assert.Equal(t, 2, d)
	}

	if !vs.Has("apps/v1") {
		assert.Fail(t, "Expected to find apps/v1")
	}

	if vs.Has("Spanish/inquisition") {
		assert.Fail(t, "No one expects the Spanish/inquisition")
	}
}

func TestDefaultVersionSet(t *testing.T) {
	if !DefaultVersionSet.Has("v1") {
		assert.Fail(t, "Expected core v1 version set")
	}
}

func TestDefaultCapabilities(t *testing.T) {
	caps := DefaultCapabilities
	kv := caps.KubeVersion
	if kv.String() != "v1.20.0" {
		assert.Equal(t, "v1.20.0", kv.String())
	}
	if kv.Version != "v1.20.0" {
		assert.Equal(t, "v1.20.0", kv.Version)
	}
	if kv.GitVersion() != "v1.20.0" {
		assert.Equal(t, "v1.20.0", kv.Version)
	}
	if kv.Major != "1" {
		assert.Equal(t, "1", kv.Major)
	}
	if kv.Minor != "20" {
		assert.Equal(t, "20", kv.Minor)
	}

	hv := caps.HelmVersion
	if hv.Version != "v4.2" {
		assert.Equal(t, "v4.2", hv.Version)
	}
}

func TestParseKubeVersion(t *testing.T) {
	kv, err := ParseKubeVersion("v1.16.0")
	if err != nil {
		assert.Fail(t, "Expected v1.16.0 to parse successfully")
	}
	if kv.Version != "v1.16.0" {
		assert.Equal(t, "v1.16.0", kv.String())
	}
	if kv.Major != "1" {
		assert.Equal(t, "1", kv.Major)
	}
	if kv.Minor != "16" {
		assert.Equal(t, "16", kv.Minor)
	}
}

func TestParseKubeVersionWithVendorSuffixes(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantVer    string
		wantString string
		wantMajor  string
		wantMinor  string
	}{
		{"GKE vendor suffix", "v1.33.4-gke.1245000", "v1.33.4-gke.1245000", "v1.33.4", "1", "33"},
		{"GKE without v", "1.30.2-gke.1587003", "v1.30.2-gke.1587003", "v1.30.2", "1", "30"},
		{"EKS trailing +", "v1.28+", "v1.28+", "v1.28", "1", "28"},
		{"EKS + without v", "1.28+", "v1.28+", "v1.28", "1", "28"},
		{"Standard version", "v1.31.0", "v1.31.0", "v1.31.0", "1", "31"},
		{"Standard without v", "1.29.0", "v1.29.0", "v1.29.0", "1", "29"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kv, err := ParseKubeVersion(tt.input)
			if err != nil {
				require.NoError(t, err)
			}
			if kv.Version != tt.wantVer {
				assert.Equal(t, tt.wantVer, kv.Version)
			}
			if kv.String() != tt.wantString {
				assert.Equal(t, tt.wantString, kv.String())
			}
			if kv.Major != tt.wantMajor {
				assert.Equal(t, tt.wantMajor, kv.Major)
			}
			if kv.Minor != tt.wantMinor {
				assert.Equal(t, tt.wantMinor, kv.Minor)
			}
		})
	}
}
