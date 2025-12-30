module forecast

go 1.20

require github.com/massive-com/client-go/v2 v2.23.0

// Required transitive dependencies (auto-filled by `go mod tidy`)
require (
	github.com/google/uuid v1.6.0
	github.com/stretchr/testify v1.11.1
	golang.org/x/net v0.32.0
)

// Massive client uses test imports that pull in older versions;
// replace to avoid conflicts and ensure security
replace github.com/stretchr/testify => github.com/stretchr/testify v1.11.1
