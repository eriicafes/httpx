module github.com/eriicafes/httpx

go 1.25.4

require (
	github.com/MarceloPetrucio/go-scalar-api-reference v0.0.0-20240521013641-ce5d2efe0e06
	github.com/eriicafes/union v0.0.0
	github.com/pb33f/libopenapi v0.34.0
	go.yaml.in/yaml/v4 v4.0.0-rc.4
)

require (
	github.com/bahlo/generic-list-go v0.2.0 // indirect
	github.com/buger/jsonparser v1.1.1 // indirect
	github.com/pb33f/jsonpath v0.8.1 // indirect
	github.com/pb33f/ordered-map/v2 v2.3.0 // indirect
	golang.org/x/sync v0.19.0 // indirect
)

replace github.com/eriicafes/union => ./internal/testdeps/union
