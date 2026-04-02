package doc

import (
	"net/http"
	"testing"

	"github.com/eriicafes/httpx/openapi/op"
	"github.com/eriicafes/httpx/openapi/pathitem"
	"github.com/eriicafes/httpx/openapi/securityscheme"
	"github.com/eriicafes/httpx/openapi/server"
	"github.com/eriicafes/httpx/openapi/store"
	"github.com/eriicafes/httpx/openapi/tag"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

func TestOptionHelpers(t *testing.T) {
	doc := &v3.Document{Info: &base.Info{}}
	st := store.New(&v3.Components{})

	opts := []Option{
		Version("3.0.3"),
		Summary("summary"),
		Description("description"),
		TermsOfService("https://example.com/tos"),
		Contact("team", "https://example.com", "api@example.com"),
		License("MIT", "https://opensource.org", "MIT"),
		ExternalDocs("https://docs.example.com", "docs"),
		Security("bearerAuth", "read"),
		Security(""),
		Server("https://{region}.example.com", "prod",
			server.Name("production"),
			server.Variable("region", "us-east",
				server.VariableEnum("us-east", "eu-west"),
				server.VariableDescription("deployment region"),
			),
		),
		Tag("health", "Health checks",
			tag.Summary("ops"),
			tag.ExternalDocs("https://example.com/health", "health docs"),
		),
		SecurityScheme("bearerAuth",
			securityscheme.HTTP("bearer"),
			securityscheme.BearerFormat("JWT"),
		),
	}
	for _, opt := range opts {
		opt(doc, st)
	}

	if got := doc.Version; got != "3.0.3" {
		t.Fatalf("Version() = %q, want %q", got, "3.0.3")
	}
	if got := doc.Info.Summary; got != "summary" {
		t.Fatalf("Summary() = %q, want %q", got, "summary")
	}
	if doc.Info.Contact == nil || doc.Info.Contact.Email != "api@example.com" {
		t.Fatalf("Contact() did not set contact: %#v", doc.Info.Contact)
	}
	if doc.Info.License == nil || doc.Info.License.Identifier != "MIT" {
		t.Fatalf("License() did not set identifier: %#v", doc.Info.License)
	}
	if doc.ExternalDocs == nil || doc.ExternalDocs.URL != "https://docs.example.com" {
		t.Fatalf("ExternalDocs() did not set external docs: %#v", doc.ExternalDocs)
	}
	if len(doc.Security) != 2 || !doc.Security[1].ContainsEmptyRequirement {
		t.Fatalf("Security() did not append expected requirements: %#v", doc.Security)
	}
	if len(doc.Servers) != 1 || doc.Servers[0].Name != "production" {
		t.Fatalf("Server() did not add server with options: %#v", doc.Servers)
	}
	region, ok := doc.Servers[0].Variables.Get("region")
	if !ok || region.Default != "us-east" || region.Description != "deployment region" {
		t.Fatalf("Server() did not configure variable: %#v ok=%v", region, ok)
	}
	if len(doc.Tags) != 1 || doc.Tags[0].Summary != "ops" {
		t.Fatalf("Tag() did not append expected tag: %#v", doc.Tags)
	}
	if ref, ok := st.GetSecurityScheme("bearerAuth"); !ok || ref.Reference != "#/components/securitySchemes/bearerAuth" {
		t.Fatalf("SecurityScheme() did not store component ref: %#v ok=%v", ref, ok)
	}
}

func TestWebhook_UsesPathItemBuilder(t *testing.T) {
	doc := &v3.Document{Info: &base.Info{}}
	st := store.New(&v3.Components{})

	p := pathitem.New(pathitem.Description("Webhook delivery"))
	p.Operation(http.MethodPost, op.Summary("deliver"))

	Webhook("user.created", p.PathItem())(doc, st)

	if doc.Webhooks == nil {
		t.Fatal("Webhook() did not initialize Webhooks")
	}
	item, ok := doc.Webhooks.Get("user.created")
	if !ok {
		t.Fatal("Webhook() did not store path item")
	}
	if item.Post == nil || item.Post.Summary != "deliver" {
		t.Fatalf("Webhook() stored wrong path item: %#v", item)
	}
}
