# OSS and PPMS Stages

## Overview

The OSS (Open Source Security) and PPMS (Product Portfolio Management System) stages are specialized compliance stages that run separately from the main pipeline, typically on a scheduled basis. These stages perform open source scanning for security vulnerabilities and IP compliance, and validate PPMS documentation requirements.

## Purpose

These stages ensure compliance with SAP corporate requirements:

- **OSS/Security Stage**: Scans for open source security vulnerabilities
- **PPMS Stage**: Validates IP compliance and PPMS documentation
- Both stages support compliance requirements for production releases

## Stage Separation in GitHub Actions

Unlike Jenkins GPP where these are part of the main pipeline, in GitHub Actions:

- **Main Pipeline**: Build, Test, Deploy stages run on push/PR
- **OSS & PPMS Workflows**: Separate workflows run on schedule
- **Benefit**: Reduces main pipeline duration, runs comprehensive scans periodically

## OSS/Security Stage

### Purpose

The OSS stage (named "Security" in the workflow) performs:

- Open source vulnerability scanning
- Security analysis of dependencies
- SBOM (Software Bill of Materials) generation
- IP scanning for license compliance
- Results upload to Cumulus

### When It Runs

- Activated when `fromJSON(inputs.active-stages-map).Security == true`
- Typically scheduled (e.g., nightly, weekly)
- Can be triggered manually
- Runs independent of main pipeline

### Scanning Tools

The OSS stage supports three scanning tools:

#### 1. WhiteSource (Mend)

- **Step**: `whitesourceExecuteScan`
- **Purpose**: Security and IP scanning
- **Condition**: `whitesourceProductName` configured
- **Capabilities**:
  - Vulnerability detection
  - License compliance
  - Policy violations
  - SBOM generation

#### 2. BlackDuck (Detect/Synopsys)

- **Step**: `detectExecuteScan`
- **Purpose**: Security and IP scanning
- **Condition**: `projectName` configured
- **Capabilities**:
  - Comprehensive vulnerability scanning
  - License and IP compliance
  - Policy enforcement
  - Risk reporting

#### 3. Protecode (BlackDuck Binary Analysis)

- **Step**: `protecodeExecuteScan`
- **Purpose**: Binary analysis
- **Condition**: `protecodeGroup` configured
- **Capabilities**:
  - Container/binary scanning
  - Vulnerability detection
  - License analysis

### OSS Stage Steps

#### 1. Checkout repository
- **Action**: `actions/checkout@v4.3.0`
- **Purpose**: Checks out source code for scanning

#### 2. Checkout global extension
- **Condition**: When extensibility is enabled
- **Path**: `.pipeline/tmp/global_extensions`

#### 3. Retrieve System Trust session token
- **Action**: `project-piper/system-trust-composite-action`
- **Permissions Required**: `id-token: write`

#### 4. preOSS
- **Condition**: When extensibility is enabled
- **Purpose**: Custom pre-scan setup

#### 5. whitesourceExecuteScan
- **Action**: `SAP/project-piper-action@v1.22`
- **Condition**: WhiteSource configured
- **Purpose**: Executes WhiteSource scan

#### 6-7. sapCumulusUpload (WhiteSource results)
- **Security Results**: Risk report, vulnerability report, SARIF, SBOM
- **IP Results**: IP scan results, risk report
- **Policy Evidence**: PSL-1 policy evidence

#### 8. detectExecuteScan
- **Action**: `SAP/project-piper-action@v1.22`
- **Condition**: BlackDuck configured
- **Purpose**: Executes BlackDuck Detect scan

#### 9-11. sapCumulusUpload (BlackDuck results)
- **Security Results**: Risk report, policy results, SARIF, SBOM
- **IP Results**: IP scan results, policy violations
- **Policy Evidence**: PSL-1 policy evidence

#### 12. protecodeExecuteScan
- **Action**: `SAP/project-piper-action@v1.22`
- **Condition**: Protecode configured
- **Purpose**: Executes Protecode binary scan

#### 13. sapCumulusUpload (Protecode results)
- **Results**: Tool run results, vulnerability data

#### 14. postOSS
- **Condition**: When extensibility is enabled
- **Purpose**: Custom post-scan actions

#### 15. Export pipeline environment
- **Purpose**: Exports CPE
- **Output**: `pipelineEnv`

### OSS Configuration

Configure in `.pipeline/config.yml`:

```yaml
steps:
  # WhiteSource Configuration
  whitesourceExecuteScan:
    # Product name in WhiteSource
    productName: 'MyProduct'

    # Product version
    productVersion: '1.0'

    # Project name
    projectName: 'my-project'

    # Scan type
    scanType: 'npm'  # npm, maven, pip, etc.

    # Security scan
    securityVulnerabilities: true

    # IP scan
    licensingVulnerabilities: true

    # Generate SARIF
    createResultIssue: true

  # BlackDuck Configuration
  detectExecuteScan:
    # Project name
    projectName: 'my-project'

    # Version
    version: '1.0'

    # Scan paths
    scanPaths:
      - '.'

    # Scan type
    scanType: 'source'  # or 'docker'

    # Fail on vulnerabilities
    failOn:
      - 'BLOCKER'
      - 'CRITICAL'

  # Protecode Configuration
  protecodeExecuteScan:
    # Group ID
    group: '12345'

    # Timeout
    timeoutMinutes: 60

    # Fail on severities
    failOnSeverity: 'high'
```

## PPMS Stage

### Purpose

The PPMS stage validates compliance with SAP's Product Portfolio Management System:

- Checks PPMS object exists and is properly documented
- Validates IP scan results
- Verifies ECCN (Export Control Classification Number) classification
- Uploads compliance evidence to Cumulus

### When It Runs

- Activated when `fromJSON(inputs.active-stages-map)['IPScan and PPMS'] == true`
- Typically scheduled (e.g., weekly, monthly)
- Can be triggered manually
- Often runs together with OSS stage

### PPMS Stage Steps

#### 1. Checkout repository
- **Action**: `actions/checkout@v4.3.0`
- **Purpose**: Checks out source code

#### 2. Checkout global extension
- **Condition**: When extensibility is enabled

#### 3. Retrieve System Trust session token
- **Action**: `project-piper/system-trust-composite-action`

#### 4. prePPMS
- **Condition**: When extensibility is enabled
- **Purpose**: Custom pre-PPMS setup

#### 5. sapCheckPPMSCompliance
- **Action**: `SAP/project-piper-action@v1.22`
- **Step**: `sapCheckPPMSCompliance`
- **Condition**: `ppmsID` configured
- **Purpose**: Validates PPMS compliance
- **Checks**:
  - PPMS object exists
  - Required fields populated
  - Documentation complete
  - Approval status

#### 6-7. sapCumulusUpload (IP scan results)
- **WhiteSource**: PPMS report from WhiteSource
- **BlackDuck**: PPMS report from BlackDuck
- **Note**: Retrieves results from OSS stage scans

#### 8. postPPMS
- **Condition**: When extensibility is enabled
- **Purpose**: Custom post-PPMS actions

#### 9. Export pipeline environment
- **Output**: `pipelineEnv`

### PPMS Configuration

```yaml
steps:
  sapCheckPPMSCompliance:
    # PPMS ID
    ppmsID: 'PPMS-12345'

    # PPMS system URL (from Vault)
    ppmsUrl: 'https://ppms.example.com'

    # Fail on non-compliance
    failOnNonCompliance: true

    # Check fields
    checkFields:
      - 'Description'
      - 'TechnicalContact'
      - 'SecurityContact'

  sapCheckECCNCompliance:
    # ECCN credentials (from Vault)
    eccnCredentialsId: 'eccn-credentials'

    # Fail if not classified
    failOnMissing: true
```

## Separate Workflow Example

### OSS and PPMS Scheduled Workflow

Create `.github/workflows/oss-ppms-scan.yml`:

```yaml
name: OSS and PPMS Compliance Scan

on:
  schedule:
    # Run weekly on Sunday at 2 AM
    - cron: '0 2 * * 0'
  workflow_dispatch:  # Allow manual trigger

permissions:
  id-token: write
  contents: read

jobs:
  init:
    uses: project-piper/piper-pipeline-github/.github/workflows/init.yml@main
    with:
      piper-version: 'latest'

  oss:
    name: OSS Security Scan
    needs: init
    if: fromJSON(needs.init.outputs.active-stages-map).Security == true
    uses: project-piper/piper-pipeline-github/.github/workflows/oss.yml@main
    with:
      piper-version: 'latest'
      on-productive-branch: 'true'
      active-stages-map: ${{ needs.init.outputs.active-stages-map }}
      active-steps-map: ${{ needs.init.outputs.active-steps-map }}
      pipeline-env: ${{ needs.init.outputs.pipeline-env }}
    secrets: inherit

  ppms:
    name: PPMS Compliance Check
    needs: [init, oss]
    if: fromJSON(needs.init.outputs.active-stages-map)['IPScan and PPMS'] == true
    uses: project-piper/piper-pipeline-github/.github/workflows/ppms.yml@main
    with:
      piper-version: 'latest'
      on-productive-branch: 'true'
      active-stages-map: ${{ needs.init.outputs.active-stages-map }}
      active-steps-map: ${{ needs.init.outputs.active-steps-map }}
      pipeline-env: ${{ needs.oss.outputs.pipeline-env }}
    secrets: inherit

  post:
    needs: [init, oss, ppms]
    if: always()
    uses: project-piper/piper-pipeline-github/.github/workflows/post.yml@main
    with:
      on-productive-branch: 'true'
      active-stages-map: ${{ needs.init.outputs.active-stages-map }}
      active-steps-map: ${{ needs.init.outputs.active-steps-map }}
      pipeline-env: ${{ needs.ppms.outputs.pipeline-env || needs.oss.outputs.pipeline-env }}
    secrets: inherit
```

## Scan Result Upload Patterns

### Security Scan Results

```yaml
# Pattern: --filePattern {files} --stepResultType {type}

# WhiteSource Security
--filePattern whitesource/*risk-report.pdf,**/toolrun_whitesource_*.json,**/piper_whitesource_vulnerability_report.html,**/piper_whitesource_vulnerability.sarif,**/piper_whitesource_sbom.xml
--stepResultType whitesource-security

# BlackDuck Security
--filePattern **/*BlackDuck_RiskReport.pdf,**/detectExecuteScan_policy_*.json,**/piper_detect_vulnerability_report.html,**/toolrun_detectExecute_*.json,**/piper_detect_vulnerability.sarif,**/piper_hub_detect_sbom.xml
--stepResultType blackduck-security

# Protecode
--filePattern **/toolrun_protecode_*.json
--stepResultType protecode
```

### IP Scan Results

```yaml
# WhiteSource IP
--filePattern **/whitesource-ip.json,whitesource/*risk-report.pdf,**/toolrun_whitesource_*.json
--stepResultType whitesource-ip

# BlackDuck IP
--filePattern **/*BlackDuck_RiskReport.pdf,**/blackduck-ip.json,**/toolrun_detectExecute_*.json,**/piper_detect_policy_violation_report.html
--stepResultType blackduck-ip
```

### Policy Evidence

```yaml
# PSL-1 Policy (IP Compliance)
--filePattern **/whitesource-ip.json
--stepResultType policy-evidence/PSL-1

--filePattern **/blackduck-ip.json
--stepResultType policy-evidence/PSL-1
```

## Extension Examples

### preOSS Extension

Create `.pipeline/extensions/preOSS.sh`:

```bash
#!/bin/bash
set -e

echo "Preparing for OSS scan..."

# Build project to generate dependency files
if [ -f "pom.xml" ]; then
    mvn dependency:tree -DoutputFile=dependency-tree.txt
elif [ -f "package.json" ]; then
    npm install
    npm list --all > dependency-tree.txt
fi

# Generate SBOM
if command -v syft &> /dev/null; then
    syft . -o cyclonedx-json > sbom.json
fi

echo "OSS scan preparation complete"
```

### postOSS Extension

```bash
#!/bin/bash
set -e

echo "Processing OSS scan results..."

# Check for critical vulnerabilities
if [ -f "piper_whitesource_vulnerability_report.html" ]; then
    CRITICAL_COUNT=$(grep -c "CRITICAL" piper_whitesource_vulnerability_report.html || true)

    if [ "$CRITICAL_COUNT" -gt 0 ]; then
        echo "⚠️ Found ${CRITICAL_COUNT} critical vulnerabilities"
        # Send alert
        curl -X POST "${SLACK_WEBHOOK}" \
            -d "{\"text\": \"Critical vulnerabilities found in ${GITHUB_REPOSITORY}\"}"
    fi
fi

# Generate summary
cat > oss-scan-summary.txt <<EOF
OSS Scan Summary
================
Repository: ${GITHUB_REPOSITORY}
Scan Date: $(date)
Critical Vulnerabilities: ${CRITICAL_COUNT}
EOF

echo "OSS scan results processed"
```

### prePPMS Extension

```bash
#!/bin/bash
set -e

echo "Preparing PPMS compliance check..."

# Verify PPMS ID is set
if [ -z "${PPMS_ID}" ]; then
    echo "Error: PPMS_ID not set"
    exit 1
fi

# Download latest IP scan results
# (from OSS stage if running in same workflow)

echo "PPMS preparation complete"
```

## Troubleshooting

### OSS Scan Issues

1. **Scan timeout**:
   - Increase timeout configuration
   - Split large projects into modules
   - Use incremental scanning

2. **Authentication failures**:
   - Verify credentials in Vault
   - Check API tokens validity
   - Validate server URLs

3. **High false positive rate**:
   - Configure suppression rules
   - Use ignore files (.wssignore, .bdignore)
   - Update scan policies

### PPMS Issues

1. **PPMS object not found**:
   - Verify PPMS ID is correct
   - Check PPMS system connectivity
   - Ensure object is created in PPMS

2. **Compliance check failures**:
   - Review required fields
   - Check documentation completeness
   - Verify approval workflow

3. **IP scan results not available**:
   - Ensure OSS stage ran first
   - Check result file upload
   - Verify file patterns

## Best Practices

### Scheduling

1. **Frequency**:
   - Security scans: Weekly or more frequent
   - PPMS checks: Monthly or before releases
   - Consider scan duration and resource usage

2. **Timing**:
   - Run during off-peak hours
   - Avoid concurrent scans
   - Coordinate with team availability

### Scan Configuration

1. **Scope**:
   - Include all dependencies
   - Scan production code and dependencies
   - Exclude test-only dependencies if appropriate

2. **Policies**:
   - Define clear failure criteria
   - Configure severity thresholds
   - Implement exemption process

3. **Results Management**:
   - Review results regularly
   - Track remediation progress
   - Maintain audit trail

### Integration

1. **Main Pipeline**:
   - Don't block main pipeline on scans
   - Run comprehensive scans separately
   - Quick scans in PR/push, full scans scheduled

2. **Notifications**:
   - Alert on new critical vulnerabilities
   - Summary reports to stakeholders
   - Track metrics over time

3. **Remediation**:
   - Automated dependency updates
   - Track vulnerability lifecycle
   - Document exemptions

## Compliance Requirements

### Corporate Requirements

- **PSL-1**: Product Security License compliance
- **IP Scanning**: License and IP compliance
- **PPMS**: Product portfolio documentation
- **ECCN**: Export control classification

### Evidence Collection

Both stages upload evidence to Cumulus:

- Scan reports (PDF, HTML, JSON)
- SARIF format for security findings
- SBOM in CycloneDX/SPDX format
- Policy evaluation results
- Compliance attestations

## Security Considerations

1. **Credentials**: Store all scanner credentials in Vault
2. **Results**: Protect scan results (may contain sensitive info)
3. **Access Control**: Limit who can view vulnerability details
4. **Remediation**: Have process for addressing findings
5. **Audit**: Maintain complete audit trail

## Related Documentation

- **Security Stage** (Jenkins): Original security stage documentation
- **PPMS Documentation**: PPMS system user guides
- **WhiteSource**: WhiteSource/Mend documentation
- **BlackDuck**: Synopsys BlackDuck documentation
- **Protecode**: BlackDuck Binary Analysis documentation

## Support and Resources

- **Hyperspace Team**: For pipeline configuration
- **Security Team**: For scan policy questions
- **PPMS Support**: For PPMS compliance issues
- **Tool Vendors**: For scanner-specific questions
