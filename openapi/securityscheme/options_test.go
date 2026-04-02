package securityscheme

import (
	"testing"

	"github.com/eriicafes/httpx/openapi/store"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

func TestNewAndFlowOptions(t *testing.T) {
	st := store.New(&v3.Components{})

	ref := New(st, "bearerAuth",
		Description("JWT auth"),
		HTTP("bearer"),
		BearerFormat("JWT"),
		Flows(
			FlowAuthorizationCode("https://example.com/auth", "https://example.com/token",
				FlowScope("users:read", "Read users"),
				FlowRefreshUrl("https://example.com/refresh"),
			),
		),
	)
	if ref.Reference != "#/components/securitySchemes/bearerAuth" {
		t.Fatalf("New() ref = %q, want component ref", ref.Reference)
	}
	stored, ok := st.GetSecurityScheme("bearerAuth")
	if !ok || stored.Reference != ref.Reference {
		t.Fatalf("New() did not store security scheme component: %#v ok=%v", stored, ok)
	}

	var ss v3.SecurityScheme
	Reference("ApiKeyRef")(&ss)
	Type(TypeApiKey)(&ss)
	Name("X-API-Key")(&ss)
	In(InHeader)(&ss)
	Scheme("bearer")(&ss)
	OpenIdConnectUrl("https://example.com/openid")(&ss)
	OAuth2MetadataUrl("https://example.com/oauth2")(&ss)
	Deprecated()(&ss)
	Flows(
		FlowImplicit("https://example.com/auth"),
		FlowPassword("https://example.com/token"),
		FlowClientCredentials("https://example.com/client-token"),
		FlowDevice("https://example.com/device-token"),
	)(&ss)
	if ss.Reference != "ApiKeyRef" || ss.Type != TypeApiKey || ss.Name != "X-API-Key" || ss.In != InHeader {
		t.Fatalf("security scheme basic options incorrect: %#v", ss)
	}
	if ss.Flows == nil || ss.Flows.Implicit == nil || ss.Flows.Password == nil || ss.Flows.ClientCredentials == nil || ss.Flows.Device == nil {
		t.Fatalf("Flows() missing flow configuration: %#v", ss.Flows)
	}
	if ss.Deprecated != true {
		t.Fatalf("Deprecated() not applied: %#v", ss.Deprecated)
	}
}
