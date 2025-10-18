// SPDX-License-Identifier: MIT
// Derived from cert-manager webhook examples.
// This implements the DNS01 webhook interface: Present / CleanUp at exact FQDN with exact TXT.

package solver

import (
	"context"
	"fmt"
	"sync"

	lego "github.com/johnnycube/cert-manager-webhook-all-inkl/pkg/lego/allinkl"
)

type AllinklClient struct {
	recordIDs   map[string]string
	recordIDsMu sync.Mutex
}

func NewAllinklClient() *AllinklClient {
	return &AllinklClient{
		recordIDs: make(map[string]string),
	}
}

func (a *AllinklClient) upsert(user, pass, zone, name, key, fqdn string) error {

	client := lego.NewClient(user)

	ctx := context.Background()

	identifier := lego.NewIdentifier(user, pass)

	credential, err := identifier.Authentication(ctx, 60, true)
	if err != nil {
		return fmt.Errorf("allinkl: %w", err)
	}
	ctx = lego.WithContext(ctx, credential)

	// create update request and send
	record := lego.DNSRequest{
		ZoneHost:   zone,
		RecordType: "TXT",
		RecordName: name,
		RecordData: key,
	}

	recordID, err := client.AddDNSSettings(ctx, record)
	if err != nil {
		return fmt.Errorf("allinkl: %w", err)
	}

	a.recordIDsMu.Lock()
	a.recordIDs[fqdn] = recordID
	a.recordIDsMu.Unlock()

	return nil
}

func (a *AllinklClient) cleanup(user, pass, fqdn string) error {

	client := lego.NewClient(user)

	ctx := context.Background()

	identifier := lego.NewIdentifier(user, pass)

	credential, err := identifier.Authentication(ctx, 60, true)
	if err != nil {
		return fmt.Errorf("allinkl: %w", err)
	}
	ctx = lego.WithContext(ctx, credential)

	// gets the record's unique ID from when we created it
	a.recordIDsMu.Lock()
	recordID, ok := a.recordIDs[fqdn]
	a.recordIDsMu.Unlock()
	if !ok {
		return fmt.Errorf("allinkl: unknown record ID for '%s'", fqdn)
	}

	_, err = client.DeleteDNSSettings(ctx, recordID)
	if err != nil {
		return fmt.Errorf("allinkl: %w", err)
	}

	return nil
}
