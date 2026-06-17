// SPDX-License-Identifier: MIT
// Derived from cert-manager webhook examples.
// This implements the DNS01 webhook interface: Present / CleanUp at exact FQDN with exact TXT.

package solver

import (
	"encoding/json"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	extapi "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

type Config struct {
	UsernameSecretRef corev1.SecretKeySelector `json:"usernameKeySecretRef"`
	PasswordSecretRef corev1.SecretKeySelector `json:"passwordKeySecretRef"`
}

// validate ensures both credential secret references are fully specified.
// Without this the solver would request a Secret named "" and fail later
// with a confusing error.
func (c *Config) validate() error {
	if c.UsernameSecretRef.Name == "" || c.UsernameSecretRef.Key == "" {
		return fmt.Errorf("usernameKeySecretRef: both name and key must be set")
	}
	if c.PasswordSecretRef.Name == "" || c.PasswordSecretRef.Key == "" {
		return fmt.Errorf("passwordKeySecretRef: both name and key must be set")
	}
	return nil
}

// loadConfig is a small helper function that decodes JSON configuration into
// the typed config struct and validates it.
func loadConfig(cfgJSON *extapi.JSON) (*Config, error) {
	cfg := Config{}

	if cfgJSON != nil {
		if err := json.Unmarshal(cfgJSON.Raw, &cfg); err != nil {
			return &cfg, fmt.Errorf("error decoding solver config: %w", err)
		}
	}

	if err := cfg.validate(); err != nil {
		return &cfg, fmt.Errorf("invalid solver config: %w", err)
	}

	return &cfg, nil
}
