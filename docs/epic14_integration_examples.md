# Epic 14 Framework - Integration Examples

## Overview

This document provides practical examples of integrating the Epic 14 Transition Chain Testing Framework with existing test suites, CI/CD pipelines, and testing tools.

## Table of Contents

1. [Integration with Go Testing](#integration-with-go-testing)
2. [Testify Integration](#testify-integration)
3. [Ginkgo Integration](#ginkgo-integration)
4. [CI/CD Pipeline Integration](#cicd-pipeline-integration)
5. [IDE Integration](#ide-integration)
6. [Docker Integration](#docker-integration)
7. [Monitoring Integration](#monitoring-integration)

## Integration with Go Testing

### Basic Integration

```go
package main

import (
    "testing"
    "github.com/mindreframer/agentpm/internal/testing/assertions"
    "github.com/mindreframer/agentpm/internal/testing/executor"
)

func TestBasicIntegration(t *testing.T) {
    result := executor.NewTransitionChain().
        StartEpic("integration-test").
        ExecutePhase("1A").
        CompleteEpic().
        Execute()
    
    // Use Check() for integration with standard Go error handling
    if err := assertions.Assert(result).
        EpicStatus("completed").
        NoErrors().
        Check(); err != nil {
        t.Fatalf("Epic 14 assertion failed: %v", err)
    }
}
```

### Table-Driven Tests Integration

```go
func TestTableDrivenIntegration(t *testing.T) {
    testCases := []struct {
        name           string
        epic           string
        phases         []string
        expectedStatus string
        shouldFail     bool
    }{
        {"success_case", "test-epic-1", []string{"1A", "1B"}, "completed", false},
        {"failure_case", "test-epic-2", []string{"1A", "INVALID"}, "failed", true},
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            chain := executor.NewTransitionChain().StartEpic(tc.epic)
            
            for _, phase := range tc.phases {
                chain = chain.ExecutePhase(phase)
            }
            
            result := chain.Execute()
            
            builder := assertions.Assert(result).EpicStatus(tc.expectedStatus)
            
            if tc.shouldFail {
                builder = builder.HasErrors()
            } else {
                builder = builder.NoErrors()
            }
            
            if err := builder.Check(); err != nil {
                t.Errorf("Test case %s failed: %v", tc.name, err)
            }
        })
    }
}
```

### Benchmark Integration

```go
func BenchmarkEpicExecution(b *testing.B) {
    for i := 0; i < b.N; i++ {
        result := executor.NewTransitionChain().
            StartEpic("benchmark-epic").
            ExecutePhase("1A").
            CompleteEpic().
            Execute()
        
        // Use minimal assertions for benchmarks
        assertions.Assert(result).
            WithDebugMode(assertions.DebugOff).
            EpicStatus("completed").
            MustPass()
    }
}
```

## Testify Integration

### Setup and Installation

```bash
go get github.com/stretchr/testify
```

### Basic Testify Integration

```go
package main

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/stretchr/testify/suite"
    "github.com/mindreframer/agentpm/internal/testing/assertions"
    "github.com/mindreframer/agentpm/internal/testing/executor"
)

func TestWithTestify(t *testing.T) {
    result := executor.NewTransitionChain().
        StartEpic("testify-integration").
        ExecutePhase("1A").
        CompleteEpic().
        Execute()
    
    // Use require for critical preconditions
    require.NotNil(t, result)
    require.NotNil(t, result.FinalState)
    
    // Use Epic 14 for domain-specific assertions
    err := assertions.Assert(result).
        EpicStatus("completed").
        PhaseStatus("1A", "completed").
        NoErrors().
        Check()
    
    // Use assert for validation
    assert.NoError(t, err)
    assert.Equal(t, "completed", result.FinalState.Status)
    assert.True(t, len(result.FinalState.Events) > 0, "Should have events")
}
```

### Testify Suite Integration

```go
type Epic14TestSuite struct {
    suite.Suite
    executor *executor.TransitionChain
}

func (suite *Epic14TestSuite) SetupTest() {
    suite.executor = executor.NewTransitionChain()
}

func (suite *Epic14TestSuite) TearDownTest() {
    suite.executor = nil
}

func (suite *Epic14TestSuite) TestEpicLifecycle() {
    result := suite.executor.
        StartEpic("suite-test").
        ExecutePhase("1A").
        CompleteEpic().
        Execute()
    
    // Combine testify assertions with Epic 14
    suite.Require().NotNil(result)
    
    err := assertions.Assert(result).
        EpicStatus("completed").
        Check()
    
    suite.NoError(err)
}

func (suite *Epic14TestSuite) TestPhaseProgression() {
    result := suite.executor.
        StartEpic("progression-test").
        ExecutePhase("1A").
        ExecutePhase("1B").
        CompleteEpic().
        Execute()
    
    // Use Epic 14 for complex state validation
    err := assertions.Assert(result).
        StateProgression([]string{"pending", "active", "completed"}).
        PhaseStatus("1A", "completed").
        PhaseStatus("1B", "completed").
        Check()
    
    suite.NoError(err)
}

func TestEpic14TestSuite(t *testing.T) {
    suite.Run(t, new(Epic14TestSuite))
}
```

### Custom Testify Assertions

```go
import (
    "github.com/stretchr/testify/assert"
)

// Custom assertion for Epic 14 results
func AssertEpicCompleted(t assert.TestingT, result *executor.TransitionChainResult, msgAndArgs ...interface{}) bool {
    if h, ok := t.(interface{ Helper() }); ok {
        h.Helper()
    }
    
    err := assertions.Assert(result).
        EpicStatus("completed").
        NoErrors().
        Check()
    
    return assert.NoError(t, err, msgAndArgs...)
}

func AssertPhaseStatus(t assert.TestingT, result *executor.TransitionChainResult, phaseID, expectedStatus string, msgAndArgs ...interface{}) bool {
    if h, ok := t.(interface{ Helper() }); ok {
        h.Helper()
    }
    
    err := assertions.Assert(result).
        PhaseStatus(phaseID, expectedStatus).
        Check()
    
    return assert.NoError(t, err, msgAndArgs...)
}

// Usage
func TestCustomAssertions(t *testing.T) {
    result := executeWorkflow()
    
    AssertEpicCompleted(t, result, "Epic should complete successfully")
    AssertPhaseStatus(t, result, "1A", "completed", "Phase 1A should be completed")
}
```

## Ginkgo Integration

### Setup and Installation

```bash
go get github.com/onsi/ginkgo/v2
go get github.com/onsi/gomega
```

### Basic Ginkgo Integration

```go
package main

import (
    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"
    "github.com/mindreframer/agentpm/internal/testing/assertions"
    "github.com/mindreframer/agentpm/internal/testing/executor"
)

var _ = Describe("Epic 14 Integration", func() {
    var result *executor.TransitionChainResult
    
    BeforeEach(func() {
        result = executor.NewTransitionChain().
            StartEpic("ginkgo-test").
            ExecutePhase("1A").
            CompleteEpic().
            Execute()
    })
    
    Context("When epic completes successfully", func() {
        It("should have completed status", func() {
            err := assertions.Assert(result).
                EpicStatus("completed").
                Check()
            
            Expect(err).NotTo(HaveOccurred())
        })
        
        It("should have no errors", func() {
            err := assertions.Assert(result).
                NoErrors().
                Check()
            
            Expect(err).NotTo(HaveOccurred())
        })
        
        It("should complete all phases", func() {
            err := assertions.Assert(result).
                PhaseStatus("1A", "completed").
                Check()
            
            Expect(err).NotTo(HaveOccurred())
        })
    })
    
    Context("When validating event sequence", func() {
        It("should have correct event progression", func() {
            err := assertions.Assert(result).
                EventSequence([]string{
                    "epic_started",
                    "phase_started",
                    "phase_completed",
                    "epic_completed",
                }).
                Check()
            
            Expect(err).NotTo(HaveOccurred())
        })
    })
})
```

### Ginkgo with Custom Matchers

```go
import (
    "github.com/onsi/gomega/types"
)

// Custom Gomega matcher for Epic 14
func BeCompletedEpic() types.GomegaMatcher {
    return &completedEpicMatcher{}
}

type completedEpicMatcher struct{}

func (matcher *completedEpicMatcher) Match(actual interface{}) (success bool, err error) {
    result, ok := actual.(*executor.TransitionChainResult)
    if !ok {
        return false, fmt.Errorf("BeCompletedEpic matcher expects a TransitionChainResult")
    }
    
    err = assertions.Assert(result).
        EpicStatus("completed").
        NoErrors().
        Check()
    
    return err == nil, nil
}

func (matcher *completedEpicMatcher) FailureMessage(actual interface{}) (message string) {
    return fmt.Sprintf("Expected epic to be completed, but it wasn't")
}

func (matcher *completedEpicMatcher) NegatedFailureMessage(actual interface{}) (message string) {
    return fmt.Sprintf("Expected epic not to be completed, but it was")
}

// Usage
var _ = Describe("Custom Matchers", func() {
    It("should use custom epic matcher", func() {
        result := executeWorkflow()
        Expect(result).To(BeCompletedEpic())
    })
})
```

## CI/CD Pipeline Integration

### GitHub Actions Integration

```yaml
# .github/workflows/epic14-tests.yml
name: Epic 14 Testing

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        go-version: [1.21, 1.22]
        test-suite: [unit, integration, performance]
        
    steps:
    - uses: actions/checkout@v4
    
    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: ${{ matrix.go-version }}
        
    - name: Cache dependencies
      uses: actions/cache@v3
      with:
        path: |
          ~/.cache/go-build
          ~/go/pkg/mod
        key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
        restore-keys: |
          ${{ runner.os }}-go-
          
    - name: Install dependencies
      run: go mod download
      
    - name: Run Epic 14 tests
      run: |
        case "${{ matrix.test-suite }}" in
          unit)
            go test -v -short -tags=unit ./internal/testing/assertions/...
            ;;
          integration)
            go test -v -tags=integration ./internal/testing/assertions/...
            ;;
          performance)
            go test -v -tags=performance -timeout=30m ./internal/testing/assertions/...
            ;;
        esac
      env:
        CI: true
        DEBUG_LEVEL: basic
        
    - name: Upload test results
      uses: actions/upload-artifact@v3
      if: always()
      with:
        name: test-results-${{ matrix.go-version }}-${{ matrix.test-suite }}
        path: |
          coverage.out
          test-results.xml
          
    - name: Upload coverage to Codecov
      if: matrix.test-suite == 'unit'
      uses: codecov/codecov-action@v3
      with:
        file: coverage.out
```

### Jenkins Pipeline Integration

```groovy
// Jenkinsfile
pipeline {
    agent any
    
    parameters {
        choice(
            name: 'TEST_LEVEL',
            choices: ['unit', 'integration', 'performance', 'all'],
            description: 'Test level to run'
        )
        booleanParam(
            name: 'ENABLE_DEBUG',
            defaultValue: false,
            description: 'Enable debug output'
        )
    }
    
    environment {
        GO_VERSION = '1.21'
        DEBUG_LEVEL = "${params.ENABLE_DEBUG ? 'verbose' : 'off'}"
    }
    
    stages {
        stage('Setup') {
            steps {
                sh 'go version'
                sh 'go mod download'
            }
        }
        
        stage('Unit Tests') {
            when {
                anyOf {
                    params.TEST_LEVEL == 'unit'
                    params.TEST_LEVEL == 'all'
                }
            }
            steps {
                sh '''
                    go test -v -short -tags=unit \
                    -coverprofile=coverage.out \
                    ./internal/testing/assertions/...
                '''
            }
            post {
                always {
                    publishHTML([
                        allowMissing: false,
                        alwaysLinkToLastBuild: true,
                        keepAll: true,
                        reportDir: '.',
                        reportFiles: 'coverage.html',
                        reportName: 'Coverage Report'
                    ])
                }
            }
        }
        
        stage('Integration Tests') {
            when {
                anyOf {
                    params.TEST_LEVEL == 'integration'
                    params.TEST_LEVEL == 'all'
                }
            }
            steps {
                sh '''
                    go test -v -tags=integration \
                    ./internal/testing/assertions/...
                '''
            }
        }
        
        stage('Performance Tests') {
            when {
                anyOf {
                    params.TEST_LEVEL == 'performance'
                    params.TEST_LEVEL == 'all'
                }
            }
            steps {
                sh '''
                    go test -v -tags=performance -timeout=30m \
                    ./internal/testing/assertions/...
                '''
            }
        }
    }
    
    post {
        always {
            archiveArtifacts artifacts: 'coverage.out', allowEmptyArchive: true
            publishTestResults testResultsPattern: 'test-results.xml'
        }
        failure {
            emailext (
                subject: "Epic 14 Tests Failed: ${env.JOB_NAME} - ${env.BUILD_NUMBER}",
                body: "Test suite failed. Check ${env.BUILD_URL} for details.",
                to: "${env.CHANGE_AUTHOR_EMAIL ?: 'team@example.com'}"
            )
        }
    }
}
```

### GitLab CI Integration

```yaml
# .gitlab-ci.yml
stages:
  - test
  - performance
  - deploy

variables:
  GO_VERSION: "1.21"
  CI: "true"

.go_template: &go_template
  image: golang:${GO_VERSION}
  before_script:
    - go version
    - go mod download
  cache:
    paths:
      - .cache/go-build/
      - go/pkg/mod/

unit_tests:
  <<: *go_template
  stage: test
  script:
    - go test -v -short -tags=unit -coverprofile=coverage.out ./internal/testing/assertions/...
    - go tool cover -html=coverage.out -o coverage.html
  coverage: '/coverage: \d+.\d+% of statements/'
  artifacts:
    reports:
      coverage_report:
        coverage_format: cobertura
        path: coverage.xml
    paths:
      - coverage.html
  rules:
    - if: $CI_PIPELINE_SOURCE == "merge_request_event"
    - if: $CI_COMMIT_BRANCH == "main"

integration_tests:
  <<: *go_template
  stage: test
  script:
    - go test -v -tags=integration ./internal/testing/assertions/...
  rules:
    - if: $CI_PIPELINE_SOURCE == "merge_request_event"
    - if: $CI_COMMIT_BRANCH == "main"

performance_tests:
  <<: *go_template
  stage: performance
  script:
    - go test -v -tags=performance -timeout=30m ./internal/testing/assertions/...
  artifacts:
    reports:
      performance: performance.json
  rules:
    - if: $CI_COMMIT_BRANCH == "main"
    - when: manual
```

## IDE Integration

### VS Code Integration

Create `.vscode/settings.json`:

```json
{
    "go.testFlags": [
        "-v",
        "-tags=unit"
    ],
    "go.testTimeout": "30s",
    "go.coverOnSave": true,
    "go.coverOnSingleTest": true,
    "go.coverageDecorator": {
        "type": "gutter",
        "coveredHighlightColor": "rgba(64,128,64,0.5)",
        "uncoveredHighlightColor": "rgba(128,64,64,0.25)"
    },
    "go.testExplorer.enable": true,
    "go.testExplorer.showDynamicSubtestsInEditor": true
}
```

Create `.vscode/tasks.json`:

```json
{
    "version": "2.0.0",
    "tasks": [
        {
            "label": "Run Epic 14 Unit Tests",
            "type": "shell",
            "command": "go",
            "args": [
                "test",
                "-v",
                "-short",
                "-tags=unit",
                "./internal/testing/assertions/..."
            ],
            "group": {
                "kind": "test",
                "isDefault": true
            },
            "presentation": {
                "echo": true,
                "reveal": "always",
                "focus": false,
                "panel": "shared"
            },
            "problemMatcher": "$go"
        },
        {
            "label": "Run Epic 14 Integration Tests",
            "type": "shell",
            "command": "go",
            "args": [
                "test",
                "-v",
                "-tags=integration",
                "./internal/testing/assertions/..."
            ],
            "group": "test",
            "presentation": {
                "echo": true,
                "reveal": "always",
                "focus": false,
                "panel": "shared"
            },
            "problemMatcher": "$go"
        }
    ]
}
```

### GoLand Integration

Create run configurations:

```xml
<!-- .idea/runConfigurations/Epic_14_Unit_Tests.xml -->
<component name="ProjectRunConfigurationManager">
  <configuration default="false" name="Epic 14 Unit Tests" type="GoTestRunConfiguration" factoryName="Go Test">
    <module name="agentpm" />
    <working_directory value="$PROJECT_DIR$" />
    <go_parameters value="-i" />
    <parameters value="-v -short -tags=unit" />
    <kind value="PACKAGE" />
    <package value="github.com/mindreframer/agentpm/internal/testing/assertions" />
    <directory value="$PROJECT_DIR$" />
    <filePath value="$PROJECT_DIR$" />
    <framework value="gotest" />
    <method v="2" />
  </configuration>
</component>
```

## Docker Integration

### Dockerfile for Testing

```dockerfile
# Dockerfile.test
FROM golang:1.21-alpine AS builder

# Install dependencies
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Run tests
FROM builder AS test
ENV CI=true
ENV DEBUG_LEVEL=basic

# Run unit tests
RUN go test -v -short -tags=unit -coverprofile=coverage.out ./internal/testing/assertions/...

# Run integration tests
RUN go test -v -tags=integration ./internal/testing/assertions/...

# Generate coverage report
RUN go tool cover -html=coverage.out -o coverage.html

# Performance testing stage
FROM builder AS performance
ENV CI=true
ENV DEBUG_LEVEL=off

RUN go test -v -tags=performance -timeout=30m ./internal/testing/assertions/...

# Final stage with test results
FROM alpine:latest AS results
RUN apk add --no-cache ca-certificates
COPY --from=test /app/coverage.html /results/
COPY --from=test /app/coverage.out /results/
COPY --from=performance /app/performance.log /results/

CMD ["sh", "-c", "echo 'Test results available in /results/'"]
```

### Docker Compose for Testing

```yaml
# docker-compose.test.yml
version: '3.8'

services:
  unit-tests:
    build:
      context: .
      dockerfile: Dockerfile.test
      target: test
    environment:
      - CI=true
      - DEBUG_LEVEL=basic
    volumes:
      - ./test-results:/results
    command: ["go", "test", "-v", "-short", "-tags=unit", "./internal/testing/assertions/..."]

  integration-tests:
    build:
      context: .
      dockerfile: Dockerfile.test
      target: test
    environment:
      - CI=true
      - DEBUG_LEVEL=basic
    depends_on:
      - unit-tests
    volumes:
      - ./test-results:/results
    command: ["go", "test", "-v", "-tags=integration", "./internal/testing/assertions/..."]

  performance-tests:
    build:
      context: .
      dockerfile: Dockerfile.test
      target: performance
    environment:
      - CI=true
      - DEBUG_LEVEL=off
    depends_on:
      - integration-tests
    volumes:
      - ./test-results:/results
    command: ["go", "test", "-v", "-tags=performance", "-timeout=30m", "./internal/testing/assertions/..."]
```

### Usage

```bash
# Run all tests
docker-compose -f docker-compose.test.yml up --build

# Run specific test suite
docker-compose -f docker-compose.test.yml run unit-tests

# Run with specific environment
docker-compose -f docker-compose.test.yml run -e DEBUG_LEVEL=verbose unit-tests
```

## Monitoring Integration

### Prometheus Metrics Integration

```go
package main

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    testsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "epic14_tests_total",
            Help: "Total number of Epic 14 tests executed",
        },
        []string{"status", "test_type"},
    )
    
    testDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "epic14_test_duration_seconds",
            Help: "Duration of Epic 14 test execution",
        },
        []string{"test_type"},
    )
    
    assertionFailures = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "epic14_assertion_failures_total",
            Help: "Total number of Epic 14 assertion failures",
        },
        []string{"assertion_type"},
    )
)

func TestWithMetrics(t *testing.T) {
    timer := prometheus.NewTimer(testDuration.WithLabelValues("unit"))
    defer timer.ObserveDuration()
    
    result := executeWorkflow()
    
    err := assertions.Assert(result).
        EpicStatus("completed").
        Check()
    
    if err != nil {
        testsTotal.WithLabelValues("failed", "unit").Inc()
        assertionFailures.WithLabelValues("epic_status").Inc()
        t.Fatal(err)
    } else {
        testsTotal.WithLabelValues("passed", "unit").Inc()
    }
}
```

### Datadog Integration

```go
package main

import (
    "github.com/DataDog/datadog-go/statsd"
)

func TestWithDatadog(t *testing.T) {
    client, err := statsd.New("127.0.0.1:8125")
    if err != nil {
        t.Skip("Datadog StatsD not available")
    }
    defer client.Close()
    
    start := time.Now()
    result := executeWorkflow()
    duration := time.Since(start)
    
    // Send timing metric
    client.Timing("epic14.test.duration", duration, []string{"test:unit"}, 1)
    
    err = assertions.Assert(result).
        EpicStatus("completed").
        Check()
    
    if err != nil {
        client.Incr("epic14.test.failure", []string{"test:unit"}, 1)
        t.Fatal(err)
    } else {
        client.Incr("epic14.test.success", []string{"test:unit"}, 1)
    }
}
```

### Custom Test Reporter

```go
type Epic14TestReporter struct {
    results []TestResult
    mu      sync.Mutex
}

type TestResult struct {
    Name      string
    Duration  time.Duration
    Status    string
    Errors    []string
    Timestamp time.Time
}

func (r *Epic14TestReporter) ReportTest(name string, duration time.Duration, err error) {
    r.mu.Lock()
    defer r.mu.Unlock()
    
    result := TestResult{
        Name:      name,
        Duration:  duration,
        Timestamp: time.Now(),
    }
    
    if err != nil {
        result.Status = "failed"
        result.Errors = []string{err.Error()}
    } else {
        result.Status = "passed"
    }
    
    r.results = append(r.results, result)
}

func (r *Epic14TestReporter) GenerateReport() string {
    r.mu.Lock()
    defer r.mu.Unlock()
    
    // Generate JSON or XML report
    data, _ := json.MarshalIndent(r.results, "", "  ")
    return string(data)
}

// Usage
func TestWithReporting(t *testing.T) {
    reporter := &Epic14TestReporter{}
    
    start := time.Now()
    result := executeWorkflow()
    
    err := assertions.Assert(result).
        EpicStatus("completed").
        Check()
    
    reporter.ReportTest(t.Name(), time.Since(start), err)
    
    if err != nil {
        t.Fatal(err)
    }
    
    // At end of test suite
    t.Cleanup(func() {
        report := reporter.GenerateReport()
        os.WriteFile("epic14-report.json", []byte(report), 0644)
    })
}
```

These integration examples demonstrate how to effectively incorporate the Epic 14 framework into existing testing workflows and infrastructure, providing flexibility while maintaining the framework's powerful assertion capabilities.