// SPDX-License-Identifier: MIT
// This implements the DNS01 webhook interface: Present / CleanUp at exact FQDN with exact TXT.

package solver

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/johnnycube/kasapi"
)

type AllinklClient struct{}

func NewAllinklClient() *AllinklClient {
	return &AllinklClient{}
}

// Endpoint overrides for tests; empty means the kasapi defaults.
var (
	kasAPIEndpoint  string
	kasAuthEndpoint string
)

// newKasClient builds a kasapi client for the given credentials. Session
// handling and KAS flood-protection delays are managed inside the client.
func newKasClient(user, pass string) (*kasapi.Client, error) {
	return kasapi.New(kasapi.Config{
		Login:    user,
		Password: pass,
		// KAS accounts commonly have sha1 session auth disabled
		// (kas_auth_type_disabled); plain matches the previous behavior.
		AuthType:     kasapi.AuthPlain,
		APIEndpoint:  kasAPIEndpoint,
		AuthEndpoint: kasAuthEndpoint,
	})
}

func (a *AllinklClient) upsert(user, pass, zone, name, key string) error {
	client, err := newKasClient(user, pass)
	if err != nil {
		return fmt.Errorf("allinkl: %w", err)
	}

	slog.Info("creating DNS TXT record", "zone", zone, "name", name)

	_, err = client.DNS.Create(context.Background(), kasapi.DNSRecord{
		Zone: zone,
		Name: name,
		Type: "TXT",
		Data: key,
	})
	if err != nil {
		return fmt.Errorf("allinkl: %w", err)
	}

	return nil
}

func (a *AllinklClient) cleanup(user, pass, zone, name, key string) error {
	client, err := newKasClient(user, pass)
	if err != nil {
		return fmt.Errorf("allinkl: %w", err)
	}

	ctx := context.Background()

	records, err := client.DNS.List(ctx, zone)
	if err != nil {
		return fmt.Errorf("allinkl: get dns settings: %w", err)
	}

	for _, r := range records {
		if !strings.EqualFold(r.Type, "TXT") {
			continue
		}

		// Match on name AND value so we only delete the record this
		// challenge created, never a concurrent challenge's record.
		if r.Name == name && r.Data == key {
			slog.Info("deleting DNS TXT record", "zone", zone, "name", name)
			if err := client.DNS.Delete(ctx, r.ID); err != nil {
				return fmt.Errorf("allinkl: %w", err)
			}
		}
	}

	return nil
}
