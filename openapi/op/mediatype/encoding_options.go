package mediatype

import v3 "github.com/pb33f/libopenapi/datamodel/high/v3"

// EncodingOption configures an OpenAPI encoding object.
type EncodingOption func(*v3.Encoding)

// EncodingContentType sets the Content-Type for encoding a specific property.
func EncodingContentType(s string) EncodingOption {
	return func(e *v3.Encoding) {
		e.ContentType = s
	}
}

// EncodingStyle sets the serialization style for the encoding.
func EncodingStyle(s string) EncodingOption {
	return func(e *v3.Encoding) {
		e.Style = s
	}
}

// EncodingExplode controls whether arrays and objects generate separate parameters.
func EncodingExplode(v bool) EncodingOption {
	return func(e *v3.Encoding) {
		e.Explode = &v
	}
}

// EncodingAllowReserved allows reserved URI characters in the encoding value.
func EncodingAllowReserved() EncodingOption {
	return func(e *v3.Encoding) {
		e.AllowReserved = true
	}
}
