# Glossary

Terms and definitions used in Project Piper and CI/CD for SAP.

## Table of Contents

- [General CI/CD Terms](#general-cicd-terms)
- [SAP-Specific Terms](#sap-specific-terms)
- [Security & Compliance](#security--compliance)
- [Cloud Platforms](#cloud-platforms)
- [Build & Development](#build--development)
- [Container & Kubernetes](#container--kubernetes)
- [Testing & Quality](#testing--quality)

---

## General CI/CD Terms

### Artifact
A deployable package produced by a build process (e.g., JAR, WAR, MTA archive, Docker image). Artifacts are typically versioned and stored in artifact repositories.

### Blue-Green Deployment
A deployment strategy where two identical production environments (blue and green) exist. One serves live traffic while the other is updated, then traffic switches. Enables zero-downtime deployments.

### Build Tool
Software that automates the compilation and packaging of source code into artifacts. Examples: Maven, npm, Gradle, MTA Build Tool.

### CI/CD
**Continuous Integration/Continuous Delivery (or Deployment)**. Practices that automate code integration, testing, and deployment to production.

### Container
A lightweight, standalone executable package that includes everything needed to run software: code, runtime, libraries, and settings. Docker containers are most common.

### Container Image
A read-only template used to create containers. Images are built from Dockerfiles and stored in registries.

### Container Registry
A repository for storing and distributing container images. Examples: Docker Hub, Google Container Registry, Harbor, SAP Docker Registry.

### Pipeline
An automated sequence of stages and steps that source code passes through from commit to deployment. Defined in Jenkinsfile or similar configuration.

### Pipeline as Code
Practice of defining CI/CD pipelines in version-controlled configuration files (e.g., Jenkinsfile) rather than GUI configurations.

### Shared Library
Reusable pipeline code that can be shared across multiple projects. Project Piper is a Jenkins shared library.

### Stage
A logical grouping of related steps in a pipeline (e.g., Build stage, Test stage, Deploy stage).

### Step
A single task in a pipeline (e.g., compile code, run tests, deploy application).

### Webhook
An HTTP callback that triggers pipeline execution when events occur in source control (e.g., git push, pull request).

---

## SAP-Specific Terms

### AAKaaS
**Addon Assembly Kit as a Service**. SAP service for managing ABAP add-on development and assembly processes.

### ABAP
**Advanced Business Application Programming**. SAP's proprietary programming language for business applications.

### ABAP Environment
SAP BTP ABAP Environment (formerly known as Steampunk). Cloud-based ABAP development and runtime environment on SAP Business Technology Platform.

### Add-on
A reusable software package built on SAP systems that provides additional functionality. Add-ons can be installed in multiple customer systems.

### ATC
**ABAP Test Cockpit**. Quality assurance tool for ABAP development that performs static code analysis, checks coding conventions, and identifies potential issues.

### AUnit
**ABAP Unit**. Unit testing framework for ABAP code, similar to JUnit for Java.

### BTP
**SAP Business Technology Platform**. SAP's Platform-as-a-Service (PaaS) offering that provides development, integration, and extension capabilities.

### CAP
**Cloud Application Programming Model**. SAP's framework and set of tools for building enterprise-grade services and applications.

### CF
**Cloud Foundry**. Open-source PaaS used by SAP BTP for deploying and managing cloud applications.

### CPE
**Common Pipeline Environment**. Shared data structure in Piper pipelines that passes information between steps.

### CPI
**Cloud Platform Integration** (or SAP Integration Suite). SAP's integration platform for connecting cloud and on-premise applications.

### Fiori
SAP's user experience (UX) design approach for enterprise applications, providing responsive and intuitive interfaces.

### gCTS
**Git-enabled Change and Transport System**. Integration of Git version control with SAP's traditional transport system for ABAP development.

### HANA
**SAP High-Performance Analytic Appliance**. SAP's in-memory database platform.

### MTA
**Multi-Target Application**. SAP's application deployment format that bundles multiple modules (Java, Node.js, HTML5, etc.) with their dependencies and configurations.

### Neo
SAP's proprietary cloud runtime environment (previous generation, being phased out in favor of Cloud Foundry).

### SAP S/4HANA
SAP's next-generation ERP suite built on SAP HANA database.

### SOLMAN
**SAP Solution Manager**. SAP's application lifecycle management tool for implementation, support, and operations.

### Steampunk
Original codename for SAP BTP ABAP Environment. Now officially called ABAP Environment.

### TMS
**Transport Management Service**. SAP BTP service for managing transports across landscapes for MTA applications.

### Transport Request
A container in SAP systems that holds changes (code, configurations) for moving between systems (DEV → QA → PROD).

### UI5
**SAPUI5** or **OpenUI5**. SAP's JavaScript framework for building responsive web applications.

### XS
**XS Advanced** or **XS Classic**. SAP's application server technology for developing and running applications on SAP HANA.

---

## Security & Compliance

### BDBA
**Black Duck Binary Analysis** (formerly Protecode). Tool for scanning binaries and containers for open-source vulnerabilities and license compliance.

### CVE
**Common Vulnerabilities and Exposures**. Standardized identifier for known security vulnerabilities.

### CVSS
**Common Vulnerability Scoring System**. Industry standard for assessing severity of security vulnerabilities (scale 0-10).

### DAST
**Dynamic Application Security Testing**. Security testing performed on running applications to find vulnerabilities exposed during execution.

### License Compliance
Ensuring that open-source and third-party components used in software comply with their license terms and organizational policies.

### OWASP
**Open Web Application Security Project**. Nonprofit organization focused on improving software security. Known for OWASP Top 10 security risks.

### SAST
**Static Application Security Testing**. Security testing performed on source code without executing the program to find vulnerabilities.

### SCA
**Software Composition Analysis**. Process of identifying and analyzing open-source and third-party components in applications for security and license risks.

### SBOM
**Software Bill of Materials**. Inventory of all components (libraries, dependencies) in a software application, used for security and compliance tracking.

### Secrets Management
Practice of securely storing, accessing, and managing sensitive information (passwords, API keys, certificates) used by applications and pipelines.

### Vulnerability
A weakness in software that can be exploited to compromise security, availability, or integrity.

### Vulnerability Threshold
Acceptable number of vulnerabilities of each severity level before failing a security scan or build.

---

## Cloud Platforms

### API Management
Service for creating, securing, publishing, and analyzing APIs. SAP provides API Management on BTP.

### Cloud Foundry Space
Isolated environment within a Cloud Foundry organization where applications and services are deployed.

### Cloud Foundry Org
**Organization**. Top-level container in Cloud Foundry that contains spaces and has resource quotas.

### Kubernetes
Open-source container orchestration platform for automating deployment, scaling, and management of containerized applications.

### Manifest
YAML file that describes application deployment configuration for Cloud Foundry (manifest.yml) or Kubernetes (deployment.yaml).

### Namespace
Logical isolation mechanism in Kubernetes for grouping resources and providing scope for names.

### PaaS
**Platform as a Service**. Cloud computing model providing a platform for developing, running, and managing applications without managing infrastructure.

### Service Binding
Connection between an application and a service instance (e.g., database) that provides credentials and connection information.

### Service Broker
Component that manages lifecycle of service instances and creates service bindings in Cloud Foundry.

### Service Instance
Running instance of a service (database, messaging, etc.) provisioned for use by applications.

### Service Key
Credentials for accessing a service instance, typically used by external applications or APIs.

---

## Build & Development

### Build Tool
Software for compiling, testing, and packaging code. Examples: Maven (Java), npm (Node.js), pip (Python), Gradle (Java/Kotlin).

### Dependency Management
Process of managing external libraries and frameworks that an application depends on. Build tools handle dependency resolution.

### Gradle
Build automation tool primarily for Java projects. Uses Groovy or Kotlin DSL for build configuration.

### Maven
Build automation and dependency management tool for Java projects. Uses XML (pom.xml) for configuration.

### Maven Central
Central repository for Java libraries and dependencies.

### npm
**Node Package Manager**. Package manager and build tool for JavaScript/Node.js projects.

### Package Manager
Tool for installing, updating, and managing software dependencies. Examples: npm, yarn, pnpm, pip, Maven.

### pnpm
Fast, disk-space-efficient alternative to npm for managing JavaScript dependencies.

### POM
**Project Object Model**. XML file (pom.xml) used by Maven to define project configuration, dependencies, and build settings.

### Semantic Versioning
Versioning scheme using MAJOR.MINOR.PATCH format (e.g., 1.4.2). Defines when to increment each number based on type of changes.

### Snapshot Version
Development version of an artifact (e.g., 1.0.0-SNAPSHOT) that may change. Contrasts with release versions which are immutable.

### yarn
Alternative package manager for JavaScript, compatible with npm registry.

---

## Container & Kubernetes

### CNB
**Cloud Native Buildpacks**. CNCF project providing framework for building container images from source code without Dockerfiles.

### Deployment
Kubernetes resource that manages replicated application pods and handles updates.

### Docker
Platform for building, distributing, and running containers using OS-level virtualization.

### Dockerfile
Text file containing instructions for building a Docker container image.

### Docker Compose
Tool for defining and running multi-container Docker applications using YAML configuration.

### Docker Hub
Public container registry maintained by Docker, Inc. Default registry for Docker.

### Docker-in-Docker (DinD)
Running Docker daemon inside a Docker container. Often used in CI/CD pipelines.

### Hadolint
Linter for Dockerfiles that checks for best practices and common mistakes.

### Helm
Package manager for Kubernetes. Uses charts (packages) to define, install, and upgrade Kubernetes applications.

### Helm Chart
Package format for Kubernetes applications, containing all resource definitions and configuration.

### Image Tag
Version identifier for container images (e.g., `node:18-alpine` where `18-alpine` is the tag).

### Kaniko
Tool for building container images from Dockerfiles inside containers without requiring Docker daemon. Used in Kubernetes and restricted environments.

### kubectl
Command-line tool for interacting with Kubernetes clusters.

### Kubeconfig
Configuration file containing cluster connection information and credentials for kubectl.

### Multi-Stage Build
Dockerfile technique using multiple FROM statements to optimize image size and security by separating build and runtime stages.

### Pod
Smallest deployable unit in Kubernetes, containing one or more containers.

### Registry Mirror
Caching proxy for container images that reduces bandwidth and speeds up pulls.

---

## Testing & Quality

### Code Coverage
Metric measuring percentage of code executed by tests. Generated by tools like JaCoCo, Istanbul.

### E2E Testing
**End-to-End Testing**. Testing complete application workflows from user perspective.

### Integration Testing
Testing interactions between different components or systems.

### JaCoCo
**Java Code Coverage**. Library for measuring code coverage in Java applications.

### JUnit
Unit testing framework for Java.

### Karma
Test runner for JavaScript, commonly used with Angular applications.

### Linting
Static code analysis to check for programming errors, bugs, stylistic issues, and suspicious constructs.

### Newman
Command-line runner for Postman collections, used for API testing.

### Performance Testing
Testing to determine system behavior under load. Tools: Gatling, JMeter.

### PMD
Static code analyzer for Java that finds common programming flaws.

### SonarQube
Platform for continuous inspection of code quality, performing static analysis to detect bugs, code smells, and security vulnerabilities.

### SpotBugs
Static analysis tool for finding bugs in Java code (successor to FindBugs).

### Test Automation
Practice of using software tools to execute tests automatically rather than manually.

### UIVeri5
End-to-end testing framework for UI5 applications.

### Unit Testing
Testing individual components or functions in isolation.

---

## Project Piper Specific

### Common Pipeline Environment (CPE)
Shared memory structure in Piper that stores and passes data between pipeline steps. Accessible via `script.commonPipelineEnvironment`.

### Config.yml
Primary configuration file for Piper pipelines located at `.pipeline/config.yml` in project root.

### Piper
**Project Piper**. SAP's open-source project providing CI/CD pipelines and reusable steps for SAP technologies.

### Piper Binary
Standalone executable version of Piper steps that can run outside Jenkins (CLI mode).

### Piper Step
Individual reusable building block in Piper library (e.g., `mavenBuild`, `cloudFoundryDeploy`).

### Stage Configuration
Configuration specific to a pipeline stage in `.pipeline/config.yml`.

### Step Configuration
Configuration specific to a pipeline step in `.pipeline/config.yml`.

### Stash
Temporary storage of files in Jenkins pipelines for use across different stages or nodes.

---

## Additional Resources

- **FAQ**: See [faq.md](faq.md) for common questions
- **Troubleshooting**: See [troubleshooting.md](troubleshooting.md) for issue resolution
- **Step Reference**: See [step-reference.md](step-reference.md) for all available steps
- **Main Documentation**: [https://sap.github.io/jenkins-library/](https://sap.github.io/jenkins-library/)

---

## Acronyms Quick Reference

| Acronym | Full Name |
|---------|-----------|
| AAKaaS | Addon Assembly Kit as a Service |
| ABAP | Advanced Business Application Programming |
| API | Application Programming Interface |
| ATC | ABAP Test Cockpit |
| BDBA | Black Duck Binary Analysis |
| BTP | SAP Business Technology Platform |
| CAP | Cloud Application Programming Model |
| CD | Continuous Delivery/Deployment |
| CF | Cloud Foundry |
| CI | Continuous Integration |
| CLI | Command Line Interface |
| CNB | Cloud Native Buildpacks |
| CPI | Cloud Platform Integration |
| CPE | Common Pipeline Environment |
| CVE | Common Vulnerabilities and Exposures |
| CVSS | Common Vulnerability Scoring System |
| DAST | Dynamic Application Security Testing |
| DinD | Docker-in-Docker |
| DSL | Domain-Specific Language |
| E2E | End-to-End |
| ERP | Enterprise Resource Planning |
| gCTS | Git-enabled Change and Transport System |
| GUI | Graphical User Interface |
| HANA | High-Performance Analytic Appliance |
| HTTP | Hypertext Transfer Protocol |
| HTTPS | HTTP Secure |
| IDE | Integrated Development Environment |
| JSON | JavaScript Object Notation |
| K8s | Kubernetes |
| MTA | Multi-Target Application |
| OIDC | OpenID Connect |
| OS | Operating System |
| OWASP | Open Web Application Security Project |
| PaaS | Platform as a Service |
| PMD | Programming Mistake Detector |
| REST | Representational State Transfer |
| RFC | Remote Function Call |
| S/4HANA | SAP S/4HANA |
| SAR | SAP Archive |
| SAST | Static Application Security Testing |
| SBOM | Software Bill of Materials |
| SCA | Software Composition Analysis |
| SCM | Source Code Management |
| SDK | Software Development Kit |
| SOLMAN | SAP Solution Manager |
| SQL | Structured Query Language |
| SSH | Secure Shell |
| SSL | Secure Sockets Layer |
| TLS | Transport Layer Security |
| TMS | Transport Management Service |
| UI | User Interface |
| UI5 | SAPUI5/OpenUI5 |
| URL | Uniform Resource Locator |
| UX | User Experience |
| YAML | YAML Ain't Markup Language |
| XS | XS Advanced/Classic |
