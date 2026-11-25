package trae

import (
	"net/http"
	"testing"
)

// The UG checkin backend keys its per-device daily limit on X-Device-Id.
// Numeric IDs issued by current logins must remain bare; legacy hex IDs use
// the aha-<hex> shape. Adding aha- to a numeric ID is itself rejected with 9074.
func TestSetUgHeadersNormalizesDeviceID(t *testing.T) {
	cases := []struct {
		name     string
		deviceID string
		want     string
	}{
		{"bare hex gets the aha prefix", "3ee0250fe3e0e6e03cdd8054e3a01a7b", "aha-3ee0250fe3e0e6e03cdd8054e3a01a7b"},
		{"prefixed id is left alone", "aha-3ee0250fe3e0e6e03cdd8054e3a01a7b", "aha-3ee0250fe3e0e6e03cdd8054e3a01a7b"},
		{"whitespace is trimmed", "  cfac6e546e2386ec29e13335393d79dc \n", "aha-cfac6e546e2386ec29e13335393d79dc"},
		{"numeric id is sent bare", "2848185088739722", "2848185088739722"},
		{"numeric aha prefix is preserved", "aha-2848185088739722", "aha-2848185088739722"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			header := http.Header{}
			SetUgHeaders(header, Credential{AccessToken: "at", DeviceID: testCase.deviceID})
			if got := header.Get("X-Device-Id"); got != testCase.want {
				t.Fatalf("X-Device-Id=%q want %q", got, testCase.want)
			}
			if got := header.Get("Authorization"); got != "Cloud-IDE-JWT at" {
				t.Fatalf("Authorization=%q", got)
			}
		})
	}

	header := http.Header{}
	SetUgHeaders(header, Credential{AccessToken: "at"})
	if got := header.Get("X-Device-Id"); got != "" {
		t.Fatalf("missing device id must not be sent: %q", got)
	}
}

// The catalog endpoint keys its returned model set on the IDE build identity.
// cli2api v0.7.45 briefly advertised a newer Trae build; a direct curl with the
// same account returned a much smaller catalog (chat_v3 45 vs 65, solo 15 vs 54)
// and omitted newly released models. Assert the identity used by cli2api
// v0.7.44 so the regression cannot return silently.
func TestSetCatalogHeadersUseKnownGoodBuildIdentity(t *testing.T) {
	header := http.Header{}
	SetCatalogHeaders(header, Credential{AccessToken: "at"})

	want := map[string]string{
		"User-Agent":         "Trae/3.3.62",
		"X-Ide-Version":      "3.3.62",
		"X-Ide-Version-Code": "20260811",
		"X-App-Version-Code": "20260811",
	}
	for name, expected := range want {
		if got := header.Get(name); got != expected {
			t.Errorf("%s=%q want %q", name, got, expected)
		}
	}
}
