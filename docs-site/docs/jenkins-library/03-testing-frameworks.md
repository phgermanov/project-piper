# Testing Frameworks in Jenkins Library

This guide covers all testing steps available in the Jenkins library for various testing scenarios including unit tests, integration tests, performance tests, and end-to-end tests.

## Table of Contents

- [batsExecuteTests - Bash Testing](#batsexecutetests)
- [gatlingExecuteTests - Performance Testing](#gatlingexecutetests)
- [gaugeExecuteTests - BDD Testing](#gaugeexecutetests)
- [karmaExecuteTests - JavaScript Unit Testing](#karmaexecutetests)
- [newmanExecute - API Testing](#newmanexecute)
- [seleniumExecuteTests - UI Testing](#seleniumexecutetests)
- [uiVeri5ExecuteTests - SAP UI5 Testing](#uiveri5executetests)

---

## batsExecuteTests

### Overview
Executes Bash tests using the [Bash Automated Testing System (Bats)](https://github.com/bats-core/bats-core). Bats is a TAP-compliant testing framework for Bash that provides a simple way to verify UNIX programs behave as expected.

### Basic Usage

```groovy
batsExecuteTests script: this
testsPublishResults junit: [pattern: '**/Test-*.xml', archive: true]
```

### Configuration (YAML)

```yaml
steps:
  batsExecuteTests:
    testPath: 'src/test'
    outputFormat: 'junit'
    testPackage: 'my-project-tests'
    envVars:
      - 'CONTAINER_NAME=my-container'
      - 'IMAGE_NAME=my-image:latest'
```

### Key Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| testPath | string | `src/test` | Directory containing test files (*.bats) or a single file |
| outputFormat | string | `junit` | Test result format: `junit` or `tap` |
| repository | string | `https://github.com/bats-core/bats-core.git` | Bats-core repository URL |
| testPackage | string | `piper-bats` | Test package name in xUnit result file |
| envVars | []string | - | Environment variables injected during test execution |

### Features
- TAP-compliant test results
- JUnit format support for CI/CD integration
- Environment variable injection for dynamic test configuration
- Docker container testing support

### Best Practices
- Store test files in `src/test` directory by default
- Use descriptive test names with `@test` annotations
- Leverage environment variables for flexible test configuration
- Use JUnit format for automated build environments

---

## gatlingExecuteTests

### Overview
Executes performance tests using [Gatling](https://gatling.io/). Requires the Jenkins Gatling plugin to be installed. Ideal for load testing and performance benchmarking.

### Basic Usage

```groovy
gatlingExecuteTests script: this, pomPath: 'performance-tests/pom.xml'
```

### Configuration (YAML)

```yaml
steps:
  gatlingExecuteTests:
    pomPath: 'performance-tests/pom.xml'
    failOnError: true
    appUrls:
      - url: 'http://test-app.example.com'
        credentialsId: 'test-credentials'
```

### Key Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| pomPath | string | - | Path to pom.xml containing the performance test Maven module (mandatory) |
| failOnError | bool | `true` | Behavior when tests fail |
| appUrls | []Map | - | List of app URLs with Jenkins credential IDs |

### Features
- Maven-based test execution
- Credential management for protected endpoints
- Automatic test result archiving via Gatling plugin
- Support for multiple target URLs

### Best Practices
- Define performance test scenarios in separate Maven module
- Use `appUrls` parameter for testing multiple environments
- Store credentials in Jenkins credential store
- Review Gatling reports for performance insights

---

## gaugeExecuteTests

### Overview
Executes BDD acceptance tests using [Gauge](https://gauge.org). Supports three-tier test layout: acceptance criteria, test implementation layer, and application driver layer.

### Basic Usage

```groovy
gaugeExecuteTests script: this, testServerUrl: 'http://test.url'
```

### Configuration (YAML)

```yaml
steps:
  gaugeExecuteTests:
    installCommand: 'npm install -g @getgauge/cli@1.2.1'
    languageRunner: 'java'
    runCommand: 'run -s -p specs/'
    testOptions: '-e chrome'
```

### Key Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| installCommand | string | - | Command for installing Gauge (e.g., npm install) |
| languageRunner | string | - | Gauge language runner (e.g., `java`, `javascript`) |
| runCommand | string | - | Command for executing Gauge tests (mandatory) |
| testOptions | string | - | Specific options for Gauge execution |
| testServerUrl | string | - | URL of the test server |

### Features
- Markdown-based test specifications
- Selenium integration with Chrome sidecar
- Multi-language support (Java, JavaScript, Python, etc.)
- Maintainable acceptance test suites

### Best Practices
- Write specifications in Markdown for readability
- Use Gauge Maven archetypes for project setup
- Separate test specifications from implementation
- Leverage Selenium integration for UI testing

---

## karmaExecuteTests

### Overview
Executes JavaScript unit tests using [Karma test runner](http://karma-runner.github.io). **Note: Karma is deprecated as of 04/2023**.

### Basic Usage

```groovy
karmaExecuteTests script: this, modules: ['./shoppinglist', './catalog']
```

### Configuration (YAML)

```yaml
steps:
  karmaExecuteTests:
    installCommand: 'npm install --quiet'
    runCommand: 'npm run karma'
    modules:
      - './module1'
      - './module2'
```

### Key Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| installCommand | string | `npm install --quiet` | Command to install test tool (mandatory) |
| runCommand | string | `npm run karma` | Command to start tests (mandatory) |
| modules | []string | `['.']` | Paths of modules to execute tests on (mandatory) |

### Features
- Selenium Chrome sidecar container
- Multi-module testing support
- WebDriver-based browser testing
- Test result and coverage reporting

### Best Practices
- Configure karma-webdriver-launcher for WebDriver support
- Use `localhost` hostname in Kubernetes environments
- Test multiple modules in a single execution
- Publish test results using testsPublishResults step

---

## newmanExecute

### Overview
Executes API tests using [Newman](https://www.getpostman.com/docs/v6/postman/collection_runs/command_line_integration_with_newman), the command-line runner for Postman collections.

### Basic Usage

```groovy
newmanExecute script: this
testsPublishResults script: this, junit: [pattern: '**/newman/TEST-*.xml']
```

### Configuration (YAML)

```yaml
steps:
  newmanExecute:
    newmanCollection: '**/*.postman_collection.json'
    newmanEnvironment: 'test-env.json'
    newmanGlobals: 'globals.json'
    failOnError: false
    runOptions:
      - 'run'
      - '{{.NewmanCollection}}'
      - '--reporters'
      - 'cli,junit,html'
      - '--reporter-junit-export'
      - 'target/newman/TEST-{{.CollectionDisplayName}}.xml'
```

### Key Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| newmanCollection | string | `**/*.postman_collection.json` | Test collection file pattern |
| newmanEnvironment | string | - | Environment file path or URL |
| newmanGlobals | string | - | Global variables file path or URL |
| runOptions | []string | See default | Newman command options using Go templating |
| newmanInstallCommand | string | `npm install newman newman-reporter-html --global --quiet` | Newman installation command |
| failOnError | bool | `true` | Behavior when tests fail |

### Features
- Multiple reporter support (CLI, JUnit, HTML)
- Environment and global variable support
- Go templating for dynamic configuration
- Credential injection via environment variables

### Best Practices
- Use `runOptions` instead of deprecated `newmanRunCommand`
- Pass credentials via environment variables, not CLI options
- Generate both JUnit and HTML reports
- Use Go templating syntax: `{{.NewmanCollection}}`

### Passing Credentials Example

```yaml
runOptions:
  - 'run'
  - '{{.NewmanCollection}}'
  - '--environment'
  - '{{.Config.NewmanEnvironment}}'
  - '--env-var'
  - 'username={{getenv "PIPER_TESTCREDENTIAL_USERNAME"}}'
  - '--env-var'
  - 'password={{getenv "PIPER_TESTCREDENTIAL_PASSWORD"}}'
```

---

## seleniumExecuteTests

### Overview
Enables UI test execution with Selenium in a sidecar container. Provides a flexible framework for running Selenium-based tests with WebDriverIO or similar tools.

### Basic Usage

```groovy
seleniumExecuteTests(script: this) {
    git url: 'https://github.com/myorg/WebDriverIOTest.git'
    sh '''npm install
        node index.js'''
}
```

### Configuration (YAML)

```yaml
steps:
  seleniumExecuteTests:
    dockerImage: 'node:24-bookworm'
    dockerName: 'selenium-tests'
    sidecarImage: 'selenium/standalone-chrome'
    sidecarName: 'selenium'
    failOnError: true
    testRepository: 'git@github.com:myorg/tests.git'
    gitBranch: 'main'
```

### Key Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| buildTool | string | - | Tool for test execution: `maven`, `npm`, `bundler` |
| dockerImage | string | - | Docker image for test execution |
| dockerName | string | - | Name of Docker container |
| sidecarImage | string | - | Selenium sidecar image |
| sidecarName | string | - | Selenium sidecar container name |
| failOnError | bool | `true` | Behavior when tests fail |
| testRepository | string | - | External repository with test implementation |
| gitBranch | string | `master` | Test repository branch |
| seleniumHubCredentialsId | string | - | Credentials for Selenium Hub |

### Features
- Selenium Chrome sidecar container
- Support for external test repositories
- Docker network isolation
- WebDriver environment variable injection
- Kubernetes and local Docker support

### Best Practices
- Set Selenium host to `selenium` in local Docker, `localhost` in Kubernetes
- Add `localhost`/`selenium` to NO_PROXY environment variable
- Use dedicated test repositories for better organization
- Leverage environment variables for configuration

### WebDriverIO Configuration Example

```javascript
// Local Docker
var options = {
    host: 'selenium',
    port: 4444,
    desiredCapabilities: { browserName: 'chrome' }
};

// Kubernetes
var options = {
    host: 'localhost',
    port: 4444,
    desiredCapabilities: { browserName: 'chrome' }
};
```

---

## uiVeri5ExecuteTests

### Overview
Executes SAP UI5 end-to-end tests using [UIVeri5](https://github.com/SAP/ui5-uiveri5). Designed specifically for testing SAPUI5 and OpenUI5 applications.

### Basic Usage

```groovy
uiVeri5ExecuteTests script: this,
    runOptions: ['--seleniumAddress=http://localhost:4444/wd/hub', './uiveri5/conf.js']
```

### Configuration (YAML)

```yaml
steps:
  uiVeri5ExecuteTests:
    installCommand: 'npm install @ui5/uiveri5 --global --quiet'
    runCommand: '/home/node/.npm-global/bin/uiveri5'
    runOptions:
      - '--seleniumAddress=http://localhost:4444/wd/hub'
      - './path/to/conf.js'
    testServerUrl: 'http://my-app.example.com'
```

### Key Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| installCommand | string | `npm install @ui5/uiveri5 --global --quiet` | UIVeri5 installation command (mandatory) |
| runCommand | string | `/home/node/.npm-global/bin/uiveri5` | Command to start tests (mandatory) |
| runOptions | []string | `['--seleniumAddress=http://localhost:4444/wd/hub']` | Options including path to conf.js (mandatory) |
| testServerUrl | string | - | URL pointing to the deployment |

### Features
- Selenium Chrome sidecar integration
- SAPUI5/OpenUI5 specific testing capabilities
- Credential management support
- Environment variable templating
- Authentication handling (XSUAA, Basic, etc.)

### Best Practices
- Always specify `--seleniumAddress` in runOptions
- Use `localhost` for Kubernetes, may use `selenium` for native Jenkins
- Point runOptions to your `conf.js` file path
- Store credentials in Jenkins or Vault, inject via environment
- Use environment variables for authentication parameters

### Passing Credentials Example

```groovy
withCredentials([usernamePassword(
    credentialsId: 'MY_ACCEPTANCE_CREDENTIALS',
    passwordVariable: 'TEST_PASS',
    usernameVariable: 'TEST_USER'
)]) {
    uiVeri5ExecuteTests script: this,
        runOptions: [
            '--seleniumAddress=http://localhost:4444/wd/hub',
            './uiveri5/conf.js'
        ]
}
```

### UIVeri5 Config with Authentication

```javascript
const defaultParams = {
    url: process.env.TARGET_SERVER_URL,
    user: process.env.TEST_USER,
    pass: process.env.TEST_PASS
};

exports.config = {
    profile: 'integration',
    baseUrl: '${params.url}',
    params: defaultParams,
    auth: {
        'sapcloud-form': {
            user: '${params.user}',
            pass: '${params.pass}',
            userFieldSelector: 'input[id="j_username"]',
            passFieldSelector: 'input[id="j_password"]',
            logonButtonSelector: 'button[type="submit"]'
        }
    }
};
```

---

## General Testing Best Practices

1. **Test Result Publishing**: Always publish test results using `testsPublishResults` step
2. **Failure Handling**: Use `failOnError: false` when you need to continue pipeline after test failures
3. **Credential Management**: Store credentials in Jenkins credential store or Vault, never in code
4. **Environment Variables**: Use environment variables for configuration flexibility
5. **Docker Images**: Use specific image versions instead of `latest` for reproducibility
6. **Reporting**: Generate multiple report formats (JUnit, HTML) for different audiences
7. **Test Organization**: Separate test code from application code in dedicated directories
8. **CI/CD Integration**: Ensure all test steps produce machine-readable results (JUnit XML)

## Additional Resources

- [Project Piper Documentation](https://www.project-piper.io/)
- [Jenkins Library Steps](https://www.project-piper.io/steps/)
- [Configuration Guide](https://www.project-piper.io/configuration/)
