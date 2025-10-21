// SPDX-License-Identifier: MIT
// Derived from cert-manager webhook examples.
// This implements the DNS01 webhook interface: Present / CleanUp at exact FQDN with exact TXT.

package solver

import (
	"context"
	"fmt"
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

	fmt.Println("create entry for zone: " + zone + " name: " + name + " key: " + key)
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

func (a *AllinklClient) cleanup(user, pass, zone, name string) error {

	client := lego.NewClient(user)

	ctx := context.Background()

	identifier := lego.NewIdentifier(user, pass)

	credential, err := identifier.Authentication(ctx, 60, true)
	if err != nil {
		return fmt.Errorf("allinkl: %w", err)
	}
	ctx = lego.WithContext(ctx, credential)

	info, _ := client.GetDNSSettings(ctx, zone, "")
	fmt.Printf("zone: %s: %+v\n", zone, info)

	for _, i := range info {
		if strings.ToUpper(i.Type) != "TXT" {
			continue
		}

		if i.Name == name {
			// This is ugly but needed to prevent KAS flood protection
			fmt.Println("delete entry for zone: " + zone + " name: " + name)
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
