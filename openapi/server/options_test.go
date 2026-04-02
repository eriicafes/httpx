package server

import (
	"testing"

	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

func TestNewAndVariableOptions(t *testing.T) {
	s := New("https://{region}.example.com", "prod",
		Name("production"),
		Variable("region", "us-east",
			VariableEnum("us-east", "eu-west"),
			VariableDescription("deployment region"),
		),
	)
	if s.Name != "production" || s.URL != "https://{region}.example.com" {
		t.Fatalf("New() incorrect: %#v", s)
	}
	region, ok := s.Variables.Get("region")
	if !ok || region.Default != "us-east" || len(region.Enum) != 2 || region.Description != "deployment region" {
		t.Fatalf("Variable() incorrect: %#v ok=%v", region, ok)
	}

	var sv v3.ServerVariable
	VariableEnum("a", "b")(&sv)
	VariableDescription("desc")(&sv)
	if len(sv.Enum) != 2 || sv.Description != "desc" {
		t.Fatalf("VariableOption helpers incorrect: %#v", sv)
	}
}
