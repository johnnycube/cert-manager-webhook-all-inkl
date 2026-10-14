// SPDX-License-Identifier: MIT

package solver

import (
	"fmt"
	"strings"
	"testing"

	"github.com/cert-manager/cert-manager/pkg/acme/webhook/apis/acme/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	extapi "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
)

const solverConfig = `{"usernameKeySecretRef":{"name":"kas","key":"user"},"passwordKeySecretRef":{"name":"kas","key":"pass"}}`

// newTestSolver returns a solver whose Kubernetes client holds the given Secrets.
func newTestSolver(secrets ...*corev1.Secret) *allinklSolver {
	cs := fake.NewClientset()
	for _, s := range secrets {
		_ = cs.Tracker().Add(s)
	}
	return &allinklSolver{k: cs, a: NewAllinklClient()}
}

func kasSecret(data map[string]string) *corev1.Secret {
	s := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "kas", Namespace: "cert-manager"}, Data: map[string][]byte{}}
	for k, v := range data {
		s.Data[k] = []byte(v)
	}
	return s
}

func challengeRequest(fqdn string) *v1alpha1.ChallengeRequest {
	return &v1alpha1.ChallengeRequest{
		ResolvedFQDN:      fqdn,
		ResolvedZone:      "example.com.",
		ResourceNamespace: "cert-manager",
		Key:               "token-value",
		Config:            &extapi.JSON{Raw: []byte(solverConfig)},
	}
}

func TestNameAndNew(t *testing.T) {
	if got := New().Name(); got != "allinkl" {
		t.Errorf("Name() = %q, want allinkl", got)
	}
}

func TestInitialize(t *testing.T) {
	s := &allinklSolver{}
	if err := s.Initialize(&rest.Config{Host: "https://127.0.0.1:6443"}, nil); err != nil {
		t.Fatalf("Initialize: %v", err)
	}
	if s.k == nil || s.a == nil {
		t.Fatal("Initialize left a client nil")
	}

	// client-go rejects a QPS without a burst.
	if err := (&allinklSolver{}).Initialize(&rest.Config{Host: "https://127.0.0.1:6443", QPS: 1}, nil); err == nil {
		t.Fatal("Initialize with QPS and no burst: want error")
	}
}

func TestPresent(t *testing.T) {
	var got map[string]any
	withFakeKas(t, func(action string, params map[string]any) (string, string) {
		if action != "add_dns_settings" {
			return "", "unexpected_action_" + action
		}
		got = params
		return `<value xsi:type="xsd:string">1</value>`, ""
	})

	s := newTestSolver(kasSecret(map[string]string{"user": "w0123456", "pass": "secret"}))
	if err := s.Present(challengeRequest("_acme-challenge.www.example.com.")); err != nil {
		t.Fatalf("Present: %v", err)
	}
	if fmt.Sprint(got["record_name"]) != "_acme-challenge.www" || fmt.Sprint(got["record_data"]) != "token-value" {
		t.Errorf("add_dns_settings params = %v, want name _acme-challenge.www, data token-value", got)
	}
}

func TestCleanUp(t *testing.T) {
	var deleted []string
	withFakeKas(t, func(action string, params map[string]any) (string, string) {
		switch action {
		case "get_dns_settings":
			return dnsRecordXML("7", "_acme-challenge", "TXT", "token-value"), ""
		case "delete_dns_settings":
			deleted = append(deleted, fmt.Sprint(params["record_id"]))
			return `<value xsi:type="xsd:string">TRUE</value>`, ""
		default:
			return "", "unexpected_action_" + action
		}
	})

	s := newTestSolver(kasSecret(map[string]string{"user": "w0123456", "pass": "secret"}))
	if err := s.CleanUp(challengeRequest("_acme-challenge.example.com.")); err != nil {
		t.Fatalf("CleanUp: %v", err)
	}
	if strings.Join(deleted, ",") != "7" {
		t.Errorf("deleted = %v, want [7]", deleted)
	}
}

func TestPresentAndCleanUpErrors(t *testing.T) {
	withFakeKas(t, func(action string, _ map[string]any) (string, string) {
		return "", "unexpected_action_" + action
	})
	good := kasSecret(map[string]string{"user": "w0123456", "pass": "secret"})

	tests := []struct {
		name    string
		s       *allinklSolver
		ch      *v1alpha1.ChallengeRequest
		wantErr string
	}{
		{"invalid config", newTestSolver(good), &v1alpha1.ChallengeRequest{Config: &extapi.JSON{Raw: []byte(`{}`)}}, "invalid solver config"},
		{"secret missing", newTestSolver(), challengeRequest("_acme-challenge.example.com."), `failed to load secret "cert-manager/kas"`},
		{"username key missing", newTestSolver(kasSecret(map[string]string{"pass": "secret"})), challengeRequest("_acme-challenge.example.com."), `no key "user"`},
		{"password key missing", newTestSolver(kasSecret(map[string]string{"user": "w0123456"})), challengeRequest("_acme-challenge.example.com."), `no key "pass"`},
		{"fqdn outside zone", newTestSolver(good), challengeRequest("_acme-challenge.example.org."), "not within zone"},
		{"kas rejects the call", newTestSolver(good), challengeRequest("_acme-challenge.example.com."), "unexpected_action_"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for op, fn := range map[string]func(*v1alpha1.ChallengeRequest) error{"Present": tt.s.Present, "CleanUp": tt.s.CleanUp} {
				if err := fn(tt.ch); err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("%s: err = %v, want it to contain %q", op, err, tt.wantErr)
				}
			}
		})
	}
}

func TestRelativeName(t *testing.T) {
	tests := []struct{ fqdn, zone, want string }{
		{"_acme-challenge.subdomain.example.com.", ".example.com.", "_acme-challenge.subdomain"},
		{"_acme-challenge.example.com", "example.com", "_acme-challenge"},
		{"_ACME-Challenge.Example.COM.", "example.com.", "_acme-challenge"},
		{"example.com.", "example.com", ""},
	}
	for _, tt := range tests {
		got, err := RelativeName(tt.fqdn, tt.zone)
		if err != nil || got != tt.want {
			t.Errorf("RelativeName(%q, %q) = %q, %v; want %q", tt.fqdn, tt.zone, got, err, tt.want)
		}
	}
	if _, err := RelativeName("example.org.", "example.com."); err == nil {
		t.Error("RelativeName outside the zone: want error")
	}
	if _, err := RelativeName("badexample.com.", "example.com."); err == nil {
		t.Error("RelativeName with a suffix that is not a label boundary: want error")
	}
}
