package paramdescriptions

// Source of truth for CLI -h parameter descriptions.
// After replacing params.json, regenerate bindata:
//
//go:generate go-bindata -pkg paramdescriptions -o bindata.go params.json
//go:generate gofmt -w bindata.go
