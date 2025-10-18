// SPDX-License-Identifier: MIT
// Derived from cert-manager webhook examples.
// This implements the DNS01 webhook interface: Present / CleanUp at exact FQDN with exact TXT.

package solver

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/cert-manager/cert-manager/pkg/acme/webhook/apis/acme/v1alpha1"
	"github.com/pkg/errors"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// allinklSolver implements the cert-manager DNS01 solver for All-Inkl.
type allinklSolver struct {
	k *kubernetes.Clientset
	a *AllinklClient
}

// New returns a new, uninitialized allinklSolver.
//
// The returned solver has no external dependencies set; call Initialize
// to attach the Kubernetes and All-Inkl clients before use. New never
// returns nil.
func New() *allinklSolver { return &allinklSolver{} }

// The name must match .spec.acme.solvers[].dns01.webhook.solverName in the Issuer.
func (s *allinklSolver) Name() string { return "allinkl" }

// Initialize sets up the solver's dependencies.
//
// Initialize builds a Kubernetes client from kubeClientConfig and stores it on
// the receiver. It also constructs a new All-Inkl API client. This method is
// typically called once during webhook startup.
//
// The stopCh is provided to satisfy the cert-manager webhook interface and is
// not used.
//
// On success, c.k and c.a are non-nil. If the Kubernetes client cannot be
// created, Initialize returns the underlying error and leaves c unchanged.
//
// Initialize is not concurrency-safe; call it during process startup before
// using the solver from multiple goroutines.
func (c *allinklSolver) Initialize(kubeClientConfig *rest.Config, stopCh <-chan struct{}) error {
	cl, err := kubernetes.NewForConfig(kubeClientConfig)
	if err != nil {
		return err
	}

	c.k = cl
	c.a = NewAllinklClient()

	return nil
}

// Present creates/updates the DNS-01 TXT record for the ACME challenge.
//
// It loads provider config from ch.Config, resolves credentials from Secrets in
// ch.ResourceNamespace, computes the record name relative to ch.ResolvedZone,
// and asks the All-Inkl client to upsert the TXT record for ch.ResolvedFQDN.
//
// Returns nil on success. If the record already exists with the same value,
// Present is idempotent and returns nil. Requires Initialize to have been called.
func (c *allinklSolver) Present(ch *v1alpha1.ChallengeRequest) error {

	log.Printf("CHALLENGE: %+v", ch)

	cfg, err := loadConfig(ch.Config)
	if err != nil {
		log.Printf("ERR loadConfig: %v\n", err)
		return err
	}

	username, password, err := c.getCredential(cfg, ch.ResourceNamespace)
	if err != nil {
		log.Printf("ERR getCredential: %v\n", err)
		return err
	}

	name, err := RelativeName(ch.ResolvedFQDN, ch.ResolvedZone)
	if err != nil {
		log.Printf("ERR getCredential: %v\n", err)
		return err
	}

	err = c.a.upsert(string(username), string(password), ch.ResolvedZone, name, ch.Key, ch.ResolvedFQDN)
	if err != nil {
		log.Printf("ERR upsert: %v\n", err)
		return fmt.Errorf("allinkl: %w", err)
	}

	return nil
}

// CleanUp removes the DNS-01 TXT record previously created for the challenge.
//
// It reads provider configuration from ch.Config, resolves All-Inkl credentials
// (scoped to ch.ResourceNamespace), and instructs the All-Inkl API client to
// delete the TXT record for ch.ResolvedFQDN.
func (c *allinklSolver) CleanUp(ch *v1alpha1.ChallengeRequest) error {
	cfg, err := loadConfig(ch.Config)
	if err != nil {
		return err
	}

	username, password, err := c.getCredential(cfg, ch.ResourceNamespace)
	if err != nil {
		return err
	}

	err = c.a.cleanup(string(username), string(password), ch.ResolvedFQDN)
	if err != nil {
		return fmt.Errorf("allinkl: %w", err)
	}

	return nil
}

// getCredential returns the All-Inkl username and password from Kubernetes
// Secrets as configured in cfg.
//
// It looks up cfg.UsernameSecretRef and cfg.PasswordSecretRef in the
// namespace ns via getSecretData. On success it returns (username, password)
// as raw bytes. Callers should avoid logging these values and should zero
// them after use if they’re copied to longer-lived buffers.
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

// getSecretData returns the value of selector.Key from the Secret named
// selector.Name in namespace ns.
//
// It fetches the Secret via the Kubernetes CoreV1 client. On success it
// returns the raw bytes from secret.Data[selector.Key]. If the Secret is
// missing, or the key does not exist, it returns a wrapped error. Secrets
// are never logged and their values are not included in errors.
func (s *allinklSolver) getSecretData(selector corev1.SecretKeySelector, ns string) ([]byte, error) {
	secret, err := s.k.CoreV1().Secrets(ns).Get(context.TODO(), selector.Name, v1.GetOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to load secret %q", ns+"/"+selector.Name)
	}

	if data, ok := secret.Data[selector.Key]; ok {
		return data, nil
	}

	return nil, errors.Errorf("no key %q in secret %q", selector.Key, ns+"/"+selector.Name)
}

// normalize: lower-case, trim leading/trailing dots
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
//	fqdn="example.com.",                           zone="example.com"     -> "@"
func RelativeName(fqdn, zone string) (string, error) {
	f := normDomain(fqdn)
	z := normDomain(zone)

	if f == z {
		return "@", nil // apex
	}
	if strings.HasSuffix(f, "."+z) {
		rel := strings.TrimSuffix(f, "."+z)
		return rel, nil
	}
	return "", fmt.Errorf("fqdn %q is not within zone %q", fqdn, zone)
}
