package mediatype

import (
	"testing"

	"github.com/eriicafes/httpx/openapi/op/example"
	"github.com/eriicafes/httpx/openapi/store"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

type mediaDefaults struct{}

func (mediaDefaults) MediaType() Option {
	return NamedExample("fromType", "value")
}

func TestNew_AppliesOptions(t *testing.T) {
	st := store.New(nil)

	m := New[string](st,
		Example("inline"),
		NamedExample("sample", "value", example.Summary("sample")),
		NamedExampleRef("stored", "StoredExample"),
		ItemSchema[string](),
		Encoding("avatar", EncodingContentType("image/png"), EncodingStyle("form"), EncodingExplode(true), EncodingAllowReserved()),
		ItemEncoding("payload", EncodingContentType("application/json")),
	)
	if m.Example == nil {
		t.Fatal("Example() did not set media example")
	}
	if m.Examples == nil || m.Examples.Len() != 2 {
		t.Fatalf("expected 2 inline examples, got %d", m.Examples.Len())
	}
	if m.ItemSchema == nil || m.ItemSchema.Schema().Type[0] != "string" {
		t.Fatalf("ItemSchema() incorrect: hasSchema=%t type=%v", m.ItemSchema != nil, func() []string { if m.ItemSchema == nil || m.ItemSchema.Schema() == nil { return nil }; return m.ItemSchema.Schema().Type }())
	}
	if m.Encoding == nil || m.ItemEncoding == nil {
		t.Fatalf("Encoding() / ItemEncoding() not set: encodingNil=%t itemEncodingNil=%t", m.Encoding == nil, m.ItemEncoding == nil)
	}
	enc, _ := m.Encoding.Get("avatar")
	if enc == nil || enc.ContentType != "image/png" || enc.Explode == nil || !*enc.Explode || !enc.AllowReserved {
		t.Fatalf("Encoding() incorrect: nil=%t contentType=%q explode=%v allowReserved=%t", enc == nil, func() string { if enc == nil { return "" }; return enc.ContentType }(), func() any { if enc == nil { return nil }; return enc.Explode }(), func() bool { if enc == nil { return false }; return enc.AllowReserved }())
	}
}

func TestNew_AppliesTypeDefaults(t *testing.T) {
	st := store.New(nil)

	m := New[mediaDefaults](st)
	if m.Examples == nil || m.Examples.Len() != 1 {
		t.Fatalf("type defaults not applied: count=%d", m.Examples.Len())
	}
}

func TestEncodingOptionsStandalone(t *testing.T) {
	enc := &v3.Encoding{}
	EncodingContentType("image/png")(enc)
	EncodingStyle("form")(enc)
	EncodingExplode(true)(enc)
	EncodingAllowReserved()(enc)
	if enc.ContentType != "image/png" || enc.Style != "form" || enc.Explode == nil || !*enc.Explode || !enc.AllowReserved {
		t.Fatalf("EncodingOption helpers incorrect: contentType=%q style=%q explode=%v allowReserved=%t", enc.ContentType, enc.Style, enc.Explode, enc.AllowReserved)
	}
}
