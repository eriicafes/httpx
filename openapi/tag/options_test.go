package tag

import "testing"

func TestNewAndOptions(t *testing.T) {
	tag := New("health", "Health checks",
		Summary("ops"),
		ExternalDocs("https://docs.example.com/health", "docs"),
		Parent("ops"),
		Kind("nav"),
	)
	if tag.Name != "health" || tag.Description != "Health checks" {
		t.Fatalf("New() incorrect: %#v", tag)
	}
	if tag.Summary != "ops" || tag.Parent != "ops" || tag.Kind != "nav" {
		t.Fatalf("tag options incorrect: %#v", tag)
	}
	if tag.ExternalDocs == nil || tag.ExternalDocs.URL != "https://docs.example.com/health" {
		t.Fatalf("ExternalDocs() incorrect: %#v", tag.ExternalDocs)
	}
}
