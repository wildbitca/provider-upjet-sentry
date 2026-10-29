package clients

import (
	"strings"
	"testing"
)

const canary = "sntrys_CANARY_SECRET_VALUE"

func TestBuildConfiguration(t *testing.T) {
	cases := map[string]struct {
		creds   map[string]string
		want    map[string]any
		wantErr string
	}{
		"TokenOnlyOmitsBaseURL": {
			creds: map[string]string{"token": canary},
			want:  map[string]any{"token": canary},
		},
		"TokenAndBaseURL": {
			creds: map[string]string{"token": canary, "base_url": "https://de.sentry.io/api/"},
			want:  map[string]any{"token": canary, "base_url": "https://de.sentry.io/api/"},
		},
		"EmptyBaseURLIsSkipped": {
			creds: map[string]string{"token": canary, "base_url": ""},
			want:  map[string]any{"token": canary},
		},
		"EmptyTokenIsAnError": {
			creds:   map[string]string{"token": "", "base_url": "https://sentry.io/api/"},
			wantErr: errMissingToken,
		},
		"MissingTokenIsAnError": {
			creds:   map[string]string{"base_url": "https://sentry.io/api/"},
			wantErr: errMissingToken,
		},
		"BaseURLWithoutAPISuffixIsAnError": {
			creds:   map[string]string{"token": canary, "base_url": "https://sentry.example.com"},
			wantErr: errInvalidBaseURL,
		},
		"PlainHTTPBaseURLIsAnError": {
			creds:   map[string]string{"token": canary, "base_url": "http://sentry.example.com/api/"},
			wantErr: errInvalidBaseURL,
		},
		"SchemeLessBaseURLIsAnError": {
			creds:   map[string]string{"token": canary, "base_url": "sentry.example.com/api/"},
			wantErr: errInvalidBaseURL,
		},
		"BaseURLWithUserinfoIsAnError": {
			creds:   map[string]string{"token": canary, "base_url": "https://user:pw@sentry.example.com/api/"},
			wantErr: errInvalidBaseURL,
		},
		"BaseURLWithQueryIsAnError": {
			creds:   map[string]string{"token": canary, "base_url": "https://sentry.example.com/api/?x=1"},
			wantErr: errInvalidBaseURL,
		},
		"BaseURLWithFragmentIsAnError": {
			creds:   map[string]string{"token": canary, "base_url": "https://sentry.example.com/api/#frag"},
			wantErr: errInvalidBaseURL,
		},
		"BaseURLWithoutHostIsAnError": {
			creds:   map[string]string{"token": canary, "base_url": "https:///api/"},
			wantErr: errInvalidBaseURL,
		},
		"BaseURLWithDotSegmentsIsAnError": {
			creds:   map[string]string{"token": canary, "base_url": "https://sentry.example.com/x/../api/"},
			wantErr: errInvalidBaseURL,
		},
		"BaseURLWithPortAndPathPrefixIsValid": {
			creds: map[string]string{"token": canary, "base_url": "https://sentry.example.com:8443/self/api/"},
			want:  map[string]any{"token": canary, "base_url": "https://sentry.example.com:8443/self/api/"},
		},
		"WhitespaceOnlyTokenIsAnError": {
			creds:   map[string]string{"token": "  \n"},
			wantErr: errMissingToken,
		},
		"ValuesAreTrimmed": {
			creds: map[string]string{"token": " " + canary + "\n", "base_url": " https://sentry.io/api/ "},
			want:  map[string]any{"token": canary, "base_url": "https://sentry.io/api/"},
		},
		"WhitespaceOnlyBaseURLIsSkipped": {
			creds: map[string]string{"token": canary, "base_url": "   "},
			want:  map[string]any{"token": canary},
		},
		"UnknownKeysAreIgnored": {
			creds: map[string]string{"token": canary, "other": "x"},
			want:  map[string]any{"token": canary},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := buildConfiguration(tc.creds)
			if tc.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tc.wantErr)
				}
				if !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("expected error containing %q, got %q", tc.wantErr, err.Error())
				}
				if strings.Contains(err.Error(), canary) {
					t.Fatalf("error leaks the token: %q", err.Error())
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(got) != len(tc.want) {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
			for k, v := range tc.want {
				if got[k] != v {
					t.Fatalf("key %q: got %v, want %v", k, got[k], v)
				}
			}
		})
	}
}

func TestParseCredentials(t *testing.T) {
	cases := map[string]struct {
		data    string
		wantErr bool
	}{
		"Valid":              {data: `{"token":"` + canary + `"}`},
		"InvalidJSON":        {data: `{"token":"` + canary, wantErr: true},
		"NonStringValue":     {data: `{"token":"` + canary + `","base_url":123}`, wantErr: true},
		"NotAnObject":        {data: `"` + canary + `"`, wantErr: true},
		"TokenWrongTypeLeak": {data: `{"token":["` + canary + `"]}`, wantErr: true},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := parseCredentials([]byte(tc.data))
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected an error")
				}
				if strings.Contains(err.Error(), canary) {
					t.Fatalf("error leaks the token: %q", err.Error())
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got["token"] != canary {
				t.Fatalf("token not decoded: %v", got)
			}
		})
	}
}
