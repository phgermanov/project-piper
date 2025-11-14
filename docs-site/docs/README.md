# Project Piper Documentation

Welcome to the comprehensive documentation for **Project Piper**, SAP's enterprise-grade CI/CD tooling ecosystem.

## What is Project Piper?

Project Piper is a comprehensive CI/CD solution designed for the SAP ecosystem that provides reusable pipeline steps, stages, and complete pipeline templates across multiple CI/CD platforms. It enables teams to build, test, secure, and deploy SAP applications with standardized, best-practice workflows.

## Documentation Structure

This documentation is organized into the following sections:

### 1. [Overview](overview/architecture)
- [Architecture Overview](overview/architecture) - System architecture and design patterns
- [Project Structure](overview/project-structure) - Monorepo organization and components
- [How It Works](overview/how-it-works) - Execution flow and component interactions

### 2. [Jenkins Library (piper-os)](jenkins-library/overview)
Core Groovy shared library and Go CLI implementation:
- [Overview](jenkins-library/overview)
- [Build Tools](jenkins-library/build-tools)
- [Security Scanning](jenkins-library/security-scanning)
- [Testing Frameworks](jenkins-library/testing-frameworks)
- [Deployment](jenkins-library/deployment)
- [SAP Integration](jenkins-library/sap-integration)
- [ABAP Development](jenkins-library/abap-development)
- [Container Operations](jenkins-library/container-operations)
- [Version Control](jenkins-library/version-control)
- [Utilities](jenkins-library/utilities)

### 3. [GitHub Pipeline (GPP)](github-pipeline/overview)
Piper's General Purpose Pipeline for GitHub Actions:
- [Overview](github-pipeline/overview)
- [Init Stage](github-pipeline/init-stage)
- [Build Stage](github-pipeline/build-stage)
- [Integration Stage](github-pipeline/integration-stage)
- [Acceptance Stage](github-pipeline/acceptance-stage)
- [Performance Stage](github-pipeline/performance-stage)
- [Promote Stage](github-pipeline/promote-stage)
- [Release Stage](github-pipeline/release-stage)
- [Post Stage](github-pipeline/post-stage)
- [OSS and PPMS Stages](github-pipeline/oss-ppms-stages)

### 4. [Azure DevOps Integration](azure-devops/overview)
- [Overview](azure-devops/overview)
- [Azure Task](azure-devops/azure-task)
- [Pipeline Templates](azure-devops/pipeline-templates)

### 5. [GitHub Action](github-action/overview)
- [Overview](github-action/overview)
- [Features](github-action/features)
- [Usage Guide](github-action/usage-guide)

### 6. [Configuration](configuration/overview)
- [Overview](configuration/overview)
- [Configuration Hierarchy](configuration/configuration-hierarchy)
- [Default Settings](configuration/default-settings)
- [Platform Deviations](configuration/platform-deviations)
- [Stage Configuration](configuration/stage-configuration)
- [Step Configuration](configuration/step-configuration)
- [Credentials Management](configuration/credentials-management)

### 7. [Integration Guides](guides/jenkins-setup)
- [Getting Started with Jenkins](guides/jenkins-setup)
- [Getting Started with GitHub Actions](guides/github-setup)
- [Getting Started with Azure DevOps](guides/azure-setup)
- [Extensibility Guide](guides/extensibility)
- [Migration Guide](guides/migration)

### 8. [Resources](resources/step-reference)
- [Step Reference](resources/step-reference)
- [FAQ](resources/faq)
- [Troubleshooting](resources/troubleshooting)
- [Glossary](resources/glossary)

## Quick Start

### For Jenkins Users
```groovy
@Library('piper-lib-os') _
piperPipeline script: this
```

### For GitHub Actions Users
```yaml
- uses: SAP/project-piper-action@main
  with:
    step-name: mavenBuild
    flags: '--publish --createBOM'
```

### For Azure DevOps Users
```yaml
resources:
  repositories:
    - repository: piper
      type: github
      name: SAP/piper-pipeline-azure

stages:
  - template: sap-piper-pipeline.yml@piper
```

## Key Features

- **188+ Pipeline Steps**: Comprehensive library of reusable CI/CD steps
- **Multi-Platform Support**: Jenkins, GitHub Actions, and Azure DevOps
- **Security Scanning**: 12 integrated security scanning tools
- **Build Tool Support**: Maven, Gradle, npm, MTA, Docker, Go, Python, and more
- **SAP Integration**: Deep integration with SAP BTP, ABAP, and SAP services
- **Extensibility**: Multiple extension points for customization
- **Configuration Management**: Flexible, hierarchical configuration system
- **Enterprise Ready**: Vault integration, compliance features, and audit trails

## Technology Stack

- **Core Binary**: Go 1.24+
- **Jenkins Library**: Groovy shared library
- **GitHub Action**: TypeScript/Node.js
- **Azure Task**: Azure DevOps extension
- **Configuration**: YAML-based
- **Documentation**: MkDocs with Material theme

## Repository Structure

This is a monorepo containing 9 main components:

```
project-piper/
├── jenkins-library/           # Core Go CLI and Groovy shared library
├── piper-pipeline-github/     # GitHub Actions workflow templates
├── piper-pipeline-azure/      # Azure DevOps pipeline templates
├── project-piper-action/      # GitHub Action implementation
├── piper-azure-task/          # Azure DevOps task implementation
├── resources/                 # Shared configuration resources
├── project-piper.github.tools.sap/  # Documentation website
├── piper-simple-go/          # Demo Go project
└── gha-demo-k8s-node/        # Demo Node.js project
```

## Contributing

**Note**: The jenkins-library project is no longer accepting contributions but remains available for use.

For other components, please refer to the individual CONTRIBUTING.md files in each project directory.

## Support

- **Issues**: Report issues in the respective GitHub repositories
- **Documentation**: https://www.project-piper.io/
- **Examples**: See demo projects in `piper-simple-go/` and `gha-demo-k8s-node/`

## License

Please refer to individual LICENSE files in each project directory.

---

*This documentation was generated for Project Piper CI/CD Tool*
