package securityscheme

import (
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
)

// OAuthFlowsOption configures the OAuth2 flows object.
type OAuthFlowsOption func(*v3.OAuthFlows)

// OAuthFlowOption configures a single OAuth2 flow.
type OAuthFlowOption func(*v3.OAuthFlow)

// FlowImplicit configures the OAuth2 implicit flow.
func FlowImplicit(authorizationUrl string, opts ...OAuthFlowOption) OAuthFlowsOption {
	return func(flows *v3.OAuthFlows) {
		flows.Implicit = buildFlow("", authorizationUrl, opts)
	}
}

// FlowPassword configures the OAuth2 resource owner password flow.
func FlowPassword(tokenUrl string, opts ...OAuthFlowOption) OAuthFlowsOption {
	return func(flows *v3.OAuthFlows) {
		flows.Password = buildFlow(tokenUrl, "", opts)
	}
}

// FlowClientCredentials configures the OAuth2 client credentials flow.
func FlowClientCredentials(tokenUrl string, opts ...OAuthFlowOption) OAuthFlowsOption {
	return func(flows *v3.OAuthFlows) {
		flows.ClientCredentials = buildFlow(tokenUrl, "", opts)
	}
}

// FlowAuthorizationCode configures the OAuth2 authorization code flow.
func FlowAuthorizationCode(authorizationUrl, tokenUrl string, opts ...OAuthFlowOption) OAuthFlowsOption {
	return func(flows *v3.OAuthFlows) {
		flows.AuthorizationCode = buildFlow(tokenUrl, authorizationUrl, opts)
	}
}

// FlowDevice configures the OAuth2 device authorization flow (OpenAPI 3.2+).
func FlowDevice(tokenUrl string, opts ...OAuthFlowOption) OAuthFlowsOption {
	return func(flows *v3.OAuthFlows) {
		flows.Device = buildFlow(tokenUrl, "", opts)
	}
}

// FlowScope adds a scope to the OAuth2 flow.
func FlowScope(name, description string) OAuthFlowOption {
	return func(flow *v3.OAuthFlow) {
		if flow.Scopes == nil {
			flow.Scopes = orderedmap.New[string, string]()
		}
		flow.Scopes.Set(name, description)
	}
}

// FlowRefreshUrl sets the refresh URL for the OAuth2 flow.
func FlowRefreshUrl(url string) OAuthFlowOption {
	return func(flow *v3.OAuthFlow) {
		flow.RefreshUrl = url
	}
}

func buildFlow(tokenUrl, authorizationUrl string, opts []OAuthFlowOption) *v3.OAuthFlow {
	flow := &v3.OAuthFlow{TokenUrl: tokenUrl, AuthorizationUrl: authorizationUrl}
	for _, opt := range opts {
		opt(flow)
	}
	return flow
}
