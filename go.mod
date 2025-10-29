module github.com/bss/radb-client

go 1.23

require (
	github.com/spf13/cobra v1.8.1
	github.com/spf13/viper v1.19.0
	github.com/zalando/go-keyring v0.2.5
	golang.org/x/crypto v0.28.0
)

// Additional dependencies will be added as needed:
// - HTTP client with retry logic
// - JSON/YAML processing
// - Terminal UI (potentially charmbracelet/bubbletea or fatih/color)
// - Diff generation
// - Testing utilities
