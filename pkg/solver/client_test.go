// SPDX-License-Identifier: MIT

package solver

import (
	"fmt"
	"strings"
	"testing"

	"github.com/johnnycube/kasapi/kasapitest"
)

// dnsRecordXML renders one zone record in the apache-Map notation the KAS API
// returns from get_dns_settings.
func dnsRecordXML(id, name, typ, data string) string {
	return "<item>" +
		kasapitest.MapItem("record_id", id) +
		kasapitest.MapItem("record_name", name) +
		kasapitest.MapItem("record_type", typ) +
		kasapitest.MapItem("record_data", data) +
		kasapitest.MapItem("record_aux", "0") +
		kasapitest.MapItem("record_changeable", "Y") +
		"</item>"
}

func withFakeKas(t *testing.T, handler kasapitest.Handler) *kasapitest.Server {
	t.Helper()
	srv := kasapitest.New(t, handler)
	kasAPIEndpoint = srv.APIURL()
	kasAuthEndpoint = srv.AuthURL()
	t.Cleanup(func() {
		kasAPIEndpoint = ""
		kasAuthEndpoint = ""
	})
	return srv
}

func TestUpsert(t *testing.T) {
	var gotParams map[string]any
	withFakeKas(t, func(action string, params map[string]any) (string, string) {
		switch action {
		case "add_dns_settings":
			gotParams = params
			return `<value xsi:type="xsd:string">12345</value>`, ""
		default:
			return "", "unexpected_action_" + action
		}
	})

	a := NewAllinklClient()
	err := a.upsert("w0123456", "secret", "example.com.", "_acme-challenge", "token-value")
	if err != nil {
		t.Fatalf("upsert: %v", err)
	}

	want := map[string]string{
		"zone_host":   "example.com.",
		"record_type": "TXT",
		"record_name": "_acme-challenge",
		"record_data": "token-value",
	}
	for k, v := range want {
		if got := fmt.Sprint(gotParams[k]); got != v {
			t.Errorf("add_dns_settings param %s = %q, want %q", k, got, v)
		}
	}
}

func TestCleanup(t *testing.T) {
	var deleted []string
	withFakeKas(t, func(action string, params map[string]any) (string, string) {
		switch action {
		case "get_dns_settings":
			return dnsRecordXML("1", "_acme-challenge", "TXT", "token-value") +
				dnsRecordXML("2", "_acme-challenge", "TXT", "other-challenge") +
				dnsRecordXML("3", "www", "A", "192.0.2.1"), ""
		case "delete_dns_settings":
			deleted = append(deleted, fmt.Sprint(params["record_id"]))
			return `<value xsi:type="xsd:string">TRUE</value>`, ""
		default:
			return "", "unexpected_action_" + action
		}
	})

	a := NewAllinklClient()
	err := a.cleanup("w0123456", "secret", "example.com.", "_acme-challenge", "token-value")
	if err != nil {
		t.Fatalf("cleanup: %v", err)
	}

	if strings.Join(deleted, ",") != "1" {
		t.Errorf("deleted record ids = %v, want [1] (only the matching name+value)", deleted)
	}
}

func TestCleanupBadCredentials(t *testing.T) {
	withFakeKas(t, func(action string, params map[string]any) (string, string) {
		return "", "unexpected_action_" + action
	})

	a := NewAllinklClient()
	err := a.cleanup("w0123456", "wrong-password", "example.com.", "_acme-challenge", "token-value")
	if err == nil || !strings.Contains(err.Error(), "kas_login_incorrect") {
		t.Fatalf("cleanup with bad credentials: err = %v, want kas_login_incorrect", err)
	}
}
