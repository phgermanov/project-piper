# ABAP Development

Comprehensive guide for using Jenkins library ABAP development features on SAP BTP ABAP Environment and traditional ABAP systems.

## Overview

The Jenkins library provides extensive support for ABAP development workflows, including:
- **ABAP Environment Operations** - Manage SAP BTP ABAP Environment systems and repositories
- **Add-on Assembly** - Build and publish ABAP add-on products
- **Testing & Quality** - Run ATC checks and AUnit tests
- **gCTS Integration** - Git-enabled Change and Transport System operations

These steps enable continuous integration and delivery pipelines for ABAP development on both cloud and on-premise systems.

## ABAP Environment Steps

ABAP Environment steps provide comprehensive system and repository management capabilities for SAP BTP ABAP Environment.

### System Management

#### abapEnvironmentCreateSystem
Creates a new SAP BTP ABAP Environment system instance via Cloud Foundry.

**Key Parameters:**
- `cfServiceInstance` - Name of the service instance to create
- `cfServicePlan` - Service plan (standard, saas_oem)
- `abapSystemIsDevelopmentAllowed` - Enable/disable development (true/false)
- `abapSystemSizeOfPersistence` - Persistence size in GB
- `abapSystemSizeOfRuntime` - Runtime memory in GB

**Configuration Example:**
```yaml
steps:
  abapEnvironmentCreateSystem:
    cfApiEndpoint: 'https://api.cf.region.hana.ondemand.com'
    cfOrg: 'myOrg'
    cfSpace: 'mySpace'
    cfServiceInstance: 'abap-system-dev'
    cfServicePlan: 'standard'
    cfCredentialsId: 'cf-credentials'
    abapSystemIsDevelopmentAllowed: 'true'
```

### Repository Management

#### abapEnvironmentCloneGitRepo
Clones a git repository into an ABAP software component.

**Configuration Example:**
```yaml
steps:
  abapEnvironmentCloneGitRepo:
    abapCredentialsId: 'abap-credentials'
    host: 'https://abap-system.abap.region.hana.ondemand.com'
    repositories: 'repositories.yml'
```

#### abapEnvironmentPullGitRepo
Pulls the latest changes from a git repository to an ABAP system. Requires `repositoryName` and optional `commitID`.

#### abapEnvironmentCheckoutBranch
Checks out a specific branch of a software component. Requires `repositoryName` and `branchName`.

#### abapEnvironmentCreateTag
Creates a git tag in the ABAP system for a software component. Requires `repositoryName`, `tag`, and `commitID`.

### Testing & Quality Assurance

#### abapEnvironmentRunATCCheck
Executes ABAP Test Cockpit (ATC) checks on packages or software components.

**Configuration Example:**
```yaml
steps:
  abapEnvironmentRunATCCheck:
    abapCredentialsId: 'abap-credentials'
    host: 'https://abap-system.abap.region.hana.ondemand.com'
    atcConfig: 'atcconfig.yml'
```

**ATC Config File (atcconfig.yml):**
```yaml
checkvariant: "ABAP_CLOUD_DEVELOPMENT_DEFAULT"
objectset:
  softwarecomponents:
    - name: /NAMESPACE/COMPONENT
  packages:
    - name: /NAMESPACE/PACKAGE
```

#### abapEnvironmentRunAUnitTest
Executes ABAP Unit tests on the system. Requires `aUnitConfig` file with test scope configuration.

#### abapEnvironmentPushATCSystemConfig
Uploads ATC system configuration to the ABAP environment system. Requires `atcSystemConfigFilePath` parameter.

### Build Operations

#### abapEnvironmentBuild
Triggers a build operation on the ABAP system with configurable phases.

**Configuration Example:**
```yaml
steps:
  abapEnvironmentBuild:
    phase: 'BUILD'
    values: '[{"value_id":"ID1","value":"Value1"}]'
    maxRuntimeInMinutes: 360
    treatWarningsAsError: true
```

#### abapEnvironmentAssemblePackages
Assembles delivery packages for software component versions. Requires ABAP system credentials and host.

#### abapEnvironmentAssembleConfirm
Confirms successful assembly of packages. Called after abapEnvironmentAssemblePackages completes.

## Add-on Assembly Kit Steps

Add-on Assembly Kit as a Service (AAKaaS) steps orchestrate the build and publication of ABAP add-on products.

### Validation Steps

#### abapAddonAssemblyKitCheck
Performs initial checks for add-on product assembly.

#### abapAddonAssemblyKitCheckPV
Checks the product version validity in AAKaaS.

**Configuration Example:**
```yaml
steps:
  abapAddonAssemblyKitCheckPV:
    abapAddonAssemblyKitCredentialsId: 'aakaaS-credentials'
    addonDescriptorFileName: 'addon.yml'
```

#### abapAddonAssemblyKitCheckCVs
Validates software component versions defined in addon descriptor.

**Add-on Descriptor (addon.yml):**
```yaml
addonProduct: "/NAMESPACE/PRODUCTX"
addonVersion: "1.2.0"
repositories:
  - name: "/NAMESPACE/COMPONENTA"
    branch: "v1.2.0"
    version: "1.2.0"
    commitID: "7d4516e9"
    languages:
      - DE
      - EN
```

### Package Management

#### abapAddonAssemblyKitReserveNextPackages
Reserves the next package numbers for software component versions. Requires AAKaaS credentials.

#### abapAddonAssemblyKitRegisterPackages
Registers the assembled packages in AAKaaS. Requires AAKaaS credentials and addon descriptor.

#### abapAddonAssemblyKitReleasePackages
Releases registered packages for deployment. Final step before target vector creation.

### Target Vector Management

#### abapAddonAssemblyKitCreateTargetVector
Creates a target vector defining deliverable content. Use `targetVectorScope: 'T'` for test, `'P'` for production.

#### abapAddonAssemblyKitPublishTargetVector
Publishes target vector to make add-on version available for installation. Set scope to 'P' for production.

## Landscape Portal Integration

### abapLandscapePortalUpdateAddOnProduct
Updates add-on product information in the Landscape Portal. Requires `landscapePortalAPICredentialsId`.

## gCTS Steps

Git-enabled Change and Transport System integration for traditional ABAP systems.

### gctsExecuteABAPQualityChecks
Executes ATC checks via gCTS on traditional ABAP systems. Requires `host`, `client`, `repository`, and `credentialsId`.

### gctsExecuteABAPUnitTests
Executes ABAP Unit tests via gCTS. Requires same parameters as gctsExecuteABAPQualityChecks.

## ABAP Environment Pipeline

The ABAP Environment Pipeline provides a pre-configured CI/CD pipeline for ABAP development workflows.

### Pipeline Stages

1. **Init** - Initialize pipeline execution
2. **Initial Checks** - Validate addon descriptor and component versions
3. **Prepare System** - Create ABAP system (optional for transient systems)
4. **Clone Repositories** - Clone software components from git
5. **ATC** - Run ABAP Test Cockpit checks
6. **AUnit** - Execute ABAP Unit tests
7. **Build** - Assemble and register delivery packages
8. **Integration Tests** - Install and test add-on on test system
9. **Confirm** - Manual confirmation gate
10. **Publish** - Publish target vector to production
11. **Post** - Cleanup resources

### Pipeline Invocation

**Jenkinsfile:**
```groovy
@Library('piper-lib-os') _

abapEnvironmentPipeline script: this
```

### Pipeline Configuration

**config.yml:**
```yaml
general:
  cfApiEndpoint: 'https://api.cf.region.hana.ondemand.com'
  cfOrg: 'myOrg'
  cfSpace: 'mySpace'
  cfCredentialsId: 'cf-credentials'

stages:
  'Prepare System':
    cfServicePlan: 'standard'
    abapSystemSizeOfPersistence: 4
    abapSystemSizeOfRuntime: 2
    abapSystemIsDevelopmentAllowed: 'false'

  'Clone Repositories':
    repositories: 'repositories.yml'

  'ATC':
    atcConfig: 'atcconfig.yml'

  'Build':
    addonDescriptorFileName: 'addon.yml'

  'Integration Tests':
    cfServicePlan: 'saas_oem'

  'Publish':
    targetVectorScope: 'P'
```

**repositories.yml:**
```yaml
repositories:
  - name: '/NAMESPACE/COMPONENTA'
    branch: 'main'
  - name: '/NAMESPACE/COMPONENTB'
    branch: 'main'
```

## Complete Pipeline Examples

### Example 1: Continuous Testing Pipeline

For continuous testing on a permanent ABAP system without add-on build:

**Jenkinsfile:**
```groovy
@Library('piper-lib-os') _

abapEnvironmentPipeline script: this,
  stagesToActivate: ['Init', 'Clone Repositories', 'ATC', 'AUnit', 'Post']
```

**config.yml:**
```yaml
general:
  cfApiEndpoint: 'https://api.cf.region.hana.ondemand.com'
  cfOrg: 'myOrg'
  cfSpace: 'mySpace'
  cfCredentialsId: 'cf-credentials'
  cfServiceInstance: 'permanent-abap-system'

stages:
  'Clone Repositories':
    cfServiceKeyName: 'sap_com_0948'
    repositories: 'repositories.yml'

  'ATC':
    cfServiceKeyName: 'sap_com_0901'
    atcConfig: 'atcconfig.yml'

  'AUnit':
    cfServiceKeyName: 'sap_com_0735'
    aUnitConfig: 'aunitconfig.yml'
```

### Example 2: Add-on Build Pipeline

For building and publishing ABAP add-ons:

**Jenkinsfile:**
```groovy
@Library('piper-lib-os') _

abapEnvironmentPipeline script: this
```

**config.yml:**
```yaml
general:
  cfApiEndpoint: 'https://api.cf.region.hana.ondemand.com'
  cfOrg: 'myOrg'
  cfSpace: 'mySpace'
  cfCredentialsId: 'cf-credentials'
  cfServiceInstance: 'abap-assembly-system'

stages:
  'Initial Checks':
    abapAddonAssemblyKitCredentialsId: 'aakaaS-credentials'

  'Clone Repositories':
    repositories: 'repositories.yml'

  'ATC':
    atcConfig: 'atcconfig.yml'

  'Build':
    addonDescriptorFileName: 'addon.yml'
    abapAddonAssemblyKitCredentialsId: 'aakaaS-credentials'

  'Integration Tests':
    cfServicePlan: 'saas_oem'
    confirmDeletion: 'true'

  'Publish':
    targetVectorScope: 'P'
```

## Best Practices

### 1. System Management
- **Use transient systems** for add-on builds to ensure clean state
- **Set `abapSystemIsDevelopmentAllowed: false`** for assembly/test systems
- **Size systems appropriately** - minimum 4 GB persistence, 2 GB runtime for assembly
- **Monitor entitlements** - ensure sufficient ABAP compute units

### 2. Repository Management
- **Use commit IDs** for reproducible builds in addon.yml
- **Configure software component dependencies** in ABAP system
- **Follow branching strategy** - use version-based branches (v1.2.0)
- **Avoid parallel operations** on same software component

### 3. Quality Assurance
- **Block errors and warnings** - Set ATC quality gates to fail on findings
- **Use custom check variants** to reduce irrelevant findings
- **Include all components** in atcConfig.yml that are in addon.yml
- **Resolve findings early** - during transport release, not at build time

### 4. Add-on Versioning
- **Follow semantic versioning** - Release.SupportPackage.Patch (1.2.0)
- **Increment continuously** - No gaps allowed (1.0.0 → 2.0.0 → 2.1.0)
- **Align addon and component versions** - Leading component drives addon version
- **Support package limit** - Maximum 369 support packages per release

### 5. Build Process
- **Create API snapshots** for permanent assembly systems
- **Check platform version** - Minimum version determined by assembly system
- **Validate descriptor early** - Use Initial Checks stage
- **Archive build logs** - Available in abapEnvironmentAssemblePackages artifacts

### 6. Testing & Deployment
- **Test before production** - Use test scope target vectors first
- **Automate installation tests** - Integration Tests stage validates deployability
- **Register products** - Complete PPMS registration before first installation
- **Document languages** - Cannot change languages after first delivery

### 7. Pipeline Configuration
- **Store credentials securely** - Use Jenkins Credential Store
- **Externalize configuration** - Use config.yml instead of Jenkinsfile
- **Enable stage extensions** - Customize pipeline behavior when needed
- **Use service keys** - Simplify communication arrangement setup

### 8. Performance Optimization
- **Minimize ATC scope** - Test only changed components when possible
- **Use permanent systems** for frequent test runs
- **Parallel stage execution** - ATC and AUnit run concurrently
- **Configure timeouts** - Set appropriate maxRuntimeInMinutes

### 9. Error Handling
- **Idempotent operations** - Steps can be restarted without duplication
- **Check intermediate results** - Pipeline stores state between restarts
- **Fix forward** - Create new versions to fix errors after package registration
- **Monitor pipeline logs** - Use Timestamper plugin for debugging

### 10. Compliance & Security
- **Namespace reservation** - Reserve namespace before development
- **Technical communication users** - For AAKaaS and Cloud Foundry
- **Access control** - Limit who can trigger production deployments
- **Audit trail** - Pipeline execution provides complete deployment history

## Additional Resources

- [SAP Note 3032800](https://launchpad.support.sap.com/#/notes/3032800) - Recommended Piper library versions
- [Scenario: Build and Publish Add-ons](https://www.project-piper.io/scenarios/abapEnvironmentAddons/)
- [Scenario: Continuous Testing](https://www.project-piper.io/scenarios/abapEnvironmentTest/)
- [ABAP Environment Pipeline Documentation](https://www.project-piper.io/pipelines/abapEnvironment/introduction/)
- [SAP BTP ABAP Environment Help](https://help.sap.com/viewer/65de2977205c403bbc107264b8eccf4b/Cloud/en-US/11d77aa154f64c2da7d5db68f9cb9027.html)
- [Add-on Assembly Kit](https://help.sap.com/viewer/product/SAP_ADD-ON_ASSEMBLY_KIT/)

## Support Components

For issues, open support incidents on these components:

- **BC-CP-ABA** - ABAP Environment infrastructure, system provisioning
- **BC-CP-ABA-SC** - Software component management
- **BC-UPG-OCS** - Add-on Assembly Kit as a Service (AAKaaS)
- **BC-UPG-ADDON** - Add-on build, package assembly
- **BC-DWB-TOO-ATF** - ATC checks, AUnit tests
- **PROJ-PIPER** - Jenkins library issues (GitHub)
