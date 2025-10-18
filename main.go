// SPDX-License-Identifier: MIT
// Derived from cert-manager webhook examples.
// This implements the DNS01 webhook interface: Present / CleanUp at exact FQDN with exact TXT.

package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/cert-manager/cert-manager/pkg/acme/webhook/cmd"
	"github.com/johnnycube/cert-manager-webhook-all-inkl/pkg/solver"
)

func main() {
	group := strings.TrimSpace(os.Getenv("GROUP_NAME"))
	if group == "" {
		group = "acme.johanneskueber.com"
	}
	if err := run(group); err != nil {
		fmt.Fprintf(os.Stderr, "webhook exited: %v\n", err)
		os.Exit(1)
	}
}

func run(group string) error {
	if strings.TrimSpace(group) == "" {
		return errors.New("GROUP_NAME must be set/non-empty")
	}
	// Register and run the webhook server on :443 (cert-manager library handles TLS)
	cmd.RunWebhookServer(group, solver.New())
	return nil
}
