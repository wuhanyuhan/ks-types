package kstypes

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestSecurityScheme_APIKeyJSON(t *testing.T) {
	t.Parallel()
	s := APIKeyScheme{In: "header", Name: "X-API-Key"}
	data, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got, want map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal got: %v", err)
	}
	if err := json.Unmarshal([]byte(`{"type":"apiKey","in":"header","name":"X-API-Key"}`), &want); err != nil {
		t.Fatalf("unmarshal want: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("APIKey marshal mismatch:\n got  = %s\n want = %#v", string(data), want)
	}
}

func TestAPIKeyScheme_SchemeType(t *testing.T) {
	t.Parallel()
	s := APIKeyScheme{}
	if got := s.SchemeType(); got != "apiKey" {
		t.Errorf("SchemeType: got %q, want apiKey", got)
	}
}

func TestSecurityScheme_HTTPAuthJSON(t *testing.T) {
	t.Parallel()
	s := HTTPAuthScheme{Scheme: "bearer", BearerFormat: "JWT"}
	data, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got, want map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal got: %v", err)
	}
	if err := json.Unmarshal([]byte(`{"type":"http","scheme":"bearer","bearerFormat":"JWT"}`), &want); err != nil {
		t.Fatalf("unmarshal want: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("HTTPAuth marshal mismatch:\n got  = %s\n want = %#v", string(data), want)
	}
}

func TestSecurityScheme_HTTPAuthJSON_NoBearerFormat(t *testing.T) {
	t.Parallel()
	s := HTTPAuthScheme{Scheme: "basic"}
	data, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal got: %v", err)
	}
	if _, ok := got["bearerFormat"]; ok {
		t.Errorf("expected bearerFormat to be omitted when empty, got %v", got)
	}
}

func TestHTTPAuthScheme_SchemeType(t *testing.T) {
	t.Parallel()
	s := HTTPAuthScheme{}
	if got := s.SchemeType(); got != "http" {
		t.Errorf("SchemeType: got %q, want http", got)
	}
}

func TestSecurityScheme_OAuth2JSON(t *testing.T) {
	t.Parallel()
	s := OAuth2Scheme{
		Flows: map[string]OAuth2Flow{
			"authorizationCode": {
				AuthorizationURL: "https://example.com/oauth/authorize",
				TokenURL:         "https://example.com/oauth/token",
				Scopes: map[string]string{
					"read:agents":  "Read agent definitions",
					"write:agents": "Create and update agents",
				},
			},
		},
	}
	data, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal got: %v", err)
	}
	if got["type"] != "oauth2" {
		t.Errorf("type field: got %q, want oauth2", got["type"])
	}
	if _, ok := got["flows"]; !ok {
		t.Errorf("expected flows field in marshaled output, got %v", got)
	}
}

func TestOAuth2Scheme_SchemeType(t *testing.T) {
	t.Parallel()
	s := OAuth2Scheme{}
	if got := s.SchemeType(); got != "oauth2" {
		t.Errorf("SchemeType: got %q, want oauth2", got)
	}
}

func TestSecurityScheme_OpenIDConnectJSON(t *testing.T) {
	t.Parallel()
	s := OpenIDConnectScheme{
		OpenIDConnectURL: "https://example.com/.well-known/openid-configuration",
	}
	data, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got, want map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal got: %v", err)
	}
	if err := json.Unmarshal([]byte(`{"type":"openIdConnect","openIdConnectUrl":"https://example.com/.well-known/openid-configuration"}`), &want); err != nil {
		t.Fatalf("unmarshal want: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("OpenIDConnect marshal mismatch:\n got  = %s\n want = %#v", string(data), want)
	}
}

func TestOpenIDConnectScheme_SchemeType(t *testing.T) {
	t.Parallel()
	s := OpenIDConnectScheme{}
	if got := s.SchemeType(); got != "openIdConnect" {
		t.Errorf("SchemeType: got %q, want openIdConnect", got)
	}
}

func TestSecurityScheme_MutualTLSJSON(t *testing.T) {
	t.Parallel()
	s := MutualTLSScheme{}
	data, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got, want map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal got: %v", err)
	}
	if err := json.Unmarshal([]byte(`{"type":"mutualTLS"}`), &want); err != nil {
		t.Fatalf("unmarshal want: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("MutualTLS marshal mismatch:\n got  = %s\n want = %#v", string(data), want)
	}
}

func TestMutualTLSScheme_SchemeType(t *testing.T) {
	t.Parallel()
	s := MutualTLSScheme{}
	if got := s.SchemeType(); got != "mutualTLS" {
		t.Errorf("SchemeType: got %q, want mutualTLS", got)
	}
}
