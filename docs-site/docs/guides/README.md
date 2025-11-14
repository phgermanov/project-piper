# Project Piper Integration Guides

Comprehensive guides for setting up and using Project Piper across different CI/CD platforms.

## Available Guides

### Setup Guides

1. **[jenkins-setup.md](./jenkins-setup.md)** - Complete Jenkins setup guide
   - Jenkins shared library configuration
   - Docker and Vault setup
   - Common patterns and examples
   - Troubleshooting

2. **[github-setup.md](./github-setup.md)** - Complete GitHub Actions setup guide
   - GitHub Actions workflow configuration
   - Secrets management
   - Reusable workflows
   - Matrix builds and parallel execution

3. **[azure-setup.md](./azure-setup.md)** - Complete Azure DevOps setup guide
   - Azure DevOps task installation
   - Service connections and variable groups
   - Template-based pipelines
   - Agent configuration

### Advanced Guides

4. **[extensibility.md](./extensibility.md)** - Pipeline extensibility and customization
   - Stage extensions (Jenkins)
   - Step hooks (all platforms)
   - Custom steps creation
   - Platform-specific extensions

5. **[migration.md](./migration.md)** - Migration between platforms
   - Jenkins to GitHub Actions
   - Jenkins to Azure DevOps
   - GitHub Actions to Azure DevOps
   - Azure DevOps to GitHub Actions

## Quick Start

Choose your platform and follow the corresponding setup guide:

- **Jenkins users**: Start with `jenkins-setup.md`
- **GitHub Actions users**: Start with `github-setup.md`
- **Azure DevOps users**: Start with `azure-setup.md`

For advanced customization, refer to `extensibility.md`.

For migrating between platforms, see `migration.md`.

## Common Topics

All guides include:
- Prerequisites
- Step-by-step installation instructions
- Configuration examples
- Common patterns
- Troubleshooting tips
- Best practices

## Additional Resources

- [Project Piper Documentation](https://www.project-piper.io/)
- [Piper Steps Reference](https://www.project-piper.io/steps/)
- [GitHub Repository](https://github.com/SAP/jenkins-library)
