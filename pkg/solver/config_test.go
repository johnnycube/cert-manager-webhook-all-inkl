// SPDX-License-Identifier: MIT

package solver

import (
	"strings"
	"testing"

	extapi "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		wantErr string
	}{
		{"valid", `{"usernameKeySecretRef":{"name":"kas","key":"user"},"passwordKeySecretRef":{"name":"kas","key":"pass"}}`, ""},
		{"malformed json", `{`, "error decoding solver config"},
		{"username key missing", `{"usernameKeySecretRef":{"name":"kas"},"passwordKeySecretRef":{"name":"kas","key":"pass"}}`, "usernameKeySecretRef"},
		{"password name missing", `{"usernameKeySecretRef":{"name":"kas","key":"user"},"passwordKeySecretRef":{"key":"pass"}}`, "passwordKeySecretRef"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := loadConfig(&extapi.JSON{Raw: []byte(tt.raw)})
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("loadConfig: %v", err)
				}
				if cfg.UsernameSecretRef.Key != "user" || cfg.PasswordSecretRef.Key != "pass" {
					t.Errorf("loadConfig = %+v, want keys user/pass", cfg)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("loadConfig: err = %v, want it to mention %q", err, tt.wantErr)
			}
		})
	}
}

func TestLoadConfigNil(t *testing.T) {
	if _, err := loadConfig(nil); err == nil || !strings.Contains(err.Error(), "invalid solver config") {
		t.Fatalf("loadConfig(nil): err = %v, want invalid solver config", err)
	}
}
