package domain

import "testing"

func TestParseSecretRef_Valid(t *testing.T) {
	for _, ref := range []string{
		"secret://external_service/feishu_link/token",
		"secret://webhook/official_webhook_notify/signing-secret",
		"secret://callback/official_webhook_notify/callback-token",
		"secret://ns/a",
		"secret://ns/a_b-c/def_1-2",
	} {
		ns, name, err := ParseSecretRef(ref)
		if err != nil {
			t.Fatalf("expected %q to be valid, err=%v", ref, err)
		}
		if ns == "" || name == "" {
			t.Fatalf("expected ns/name, got ns=%q name=%q for %q", ns, name, ref)
		}
	}
}

func TestParseSecretRef_Invalid(t *testing.T) {
	for _, ref := range []string{
		"",
		"secret://",
		"secret://ns",
		"secret:///a",
		"secret://NS/a",
		"secret://ns/a/",
		"secret://ns//a",
		"secret://ns/a..b",
		"secret://ns/../a",
		"secret://ns/a\\b",
		"http://ns/a",
		"extsvc_123",
	} {
		if _, _, err := ParseSecretRef(ref); err == nil {
			t.Fatalf("expected %q to be invalid", ref)
		}
	}
}
