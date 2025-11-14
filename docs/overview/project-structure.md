# Project Piper - Repository Structure

This document provides a detailed overview of the Project Piper monorepo structure and the purpose of each component.

## Monorepo Organization

Project Piper is organized as a **monorepo** containing 9 main directories, each representing a separate but related component of the Piper CI/CD ecosystem.

```
project-piper/
├── jenkins-library/                    # Core Piper library (Go + Groovy)
├── piper-pipeline-github/             # GitHub Actions workflow templates
├── piper-pipeline-azure/              # Azure DevOps pipeline templates
├── project-piper-action/              # GitHub Action implementation
├── piper-azure-task/                  # Azure DevOps task extension
├── resources/                         # Shared configuration resources
├── project-piper.github.tools.sap/    # Documentation website
├── piper-simple-go/                   # Demo Go application
└── gha-demo-k8s-node/                 # Demo Node.js application
```

---

## 1. jenkins-library (Core Piper Library)

**Technology**: Go 1.24.0 + Groovy
**Status**: No longer accepting contributions (archived)

### Purpose
The core Project Piper library containing both:
- Standalone Go CLI tool (piper binary)
- Jenkins Shared Library (Groovy scripts)

### Directory Structure

```
jenkins-library/
├── cmd/                          # Go CLI commands (~200+ commands)
│   ├── piper.go                  # Main entry point
│   ├── mavenBuild.go             # Maven build step
│   ├── kanikoExecute.go          # Kaniko container build
│   └── ...                       # All other steps
├── pkg/                          # Go packages (70+ packages)
│   ├── config/                   # Configuration management
│   ├── piperenv/                 # Common Pipeline Environment
│   ├── docker/                   # Docker utilities
│   ├── kubernetes/               # Kubernetes utilities
│   ├── maven/                    # Maven utilities
│   ├── npm/                      # npm utilities
│   └── ...                       # Technology-specific packages
├── vars/                         # Groovy pipeline steps (188 files)
│   ├── piperPipeline.groovy      # Main general purpose pipeline
│   ├── mavenBuild.groovy         # Maven build wrapper
│   ├── kanikoExecute.groovy      # Kaniko wrapper
│   └── ...                       # All pipeline steps
├── src/                          # Groovy source classes (59 files)
│   └── com/sap/piper/            # Utility classes
│       ├── ConfigurationHelper.groovy
│       ├── GitUtils.groovy
│       ├── DockerUtils.groovy
│       └── ...
├── resources/                    # Pipeline resources
│   ├── metadata/                 # Step metadata (153+ YAML files)
│   │   ├── mavenBuild.yaml
│   │   ├── kanikoExecute.yaml
│   │   └── ...
│   ├── com.sap.piper/pipeline/  # Pipeline defaults
│   │   ├── stageDefaults.yml
│   │   └── stashSettings.yml
│   ├── default_pipeline_environment.yml
│   └── schemas/                  # JSON schemas
│       └── metadata.json
├── documentation/                # Comprehensive documentation
│   ├── docs/                     # Markdown documentation
│   │   ├── steps/                # Step documentation (153 files)
│   │   ├── stages/               # Stage documentation
│   │   └── scenarios/            # Use case scenarios
│   └── mkdocs.yml                # MkDocs configuration
├── integration/                  # Integration tests
├── test/                         # Unit tests
├── go.mod                        # Go module definition
├── main.go                       # Go CLI entry point
├── Makefile                      # Build automation
└── README.md
```

### Key Features

**Go CLI (cmd/):**
- Standalone binary that can run on any platform
- ~200+ CI/CD step implementations
- Container execution support
- Multi-platform support (Linux, macOS, Windows)

**Groovy Library (vars/):**
- 188 Jenkins pipeline steps
- Jenkins-native integration (credentials, stashing, nodes)
- General Purpose Pipeline orchestration
- Stage wrappers and helpers

**Metadata System (resources/metadata/):**
- YAML-based step definitions
- Type-safe parameter validation
- Automatic documentation generation
- Code generation from metadata

---

## 2. piper-pipeline-github (GitHub Pipeline Templates)

**Technology**: GitHub Actions (YAML workflows)

### Purpose
Piper's General Purpose Pipeline (GPP) for GitHub Actions - reusable workflow templates.

### Directory Structure

```
piper-pipeline-github/
├── .github/
│   └── workflows/                # Reusable workflows
│       ├── sap-piper-workflow.yml          # Main GPP workflow
│       ├── init.yml                        # Init stage
│       ├── build.yml                       # Build stage
│       ├── integration.yml                 # Integration stage
│       ├── acceptance.yml                  # Acceptance stage
│       ├── performance.yml                 # Performance stage
│       ├── promote.yml                     # Promote stage
│       ├── release.yml                     # Release stage
│       ├── post.yml                        # Post stage
│       ├── oss.yml                         # OSS scanning stage
│       ├── ppms.yml                        # PPMS compliance stage
│       ├── sap-oss-ppms-workflow.yml       # OSS+PPMS workflow
│       └── sap-ctp-scan.yml                # CTP scan workflow
├── docs/                         # Documentation
│   ├── adr/                      # Architecture Decision Records
│   └── extensibility.md
├── examples/                     # Example workflows
├── metadata.yaml                 # Pipeline metadata
├── .log4brains.yml               # ADR management
└── README.md
```

### Key Features

- **Stage-based workflows**: Modular, reusable stage templates
- **Extensibility**: Pre/post hooks for each stage
- **GitHub Environment integration**: Manual approval gates
- **System Trust integration**: Token management
- **Cumulus integration**: Evidence upload for compliance
- **Conditional execution**: Configurable stage activation

---

## 3. piper-pipeline-azure (Azure Pipeline Templates)

**Technology**: Azure DevOps YAML pipelines

### Purpose
Piper's general purpose pipeline for Azure DevOps with reusable pipeline templates.

### Directory Structure

```
piper-pipeline-azure/
├── sap-piper-pipeline.yml        # Main pipeline template
├── stages/                       # Stage definitions
│   ├── init.yml
│   ├── build.yml
│   ├── integration.yml
│   ├── acceptance.yml
│   ├── performance.yml
│   ├── promote.yml
│   ├── release.yml
│   └── post.yml
├── steps/                        # Reusable steps
│   └── piper-step.yml
├── metadata.yaml                 # Pipeline metadata
├── docs/                         # Documentation
│   └── adr/                      # Architecture Decision Records
└── README.md
```

### Key Features

- **Template-based**: Reusable stage and step templates
- **Variable groups**: Azure-native secret management
- **Service connections**: Azure service principal integration
- **Artifact publishing**: Azure Artifacts integration

---

## 4. project-piper-action (GitHub Action)

**Technology**: TypeScript/Node.js (v16.17.1+)

### Purpose
GitHub Action that allows running Piper steps in GitHub Actions workflows.

### Directory Structure

```
project-piper-action/
├── src/                          # TypeScript source code
│   ├── main.ts                   # Action entry point
│   ├── piper.ts                  # Piper execution logic
│   ├── setup.ts                  # Binary setup
│   └── cpe.ts                    # CPE management
├── dist/                         # Compiled JavaScript
│   └── index.js                  # Bundled action
├── test/                         # Unit tests
├── docs/                         # Documentation
│   └── adr/                      # Architecture Decision Records
├── action.yml                    # Action metadata
├── package.json                  # npm configuration
├── tsconfig.json                 # TypeScript configuration
└── README.md
```

### Key Features

- **Binary management**: Downloads/builds Piper binary
- **Caching**: Binary caching for performance
- **CPE persistence**: Artifact upload/download for state
- **Container support**: Docker execution integration
- **Error handling**: Comprehensive error reporting

---

## 5. piper-azure-task (Azure Task Extension)

**Technology**: Azure DevOps Extension

### Purpose
Azure DevOps task extension that wraps Piper binary execution.

### Directory Structure

```
piper-azure-task/
├── piper/                        # Task implementation
│   ├── task.json                 # Task definition
│   └── index.js                  # Task logic
├── docs/                         # Documentation
│   └── adr/                      # Architecture Decision Records
├── vss-extension.json            # Extension manifest
└── README.md
```

### Key Features

- **Marketplace publishing**: Available on Azure Marketplace
- **Task definition**: Azure-native task configuration
- **Binary execution**: Wraps Piper CLI
- **Credential management**: Azure credential integration

---

## 6. resources (Configuration Resources)

**Technology**: Node.js/YAML

### Purpose
Centralized configuration files for Piper binary and general purpose pipelines.

### Directory Structure

```
resources/
├── defaults/                     # Default configurations
│   ├── defaults.yml              # Base defaults
│   ├── deviation-jenkins.yml    # Jenkins-specific
│   ├── deviation-azure.yml      # Azure-specific
│   ├── deviation-github.yml     # GitHub-specific
│   ├── deviation-tools.yml      # Tools landscape
│   └── deviation-wdf.yml        # WDF landscape
├── gen/                          # Generated combined configs
│   ├── piper-defaults-jenkins-wdf.yml
│   ├── piper-defaults-azure-tools.yml
│   ├── piper-defaults-github-tools.yml
│   └── piper-defaults-github-wdf.yml
├── stageconfig/                  # Stage configurations
│   └── piper-stage-config.yml
├── package.json                  # Build scripts
└── README.md
```

### Key Features

- **Hierarchical configuration**: Base + platform-specific + instance-specific
- **Build system**: Generates combined configuration files
- **Stage definitions**: Complete pipeline structure
- **Platform deviations**: Orchestrator-specific overrides

---

## 7. project-piper.github.tools.sap (Documentation Site)

**Technology**: MkDocs with Material theme

### Purpose
Main documentation website for Project Piper (deployed to GitHub Pages).

### Directory Structure

```
project-piper.github.tools.sap/
├── docs/                         # Documentation source
│   ├── index.md                  # Homepage
│   ├── guidedtour.md            # Getting started
│   ├── configuration.md          # Configuration guide
│   ├── extensibility.md          # Extensibility guide
│   ├── steps/                    # Step documentation
│   ├── stages/                   # Stage documentation
│   ├── scenarios/                # Use case scenarios
│   ├── infrastructure/           # Infrastructure setup
│   ├── images/                   # Documentation images
│   └── ...
├── mkdocs.yml                    # MkDocs configuration
├── docker-compose.yml            # Local development
└── README.md
```

### Key Features

- **MkDocs Material**: Modern documentation theme
- **Search**: Full-text search capability
- **Navigation**: Structured documentation hierarchy
- **Code examples**: Syntax-highlighted code samples
- **Local preview**: Docker-based local development

---

## 8. piper-simple-go (Demo Go Application)

**Technology**: Go 1.20.0

### Purpose
Simple Go application demonstrating Piper usage with a Go project.

### Directory Structure

```
piper-simple-go/
├── main.go                       # Go application
├── .pipeline/                    # Piper configuration
│   └── config.yml
├── go.mod                        # Go module
├── VERSION                       # Version file
└── README.md
```

### Key Features

- **Minimal example**: Simple HTTP server using gorilla/mux
- **Pipeline configuration**: Example Piper configuration
- **Version management**: Demonstrates artifact versioning

---

## 9. gha-demo-k8s-node (Demo Node.js Application)

**Technology**: Node.js/Express

### Purpose
Demo Node.js application demonstrating GitHub Actions integration with Kubernetes deployment.

### Directory Structure

```
gha-demo-k8s-node/
├── server.js                     # Express server
├── .pipeline/                    # Piper configuration
│   └── config.yml
├── helm/                         # Helm charts
│   └── k8s-node-app/
│       ├── Chart.yaml
│       ├── values.yaml
│       └── templates/
├── tests/                        # Test suite
│   ├── karma/                    # Unit tests
│   └── gauge/                    # E2E tests
├── Dockerfile                    # Container image
├── package.json                  # npm configuration
├── postman_collection.json       # API tests
└── README.md
```

### Key Features

- **Full-stack example**: Complete application with tests
- **Kubernetes deployment**: Helm chart configuration
- **Multiple test types**: Unit, integration, E2E, API tests
- **Containerization**: Docker build configuration
- **Pipeline configuration**: Complete Piper setup

---

## Repository-Wide Features

### Architecture Decision Records (ADRs)

Multiple projects maintain ADR documentation:
- `jenkins-library/documentation/docs/adr/`
- `piper-azure-task/docs/adr/`
- `piper-pipeline-github/docs/adr/`
- `project-piper-action/docs/adr/`

All use Log4brains for ADR management.

### Configuration Files

Each project contains:
- `.editorconfig` - Editor configuration
- `.yamllint.yml` - YAML linting rules
- `.log4brains.yml` - ADR management
- `.github/workflows/` - CI/CD workflows
- `README.md` - Project documentation

### Build Systems

- **Go projects**: Makefiles, go.mod
- **Node.js projects**: package.json, npm scripts
- **Azure projects**: Azure Pipeline YAML
- **GitHub projects**: GitHub Actions YAML

---

## How Components Interact

```
User Project
    ↓
┌───────────────────────────────────────┐
│  Platform (Jenkins/GitHub/Azure)      │
│  Uses: piper-pipeline-* templates     │
└───────────────────────────────────────┘
    ↓
┌───────────────────────────────────────┐
│  Platform Adapter                      │
│  • Jenkins: vars/*.groovy             │
│  • GitHub: project-piper-action       │
│  • Azure: piper-azure-task            │
└───────────────────────────────────────┘
    ↓
┌───────────────────────────────────────┐
│  Piper Binary (jenkins-library/cmd/)  │
│  Reads: resources/defaults/*.yml      │
└───────────────────────────────────────┘
    ↓
┌───────────────────────────────────────┐
│  External Services                     │
│  (Build tools, scanners, clouds)      │
└───────────────────────────────────────┘
```

---

## Technology Stack Summary

| Component | Primary Language | Key Technologies |
|-----------|-----------------|------------------|
| jenkins-library | Go, Groovy | Cobra, Docker, Kubernetes |
| piper-pipeline-github | YAML | GitHub Actions |
| piper-pipeline-azure | YAML | Azure DevOps |
| project-piper-action | TypeScript | Node.js, GitHub Actions API |
| piper-azure-task | JavaScript | Azure DevOps SDK |
| resources | YAML | Node.js (build scripts) |
| project-piper.github.tools.sap | Markdown | MkDocs Material |
| piper-simple-go | Go | gorilla/mux |
| gha-demo-k8s-node | JavaScript | Express, Helm, Karma, Gauge |

---

## Summary

The Project Piper monorepo demonstrates:

- **Modularity**: Clear separation of concerns across components
- **Multi-platform**: Support for three major CI/CD platforms
- **Reusability**: Shared core binary and configuration
- **Documentation**: Comprehensive docs and examples
- **Extensibility**: Multiple extension points and customization options
- **Best Practices**: ADRs, linting, testing, CI/CD for itself

This structure enables teams to adopt Piper incrementally, starting with the platform they use (Jenkins, GitHub, or Azure) while benefiting from the shared core functionality.
