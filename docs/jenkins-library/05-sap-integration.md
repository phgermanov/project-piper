# Jenkins Library - SAP Integration

This document covers SAP-specific integration capabilities available in jenkins-library (excluding ABAP development, which is covered separately).

## Overview

jenkins-library provides comprehensive support for SAP Business Technology Platform (BTP) services and SAP backend systems. These steps enable automation of integration flows, API management, source code deployment, and transport management.

**SAP Integration Categories**:
- **Integration Suite**: Manage Cloud Integration artifacts (10 steps)
- **API Management**: Handle API proxies, providers, and key-value maps (8 steps)
- **gCTS**: Git-enabled Change and Transport System for ABAP (6 steps)
- **Transport Management**: Traditional SAP transports and TMS (10 steps)
- **Change Management**: Solution Manager change document validation (2 steps)

---

## Integration Suite

### Overview

SAP Integration Suite (formerly Cloud Platform Integration) enables enterprise-grade integration between cloud and on-premise systems. These steps manage integration artifacts through the OData API.

**Service**: SAP Integration Suite - Cloud Integration
**Authentication**: Service key with plan 'api'
**Documentation**: [SAP Help - Cloud Integration](https://help.sap.com/viewer/368c481cd6954bdfa5d0435479fd4eaf/Cloud/en-US)

### Available Steps

| Step | Purpose |
|------|---------|
| `integrationArtifactDeploy` | Deploy integration flow to runtime |
| `integrationArtifactUnDeploy` | Undeploy integration flow from runtime |
| `integrationArtifactUpload` | Upload/create integration artifact |
| `integrationArtifactDownload` | Download integration artifact |
| `integrationArtifactUpdateConfiguration` | Update artifact parameters |
| `integrationArtifactGetMplStatus` | Get deployment status |
| `integrationArtifactGetServiceEndpoint` | Get runtime endpoint URL |
| `integrationArtifactTriggerIntegrationTest` | Execute integration test |
| `integrationArtifactTransport` | Transport artifact between tenants |
| `integrationArtifactResource` | Manage artifact resources |

### Configuration Example

```yaml
# .pipeline/config.yml
steps:
  integrationArtifactDeploy:
    cpiApiServiceKeyCredentialsId: 'cpi-service-key'
    integrationFlowId: 'MyIntegrationFlow'

  integrationArtifactUpload:
    cpiApiServiceKeyCredentialsId: 'cpi-service-key'
    integrationFlowId: 'MyIntegrationFlow'
    filePath: 'integration-content/MyFlow.zip'
    integrationFlowName: 'My Integration Flow'
    integrationFlowVersion: '1.0.0'

  integrationArtifactUpdateConfiguration:
    cpiApiServiceKeyCredentialsId: 'cpi-service-key'
    integrationFlowId: 'MyFlow'
    parameterKey: 'apiEndpoint'
    parameterValue: 'https://api.production.example.com'
```

### Common Workflow

```groovy
stage('CPI Deployment') {
    steps {
        script {
            integrationArtifactUpload script: this, integrationFlowId: 'MyFlow', filePath: 'src/MyFlow.zip'
            integrationArtifactDeploy script: this, integrationFlowId: 'MyFlow'
            integrationArtifactTriggerIntegrationTest script: this, integrationFlowId: 'MyFlow'
        }
    }
}
```

### Best Practices

- Store service keys in Jenkins credentials, never in code
- Keep integration flow artifacts in Git repositories
- Use `integrationArtifactUpdateConfiguration` for environment-specific settings
- Always verify deployment status with `integrationArtifactGetMplStatus`
- Integrate automated testing in CI/CD pipelines
- Document integration flow IDs and dependencies

---

## API Management

### Overview

SAP API Management provides comprehensive API lifecycle management. These steps manage API proxies, API providers, and key-value maps through the OData API.

**Service**: SAP API Management
**Authentication**: Service key with plan 'api'
**Documentation**: [SAP Help - API Management](https://help.sap.com/viewer/66d066d903c2473f81ec33acfe2ccdb4/Cloud/en-US)

### Available Steps

| Step | Purpose |
|------|---------|
| `apiProxyUpload` | Upload API proxy artifact |
| `apiProxyDownload` | Download API proxy artifact |
| `apiProxyList` | List all API proxies |
| `apiProviderUpload` | Upload API provider configuration |
| `apiProviderDownload` | Download API provider configuration |
| `apiProviderList` | List all API providers |
| `apiKeyValueMapUpload` | Upload key-value map for configuration |
| `apiKeyValueMapDownload` | Download key-value map |

### Configuration Example

```yaml
# .pipeline/config.yml
steps:
  apiProxyUpload:
    apimApiServiceKeyCredentialsId: 'apim-service-key'
    filePath: 'apiproxy/MyAPI.zip'

  apiProviderUpload:
    apimApiServiceKeyCredentialsId: 'apim-service-key'
    filePath: 'providers/MyProvider.json'

  apiKeyValueMapUpload:
    apimApiServiceKeyCredentialsId: 'apim-service-key'
    filePath: 'config/dev-config.json'
    keyValueMapName: 'DevEnvironmentConfig'
```

### Common Workflow

```groovy
stage('API Deployment') {
    steps {
        sh 'zip -r MyAPI.zip apiproxy/'
        script {
            apiProxyUpload script: this, filePath: 'MyAPI.zip'
            apiKeyValueMapUpload script: this, filePath: 'config/prod-config.json', keyValueMapName: 'ProdConfig'
        }
    }
}
```

### Best Practices

- Include version numbers in API proxy names
- Use key-value maps for environment-specific settings
- Keep API provider definitions in source control
- Download APIs for backup before updates
- Use list operations for inventory auditing
- Test API policies locally before uploading

---

## gCTS (Git-enabled Change and Transport System)

### Overview

gCTS enables Git-based version control for ABAP development, allowing direct deployment from Git repositories to ABAP systems.

**Requirements**: SAP S/4HANA 2020 or later
**Authentication**: Username/password credentials
**Documentation**: [SAP Help - gCTS](https://help.sap.com/docs/ABAP_PLATFORM_NEW/4a368c163b08418890a406d413933ba7)

### Available Steps

| Step | Purpose |
|------|---------|
| `gctsDeploy` | Deploy Git repository to ABAP system |
| `gctsCreateRepository` | Create local gCTS repository |
| `gctsCloneRepository` | Clone remote Git repository |
| `gctsRollback` | Rollback to previous commit |
| `gctsExecuteABAPQualityChecks` | Run ABAP Test Cockpit (ATC) checks |
| `gctsExecuteABAPUnitTests` | Execute ABAP unit tests |

### Configuration Example

```yaml
# .pipeline/config.yml
steps:
  gctsDeploy:
    abapCredentialsId: 'abap-credentials'
    host: 'https://s4hana.example.com:44300'
    client: '100'
    repository: 'my_repo'
    remoteRepositoryURL: 'https://github.com/myorg/abap-repo.git'
    role: 'SOURCE'
    branch: 'main'
    rollback: true  # Rollback on import errors

  gctsExecuteABAPQualityChecks:
    abapCredentialsId: 'abap-credentials'
    host: 'https://s4hana.example.com:44300'
    client: '100'
    repository: 'my_repo'
    atcConfig: 'default'
```

### Common Workflow

```groovy
stage('gCTS Deployment') {
    steps {
        script {
            gctsDeploy script: this, repository: 'ZDEV_REPO', branch: 'main'
            gctsExecuteABAPQualityChecks script: this, repository: 'ZDEV_REPO'
            gctsExecuteABAPUnitTests script: this, repository: 'ZDEV_REPO'
        }
    }
}
post {
    failure {
        script { gctsRollback script: this, repository: 'ZDEV_REPO' }
    }
}
```

### Best Practices

- Use clear repository names (e.g., `ZPROJECT_DEV`)
- Assign `SOURCE` role for development systems, `TARGET` for test/production
- Implement automatic rollback in failure handlers
- Always run ATC checks before promoting to higher environments
- Deploy specific commits for reproducible deployments
- Set appropriate vSID for transport route mapping
- Only disable SSL verification in non-productive environments

---

## Transport Management

### Transport Management Service (TMS)

**Service**: SAP Cloud Transport Management
**Platforms**: SAP BTP Cloud Foundry, Neo
**Documentation**: [SAP Help - TMS](https://help.sap.com/viewer/p/TRANSPORT_MANAGEMENT_SERVICE)

#### Available Steps

| Step | Purpose |
|------|---------|
| `tmsUpload` | Upload MTA file to TMS node |
| `tmsExport` | Export transport from TMS node |

#### Configuration Example

```yaml
# .pipeline/config.yml
steps:
  tmsUpload:
    credentialsId: 'tms-service-key'
    nodeName: 'DEV'
    mtaPath: 'myapp.mtar'
    customDescription: 'Release v1.2.3'
    nodeExtDescriptorMapping:
      QA: 'qa-extension.mtaext'
      PROD: 'prod-extension.mtaext'

  tmsExport:
    credentialsId: 'tms-service-key'
    nodeName: 'QA'
```

#### Common Workflow

```groovy
stage('TMS Upload') {
    steps {
        script {
            mtaBuild script: this
            tmsUpload script: this, nodeName: 'DEV', customDescription: "Build ${env.BUILD_NUMBER}"
        }
    }
}
stage('TMS Export to QA') {
    steps {
        input message: 'Deploy to QA?'
        script { tmsExport script: this, nodeName: 'QA' }
    }
}
```

### Traditional Transport Requests

**Backends**: Solution Manager (SOLMAN), CTS, RFC
**Use Cases**: Legacy SAP systems, on-premise deployments

#### Available Steps

| Step | Purpose |
|------|---------|
| `transportRequestCreate` | Create new transport request |
| `transportRequestRelease` | Release transport request |
| `transportRequestUploadFile` | Upload file to transport |
| `transportRequestUploadCTS` | Upload to CTS transport |
| `transportRequestUploadRFC` | Upload via RFC connection |
| `transportRequestUploadSOLMAN` | Upload to Solution Manager |
| `transportRequestDocIDFromGit` | Extract document ID from Git |
| `transportRequestReqIDFromGit` | Extract request ID from Git |

#### Configuration Example

```yaml
# .pipeline/config.yml
steps:
  transportRequestCreate:
    changeManagement:
      type: 'SOLMAN'  # Options: SOLMAN, CTS, RFC
      endpoint: 'https://solman.example.com'
      credentialsId: 'solman-credentials'
      changeDocumentLabel: 'ChangeDocument\\s?:'
      git:
        from: 'origin/master'
        to: 'HEAD'
        format: '%b'
    developmentSystemId: 'DEV~ABAP/100'

  transportRequestUploadFile:
    changeManagement:
      type: 'CTS'
      endpoint: 'https://cts.example.com'
      credentialsId: 'cts-credentials'
      client: '100'
    applicationName: 'MyFioriApp'
    abapPackage: 'Z_MY_PACKAGE'
    filePath: 'dist/myapp.zip'
```

#### Common Workflow: Traditional Transport

```groovy
stage('Transport') {
    steps {
        script {
            transportRequestCreate script: this, changeDocumentId: env.CHANGE_DOC_ID

            def transportRequestId = commonPipelineEnvironment.getTransportRequestId()

            transportRequestUploadFile script: this, transportRequestId: transportRequestId, applicationId: 'APP001'

            transportRequestRelease script: this, transportRequestId: transportRequestId
        }
    }
}
```

### Best Practices

- Store service keys securely in Jenkins credentials
- Set up transport routes in TMS landscape before automation
- Use `nodeExtDescriptorMapping` for environment-specific MTA configurations
- Provide meaningful descriptions with `customDescription`
- Extract change document IDs from Git commit messages
- Use manual approval steps before production transports
- Verify transport success before releasing

---

## Change Management

### Overview

Change management steps validate that changes are properly tracked in SAP Solution Manager before deployment.

**Backend**: SAP Solution Manager
**Status**: Deprecated - follow SAP documentation for alternatives

### Available Steps

| Step | Purpose |
|------|---------|
| `checkChangeInDevelopment` | Verify change document is in development status |
| `isChangeInDevelopment` | Check if change is in development |

### Configuration Example

```yaml
# .pipeline/config.yml
steps:
  checkChangeInDevelopment:
    changeManagement:
      endpoint: 'https://solman.example.com'
      credentialsId: 'solman-credentials'
      changeDocumentLabel: 'ChangeDocument\\s?:'
      git:
        from: 'origin/master'
        to: 'HEAD'
        format: '%b'
    failIfStatusIsNotInDevelopment: true
```

### Common Workflow

```groovy
stage('Validate Change') {
    steps {
        script {
            checkChangeInDevelopment script: this, failIfStatusIsNotInDevelopment: true
        }
    }
}
stage('Build and Deploy') {
    when {
        expression { return commonPipelineEnvironment.getValue('isChangeInDevelopment') }
    }
    steps {
        echo 'Proceeding with deployment'
    }
}
```

### Best Practices

- Include change document IDs in standardized commit message format
- Configure appropriate commit ranges for Git history scanning
- Decide whether to fail or warn on validation failures
- Plan migration to modern change management tools
- Document change document ID format for team members

---

## General Best Practices

### Credential Management
- Store all service keys and passwords in Jenkins credentials store
- Use descriptive credential IDs (e.g., `cpi-dev-service-key`)
- Rotate credentials regularly and update in Jenkins
- Never commit credentials to source control

### Configuration Management
- Use `.pipeline/config.yml` for shared configuration
- Store environment-specific values separately (use extension descriptors for MTA)
- Version control all configuration files
- Document all custom parameters and their purposes

### Error Handling
- Implement proper error handling and rollback mechanisms
- Use `post` blocks for cleanup and failure handling
- Log meaningful error messages for troubleshooting
- Set up notifications for deployment failures (Slack, email)

### Testing
- Test integration artifacts before deployment to production
- Implement smoke tests after deployment
- Use separate tenants/systems for development and testing
- Validate configuration changes in lower environments first

### Documentation
- Document service endpoints and credentials locations
- Maintain runbooks for common operations and troubleshooting
- Keep deployment procedures up to date
- Track known issues and workarounds in a central location

### Security
- Follow principle of least privilege for service accounts
- Regularly audit and rotate credentials
- Use separate credentials for different environments
- Enable audit logging for compliance requirements

---

## Next Steps

- [ABAP Development Documentation](./06-abap-development.md) - For ABAP-specific steps (30+ steps)
- [Deployment Documentation](./04-deployment.md) - For Cloud Foundry and other platform deployments
- [Overview](./00-overview.md) - jenkins-library overview and architecture

## Resources

- **SAP Integration Suite**: https://help.sap.com/viewer/product/INTEGRATION_SUITE
- **SAP API Management**: https://help.sap.com/viewer/product/SAP_CLOUD_PLATFORM_API_MANAGEMENT
- **gCTS Documentation**: https://help.sap.com/docs/ABAP_PLATFORM_NEW/4a368c163b08418890a406d413933ba7
- **Transport Management Service**: https://help.sap.com/viewer/p/TRANSPORT_MANAGEMENT_SERVICE
- **Project Piper**: https://www.project-piper.io/
- **Jenkins Library GitHub**: https://github.com/SAP/jenkins-library

---

*This documentation covers SAP integration capabilities in jenkins-library. For ABAP-specific development steps, see the ABAP Development documentation.*
