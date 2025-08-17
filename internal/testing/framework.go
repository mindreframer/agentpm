package testing

import (
	"github.com/mindreframer/agentpm/internal/testing/builders"
	"github.com/mindreframer/agentpm/internal/testing/executor"
)

// Framework provides the main API for the transition chain testing framework
// This serves as the entry point for all testing functionality

// EpicBuilder creates a new epic builder for test construction
func EpicBuilder(id string) *builders.EpicBuilder {
	return builders.NewEpicBuilder(id)
}

// NewTestEnvironment creates a new isolated test execution environment
func NewTestEnvironment(epicFile string) *executor.TestExecutionEnvironment {
	return executor.NewTestExecutionEnvironment(epicFile)
}

// TransitionChain creates a new transition chain for the given environment
func TransitionChain(env *executor.TestExecutionEnvironment) *executor.TransitionChain {
	return executor.CreateTransitionChain(env)
}

// TestFramework provides access to all testing components
type TestFramework struct {
	// This can be extended in the future to hold shared configuration
}

// NewTestFramework creates a new test framework instance
func NewTestFramework() *TestFramework {
	return &TestFramework{}
}

// CreateEpicBuilder creates a new epic builder (alternative entry point)
func (tf *TestFramework) CreateEpicBuilder(id string) *builders.EpicBuilder {
	return builders.NewEpicBuilder(id)
}

// CreateTestEnvironment creates a new test environment (alternative entry point)
func (tf *TestFramework) CreateTestEnvironment(epicFile string) *executor.TestExecutionEnvironment {
	return executor.NewTestExecutionEnvironment(epicFile)
}
