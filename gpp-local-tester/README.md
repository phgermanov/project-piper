# GPP Local Testing Tool

A comprehensive Go-based tool for testing GPP (GitHub Pipeline Platform) changes locally using [act](https://github.com/nektos/act).

## Overview

This tool allows you to test changes to jenkins-library, piper-library, and github-actions locally without pushing to GitHub. It provides:

- ✅ **Single configuration file** (`config.yaml`) for all settings
- ✅ **Local Piper binary** usage instead of downloading from GitHub releases
- ✅ **Mock external services** (Vault, Cumulus, SonarQube, etc.)
- ✅ **Example npm project** pre-configured with Piper
- ✅ **Automated test execution** with act
- ✅ **Written in Go** - fast, compiled, zero dependencies

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│              GPP Local Testing Tool (Go)                     │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐   │
│  │ Builder  │  │  Mock    │  │ Runner   │  │  Config  │   │
│  │          │  │  Server  │  │          │  │          │   │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘   │
└─────────────────────────────────────────────────────────────┘
                              │
              ┌───────────────┼───────────────┐
              ▼               ▼               ▼
        ┌──────────┐    ┌──────────┐   ┌──────────────┐
        │  Local   │    │   Mock   │   │ Example NPM  │
        │  Piper   │    │  Server  │   │   Project    │
        │  Binary  │    │  (HTTP)  │   │ (with Piper) │
        └──────────┘    └──────────┘   └──────────────┘
              │               │               │
              └───────────────┼───────────────┘
                              ▼
                        ┌──────────┐
                        │   Act    │
                        │ (GitHub  │
                        │ Actions) │
                        └──────────┘
```

## Prerequisites

### Required

- **[act](https://github.com/nektos/act)** - Local GitHub Actions runner
  ```bash
  # macOS
  brew install act

  # Linux
  curl https://raw.githubusercontent.com/nektos/act/master/install.sh | sudo bash
  ```

- **[Docker](https://www.docker.com/)** - Container runtime for act
  ```bash
  # Verify Docker is running
  docker ps
  ```

- **[Go](https://golang.org/)** - v1.21 or later
  ```bash
  go version
  ```

## Installation

### Option 1: Build from Source (Recommended)

```bash
cd gpp-local-tester
go build -o gpp-test .
```

### Option 2: Install Globally

```bash
cd gpp-local-tester
go install .
```

## Quick Start

1. **Build the tool**:
   ```bash
   cd gpp-local-tester
   go build -o gpp-test .
   ```

2. **Run the complete test**:
   ```bash
   ./gpp-test run
   ```

The tool will automatically:
- Check prerequisites
- Build the Piper binary from your local jenkins-library
- Start the mock server
- Configure act environment
- Execute the pipeline with act
- Display results

## Commands

### `gpp-test run`

Run the complete GPP pipeline locally.

```bash
# Run with default config
./gpp-test run

# Run with custom config
./gpp-test run --config /path/to/config.yaml

# Skip building Piper binary
./gpp-test run --skip-build

# Skip mock server
./gpp-test run --skip-mock

# Run specific workflow
./gpp-test run --workflow .github/workflows/custom.yml

# Verbose output
./gpp-test run --verbose
```

### `gpp-test build`

Build the Piper binary only.

```bash
# Build Piper binary
./gpp-test build

# Force rebuild
./gpp-test build --force
```

### `gpp-test mock`

Run the mock server only.

```bash
# Start mock server
./gpp-test mock

# With custom config
./gpp-test mock --config config.yaml
```

Press Ctrl+C to stop the server and see statistics.

## Directory Structure

```
gpp-local-tester/
├── cmd/                       # CLI commands
│   ├── root.go               # Root command
│   ├── run.go                # Run pipeline command
│   ├── build.go              # Build Piper command
│   └── mock.go               # Mock server command
│
├── internal/                  # Internal packages
│   ├── config/               # Configuration loading
│   ├── mock/                 # Mock HTTP server
│   ├── builder/              # Piper binary builder
│   └── runner/               # Act runner
│
├── example-project/           # Example npm project with Piper
│   ├── .github/workflows/
│   │   └── piper.yml         # Workflow using Piper
│   ├── .pipeline/
│   │   └── config.yml        # Piper configuration
│   ├── package.json          # npm configuration
│   ├── index.js              # Application code
│   └── README.md
│
├── bin/                       # Built binaries
│   └── piper                 # Local Piper binary
│
├── main.go                    # Main entry point
├── go.mod                     # Go module definition
├── config.yaml                # Configuration file
├── gpp-test                   # Built tool binary
└── README.md                  # This file
```

## Configuration

All configuration is managed through `config.yaml`. See the file for detailed documentation.

### Key Configuration Sections

#### Piper Configuration

```yaml
piper:
  binaryPath: "./bin/piper"           # Path to local binary
  autoBuild: true                     # Build if missing
  sourcePath: "../jenkins-library"   # Source code location
  version: "local-dev"                # Version identifier
```

#### Mock Services

```yaml
mockServices:
  enabled: true
  host: "localhost"
  port: 8888
  endpoints:
    - path: "/v1/auth/approle/login"
      method: "POST"
      response:
        status: 200
        body:
          auth:
            client_token: "mock-vault-token-12345"
```

Add your own mock endpoints as needed!

#### Secrets

```yaml
secrets:
  PIPER_VAULTAPPROLEID: "mock-approle-id"
  PIPER_VAULTAPPROLESECRETID: "mock-approle-secret"
  GITHUB_TOKEN: "mock-github-token"
```

## Usage Examples

### Test Your Local Changes

1. **Modify jenkins-library code**:
   ```bash
   cd ../jenkins-library
   # Make your changes to Piper steps
   vim cmd/myStep.go
   ```

2. **Run local test**:
   ```bash
   cd ../gpp-local-tester
   ./gpp-test run
   ```

3. **Review results**:
   - Pipeline output in terminal
   - Mock server statistics at the end

### Debug Mode

```bash
# Run with verbose output
./gpp-test run --verbose

# Just test the mock server
./gpp-test mock

# Just build Piper
./gpp-test build --force
```

### Test Different Workflows

```bash
# Test custom workflow
./gpp-test run --workflow .github/workflows/full-piper.yml
```

## How It Works

### 1. Local Piper Binary Injection

The tool:
1. Builds Piper from your local `jenkins-library` source using Go
2. Uses Docker bind mounts to inject the binary into act containers
3. The binary is available at `/workspace/gpp-local-tester/bin/piper`

### 2. Mock Service Server

A lightweight Go HTTP server that:
- Listens on `localhost:8888`
- Responds to API calls from Piper steps
- Logs all requests with colored output
- Returns configurable mock responses
- Provides request statistics

Services mocked:
- **Vault** - Secret management
- **Cumulus** - Artifact upload
- **SonarQube** - Code quality
- **System Trust** - Session tokens
- **GitHub API** - Releases, PRs

### 3. Act Configuration

The tool automatically creates:
- `.secrets` - Secret values for the pipeline
- `.env` - Environment variables
- `.actrc` - Act runtime configuration

These files configure act to:
- Use appropriate Docker images
- Mount local directories
- Inject secrets and variables
- Connect to mock services via `host.docker.internal`

### 4. Pipeline Execution

The tool:
1. Checks all prerequisites (act, docker, go)
2. Builds Piper binary (if needed)
3. Starts mock server in background
4. Prepares act environment
5. Executes `act` with proper configuration
6. Shows real-time output
7. Displays statistics at the end

## Mock Service Endpoints

The mock server provides the following endpoints by default:

| Service | Endpoint | Method | Purpose |
|---------|----------|--------|---------|
| Vault | `/v1/auth/approle/login` | POST | AppRole authentication |
| Vault | `/v1/secret/data/piper/*` | GET | Secret retrieval |
| Cumulus | `/api/v1/upload` | POST | Artifact upload |
| Cumulus | `/api/v1/artifacts/*` | GET | Artifact list |
| SonarQube | `/api/ce/task` | GET | Analysis status |
| SonarQube | `/api/qualitygates/project_status` | GET | Quality gate status |
| System Trust | `/api/session/token` | POST | Session token |
| GitHub | `/repos/*/releases/latest` | GET | Latest release |
| Generic | `/*` | * | Catch-all success |

All endpoints return JSON responses and support CORS.

## Troubleshooting

### Mock server connection failed

**Problem**: Pipeline can't reach `host.docker.internal:8888`

**Solution**:
- On Linux, the tool automatically adds `--add-host=host.docker.internal:host-gateway`
- On macOS/Windows, `host.docker.internal` works by default

### Act fails with permission errors

**Problem**: Docker permission denied

**Solution**:
```bash
# Add user to docker group (Linux)
sudo usermod -aG docker $USER
newgrp docker
```

### Piper build fails

**Problem**: Go build errors

**Solution**:
```bash
# Try building manually
cd ../jenkins-library
go build -o ../gpp-local-tester/bin/piper .

# If network issues, try:
go env -w GOPROXY=https://proxy.golang.org,direct
```

### Act can't find workflow

**Problem**: Workflow file not found

**Solution**:
- Ensure workflow path is relative to example-project directory
- Check that `.github/workflows/` exists
- Default is `.github/workflows/piper.yml`

## Extending the Tool

### Add Custom Mock Responses

Edit `config.yaml`:
```yaml
mockServices:
  endpoints:
    - path: "/api/custom-service"
      method: "POST"
      response:
        status: 200
        body:
          customField: "customValue"
```

### Test Different Projects

1. Copy `example-project` to a new directory
2. Update `config.yaml`:
   ```yaml
   exampleProject:
     path: "./my-custom-project"
   ```
3. Modify `.pipeline/config.yml` for your project type
4. Run: `./gpp-test run`

### Add Code Features

The tool is structured for easy extension:

- **Add new commands**: Create new file in `cmd/`
- **Add new mock features**: Modify `internal/mock/server.go`
- **Add new builders**: Extend `internal/builder/`
- **Add new runners**: Extend `internal/runner/`

Example - add a new command:

```go
// cmd/validate.go
package cmd

import (
	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate configuration",
	RunE:  runValidate,
}

func init() {
	rootCmd.AddCommand(validateCmd)
}

func runValidate(cmd *cobra.Command, args []string) error {
	// Your code here
	return nil
}
```

## Performance Tips

1. **Reuse Piper binary**: Use `--skip-build` after first run
   ```bash
   ./gpp-test run --skip-build
   ```

2. **Pre-pull Docker images**: Cache act images
   ```bash
   docker pull ghcr.io/catthehacker/ubuntu:full-latest
   ```

3. **Disable unnecessary stages**: In `.pipeline/config.yml`
   ```yaml
   stages:
     Integration:
       active: false
   ```

## Development

### Running Tests

```bash
go test ./...
```

### Building

```bash
# Build for current platform
go build -o gpp-test .

# Build for Linux
GOOS=linux GOARCH=amd64 go build -o gpp-test-linux .

# Build for macOS
GOOS=darwin GOARCH=amd64 go build -o gpp-test-darwin .
```

### Dependencies

The tool uses minimal dependencies:
- `github.com/spf13/cobra` - CLI framework
- `github.com/fatih/color` - Colored output
- `gopkg.in/yaml.v3` - YAML parsing

## Contributing

1. Make your changes
2. Test thoroughly
3. Update documentation
4. Submit PR

## License

Apache-2.0

## Support

For issues or questions:
- Check the [Troubleshooting](#troubleshooting) section
- Run with `--verbose` flag
- Check mock server logs
- Review act output

## Related Documentation

- [act Documentation](https://github.com/nektos/act)
- [Project Piper](https://www.project-piper.io/)
- [GPP Documentation](../piper-pipeline-github/README.md)
- [Jenkins Library](../jenkins-library/README.md)

## Features

### ✨ Highlights

- **Fast**: Compiled Go binary, instant startup
- **Simple**: Single binary, single config file
- **Comprehensive**: Mocks all external services
- **Flexible**: Easy to customize and extend
- **Developer-friendly**: Colored output, clear error messages
- **Production-ready**: Used for testing GPP changes before deployment

### 🎯 Use Cases

- Test Piper step changes locally
- Debug GPP workflow issues
- Validate pipeline configurations
- Develop new Piper steps
- Train new team members
- CI/CD pipeline development
