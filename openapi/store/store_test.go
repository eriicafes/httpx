package store

import (
	"reflect"
	"testing"

	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

func TestStore_ComponentRefsAndDeduplication(t *testing.T) {
	components := &v3.Components{}
	s := New(components)

	type User struct{}
	userType := reflect.TypeFor[User]()
	s.SetSchema(userType, "User", base.CreateSchemaProxy(&base.Schema{Title: "User"}))
	if got, ok := s.GetSchema(reflect.TypeFor[*User]()); !ok || got == nil {
		t.Fatalf("GetSchema() did not dereference pointer types: %#v ok=%v", got, ok)
	}
	s.SetSchema(userType, "UserOverride", base.CreateSchemaProxy(&base.Schema{Title: "Override"}))
	if components.Schemas.Len() != 1 {
		t.Fatalf("SetSchema() should dedupe by type, got %d schemas", components.Schemas.Len())
	}

	paramRef := s.SetParameter("ID", &v3.Parameter{Name: "id", Reference: "ID"})
	reqBodyRef := s.SetRequestBody("CreateUser", &v3.RequestBody{Reference: "CreateUser"})
	respRef := s.SetResponse("UserResponse", &v3.Response{Reference: "UserResponse"})
	headerRef := s.SetHeader("RequestID", &v3.Header{Reference: "RequestID"})
	pathRef := s.SetPathItem("Users", &v3.PathItem{Reference: "Users"})
	linkRef := s.SetLink("UserLink", &v3.Link{Reference: "UserLink"})
	exampleRef := s.SetExample("UserExample", &base.Example{Reference: "UserExample"})
	callbackRef := s.SetCallback("OnUserCreated", &v3.Callback{Reference: "OnUserCreated"})
	securityRef := s.SetSecurityScheme("bearerAuth", &v3.SecurityScheme{Reference: "bearerAuth"})

	assertRef := func(name, got, want string) {
		t.Helper()
		if got != want {
			t.Fatalf("%s ref = %q, want %q", name, got, want)
		}
	}
	assertRef("parameter", paramRef.Reference, "#/components/parameters/ID")
	assertRef("request body", reqBodyRef.Reference, "#/components/requestBodies/CreateUser")
	assertRef("response", respRef.Reference, "#/components/responses/UserResponse")
	assertRef("header", headerRef.Reference, "#/components/headers/RequestID")
	assertRef("path item", pathRef.Reference, "#/components/pathItems/Users")
	assertRef("link", linkRef.Reference, "#/components/links/UserLink")
	assertRef("example", exampleRef.Reference, "#/components/examples/UserExample")
	assertRef("callback", callbackRef.Reference, "#/components/callbacks/OnUserCreated")
	assertRef("security scheme", securityRef.Reference, "#/components/securitySchemes/bearerAuth")

	if stored, _ := components.Parameters.Get("ID"); stored.Reference != "" {
		t.Fatalf("SetParameter() should unset stored reference, got %#v", stored.Reference)
	}
	if stored, _ := components.RequestBodies.Get("CreateUser"); stored.Reference != "" {
		t.Fatalf("SetRequestBody() should unset stored reference, got %#v", stored.Reference)
	}
}
