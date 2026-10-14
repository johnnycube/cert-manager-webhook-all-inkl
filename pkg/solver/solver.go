// SPDX-License-Identifier: MIT
// Derived from cert-manager webhook examples.
// This implements the DNS01 webhook interface: Present / CleanUp at exact FQDN with exact TXT.

// Package solver implements the cert-manager DNS-01 webhook solver for all-inkl.com (KAS API).
package solver

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/cert-manager/cert-manager/pkg/acme/webhook"
	"github.com/cert-manager/cert-manager/pkg/acme/webhook/apis/acme/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// allinklSolver implements the cert-manager DNS01 solver for All-Inkl.
type allinklSolver struct {
	k kubernetes.Interface
	a *AllinklClient
}

// New returns an uninitialized solver; Initialize attaches the Kubernetes and All-Inkl clients.
func New() webhook.Solver { return &allinklSolver{} }

// Name must match .spec.acme.solvers[].dns01.webhook.solverName in the Issuer.
func (s *allinklSolver) Name() string { return "allinkl" }

// Initialize builds the Kubernetes client from kubeClientConfig. It runs once
// at startup, before the solver is used concurrently. The stop channel is part
// of the cert-manager interface and is not needed.
func (s *allinklSolver) Initialize(kubeClientConfig *rest.Config, _ <-chan struct{}) error {
	cl, err := kubernetes.NewForConfig(kubeClientConfig)
	if err != nil {
		return err
	}

	s.k = cl
	s.a = NewAllinklClient()

	return nil
}

// challenge holds what Present and CleanUp need from a ChallengeRequest.
type challenge struct {
	user, pass string
	zone, name string
	key        string
}

// resolve loads the solver config, reads the credentials from Secrets in the
// challenge's namespace and computes the record name relative to the zone.
func (s *allinklSolver) resolve(ch *v1alpha1.ChallengeRequest) (challenge, error) {
	cfg, err := loadConfig(ch.Config)
	if err != nil {
		return challenge{}, err
	}

	username, password, err := s.getCredential(cfg, ch.ResourceNamespace)
	if err != nil {
		return challenge{}, err
	}

	name, err := RelativeName(ch.ResolvedFQDN, ch.ResolvedZone)
	if err != nil {
		return challenge{}, err
	}

	return challenge{
		user: string(username), pass: string(password),
		zone: ch.ResolvedZone, name: name, key: ch.Key,
	}, nil
}

// Present creates the DNS-01 TXT record for the ACME challenge.
func (s *allinklSolver) Present(ch *v1alpha1.ChallengeRequest) error {
	slog.Info("presenting DNS-01 challenge", "fqdn", ch.ResolvedFQDN, "zone", ch.ResolvedZone)

	c, err := s.resolve(ch)
	if err != nil {
		slog.Error("present failed", "error", err)
		return err
	}

	if err := s.a.upsert(c.user, c.pass, c.zone, c.name, c.key); err != nil {
		slog.Error("upsert failed", "error", err)
		return err
	}

	return nil
}

// CleanUp removes the TXT record Present created; records of other challenges stay.
func (s *allinklSolver) CleanUp(ch *v1alpha1.ChallengeRequest) error {
	slog.Info("cleaning up DNS-01 challenge", "fqdn", ch.ResolvedFQDN, "zone", ch.ResolvedZone)

	c, err := s.resolve(ch)
	if err != nil {
		slog.Error("cleanup failed", "error", err)
		return err
	}

	if err := s.a.cleanup(c.user, c.pass, c.zone, c.name, c.key); err != nil {
		slog.Error("cleanup failed", "error", err)
		return err
	}

	return nil
}

// getCredential returns the username and password from the Secrets cfg
// references in namespace ns. Callers must not log them.
func (s *allinklSolver) getCredential(cfg *Config, ns string) ([]byte, []byte, error) {
	username, err := s.getSecretData(cfg.UsernameSecretRef, ns)
	if err != nil {
		return nil, nil, err
	}

	password, err := s.getSecretData(cfg.PasswordSecretRef, ns)
	if err != nil {
		return nil, nil, err
	}

	return username, password, nil
}

// getSecretData returns selector.Key of the Secret selector.Name in namespace
// ns. Errors name the Secret and key, never the value.
func (s *allinklSolver) getSecretData(selector corev1.SecretKeySelector, ns string) ([]byte, error) {
	secret, err := s.k.CoreV1().Secrets(ns).Get(context.TODO(), selector.Name, v1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to load secret %q: %w", ns+"/"+selector.Name, err)
	}

	if data, ok := secret.Data[selector.Key]; ok {
		return data, nil
	}

	return nil, fmt.Errorf("no key %q in secret %q", selector.Key, ns+"/"+selector.Name)
}

// normDomain lower-cases s and trims leading and trailing dots.
func normDomain(s string) string {
	s = strings.ToLower(s)
	s = strings.TrimPrefix(s, ".")
	s = strings.TrimSuffix(s, ".")
	return s
}

// RelativeName returns the name relative to zone.
// Examples:
//
//	fqdn="_acme-challenge.subdomain.example.com.", zone=".example.com."   -> "_acme-challenge.subdomain"
//	fqdn="_acme-challenge.example.com",            zone="example.com"     -> "_acme-challenge"
//	fqdn="example.com.",                           zone="example.com"     -> "" (apex)
func RelativeName(fqdn, zone string) (string, error) {
	f := normDomain(fqdn)
	z := normDomain(zone)

	if f == z {
		return "", nil // apex; KAS uses an empty record name for the zone apex
	}
	if strings.HasSuffix(f, "."+z) {
		rel := strings.TrimSuffix(f, "."+z)
		return rel, nil
	}
	return "", fmt.Errorf("fqdn %q is not within zone %q", fqdn, zone)
}
