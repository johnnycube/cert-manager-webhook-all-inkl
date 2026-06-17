// SPDX-License-Identifier: MIT
// Derived from cert-manager webhook examples.
// This implements the DNS01 webhook interface: Present / CleanUp at exact FQDN with exact TXT.

package solver

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	lego "github.com/johnnycube/cert-manager-webhook-all-inkl/pkg/lego/allinkl"
)

type AllinklClient struct{}

func NewAllinklClient() *AllinklClient {
	return &AllinklClient{}
}

func (a *AllinklClient) upsert(user, pass, zone, name, key string) error {

	client := lego.NewClient(user)

	ctx := context.Background()

	identifier := lego.NewIdentifier(user, pass)

	credential, err := identifier.Authentication(ctx, 60, true)
	if err != nil {
		return fmt.Errorf("allinkl: %w", err)
	}
	ctx = lego.WithContext(ctx, credential)

	slog.Info("creating DNS TXT record", "zone", zone, "name", name)
	time.Sleep(1 * time.Second)

	// create update request and send
	record := lego.DNSRequest{
		ZoneHost:   zone,
		RecordType: "TXT",
		RecordName: name,
		RecordData: key,
	}

	_, err = client.AddDNSSettings(ctx, record)
	if err != nil {
		return fmt.Errorf("allinkl: %w", err)
	}

	return nil
}

func (a *AllinklClient) cleanup(user, pass, zone, name, key string) error {

	client := lego.NewClient(user)

	ctx := context.Background()

	identifier := lego.NewIdentifier(user, pass)

	credential, err := identifier.Authentication(ctx, 60, true)
	if err != nil {
		return fmt.Errorf("allinkl: %w", err)
	}
	ctx = lego.WithContext(ctx, credential)

	info, err := client.GetDNSSettings(ctx, zone, "")
	if err != nil {
		return fmt.Errorf("allinkl: get dns settings: %w", err)
	}

	for _, i := range info {
		if strings.ToUpper(i.Type) != "TXT" {
			continue
		}

		// Match on name AND value so we only delete the record this
		// challenge created, never a concurrent challenge's record.
		if i.Name == name && i.Data == key {
			// This is ugly but needed to prevent KAS flood protection
			slog.Info("deleting DNS TXT record", "zone", zone, "name", name)
			time.Sleep(1 * time.Second)
			idStr, ok := i.ID.(string)
			if !ok {
				return fmt.Errorf("allinkl: record ID is not a string for '%s'", name)
			}
			_, err = client.DeleteDNSSettings(ctx, idStr)
			if err != nil {
				return fmt.Errorf("allinkl: %w", err)
			}
		}
	}

	return nil
}
