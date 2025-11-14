# Project Piper Documentation

Welcome to the comprehensive documentation for **Project Piper**, SAP's enterprise-grade CI/CD tooling ecosystem.

## What is Project Piper?

Project Piper is a comprehensive CI/CD solution designed for the SAP ecosystem that provides reusable pipeline steps, stages, and complete pipeline templates across multiple CI/CD platforms. It enables teams to build, test, secure, and deploy SAP applications with standardized, best-practice workflows.

## Documentation Structure

This documentation is organized into the following sections:

### 1. [Overview](./overview/)
- [Architecture Overview](./overview/architecture.md) - System architecture and design patterns
- [Project Structure](./overview/project-structure.md) - Monorepo organization and components
- [How It Works](./overview/how-it-works.md) - Execution flow and component interactions

### 2. [Jenkins Library (piper-os)](./jenkins-library/)
Core Groovy shared library and Go CLI implementation:
- [Overview](./jenkins-library/00-overview.md)
- [Build Tools](./jenkins-library/01-build-tools.md)
- [Security Scanning](./jenkins-library/02-security-scanning.md)
- [Testing Frameworks](./jenkins-library/03-testing-frameworks.md)
- [Deployment](./jenkins-library/04-deployment.md)
- [SAP Integration](./jenkins-library/05-sap-integration.md)
- [ABAP Development](./jenkins-library/06-abap-development.md)
- [Container Operations](./jenkins-library/07-container-operations.md)
- [Version Control](./jenkins-library/08-version-control.md)
- [Utilities](./jenkins-library/09-utilities.md)

### 3. [GitHub Pipeline (GPP)](./github-pipeline/)
Piper's General Purpose Pipeline for GitHub Actions:
- [Overview](./github-pipeline/00-overview.md)
- [Init Stage](./github-pipeline/01-init-stage.md)
- [Build Stage](./github-pipeline/02-build-stage.md)
- [Integration Stage](./github-pipeline/03-integration-stage.md)
- [Acceptance Stage](./github-pipeline/04-acceptance-stage.md)
- [Performance Stage](./github-pipeline/05-performance-stage.md)
- [Promote Stage](./github-pipeline/06-promote-stage.md)
- [Release Stage](./github-pipeline/07-release-stage.md)
- [Post Stage](./github-pipeline/08-post-stage.md)
- [OSS and PPMS Stages](./github-pipeline/09-oss-ppms-stages.md)

### 4. [Azure DevOps Integration](./azure-devops/)
- [Overview](./azure-devops/00-overview.md)
- [Azure Task](./azure-devops/01-azure-task.md)
- [Pipeline Templates](./azure-devops/02-pipeline-templates.md)

### 5. [GitHub Action](./github-action/)
- [Overview](./github-action/00-overview.md)
- [Features](./github-action/01-features.md)
- [Usage Guide](./github-action/02-usage-guide.md)

### 6. [Configuration](./configuration/)
- [Overview](./configuration/00-overview.md)
- [Configuration Hierarchy](./configuration/01-configuration-hierarchy.md)
- [Default Settings](./configuration/02-default-settings.md)
- [Platform Deviations](./configuration/03-platform-deviations.md)
- [Stage Configuration](./configuration/04-stage-configuration.md)
- [Step Configuration](./configuration/05-step-configuration.md)
- [Credentials Management](./configuration/06-credentials-management.md)

### 7. [Integration Guides](./guides/)
- [Getting Started with Jenkins](./guides/jenkins-setup.md)
- [Getting Started with GitHub Actions](./guides/github-setup.md)
- [Getting Started with Azure DevOps](./guides/azure-setup.md)
- [Extensibility Guide](./guides/extensibility.md)
- [Migration Guide](./guides/migration.md)

### 8. [Resources](./resources/)
- [Step Reference](./resources/step-reference.md)
- [FAQ](./resources/faq.md)
- [Troubleshooting](./resources/troubleshooting.md)
- [Glossary](./resources/glossary.md)

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
