# GPP Test Example Project

This is an example npm project configured with SAP Piper for testing the GPP (GitHub Pipeline Platform) locally using act.

## Project Structure

```
.
├── .github/workflows/
│   └── piper.yml           # GitHub Actions workflow using Piper
├── .pipeline/
│   └── config.yml          # Piper pipeline configuration
├── package.json            # npm project configuration
├── index.js                # Simple application code
└── README.md              # This file
```

## Pipeline Configuration

The project is configured to run a simplified Piper pipeline with:

- **Build Stage**:
  - Version preparation (`artifactPrepareVersion`)
  - npm install and build (`npmExecuteScripts`)
  - Pipeline environment export

- **Disabled Stages**: Integration, Acceptance, Performance, Promote, Release (for local testing)

## Local Testing

To test this project with the GPP local testing tool:

```bash
cd ../
./gpp-test.sh
```

## Configuration

Key configurations in `.pipeline/config.yml`:

- `buildTool: npm` - Uses npm as the build tool
- `versioningType: cloud_noTag` - Cloud versioning without git tags
- Disabled stages for simplified local testing
- Mock service URLs pointing to local mock server

## Notes

- This is a minimal example for testing GPP locally
- External services (Vault, Cumulus, etc.) are mocked
- Security scans and deployments are disabled
- All URLs point to `host.docker.internal:8888` (mock server)
