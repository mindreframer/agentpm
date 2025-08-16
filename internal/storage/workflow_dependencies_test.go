package storage

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mindreframer/agentpm/internal/epic"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkflowRequirementsDependenciesPersistence(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "epic-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	storage := NewFileStorage()

	testEpic := epic.NewEpic("test-epic", "Test Epic")
	testEpic.Workflow = `**CRITICAL: Test-Driven Development Approach**

For **EACH** phase:
1. **Implement Code** - Complete the implementation tasks  
2. **Write Tests IMMEDIATELY** - Create comprehensive test coverage
3. **Run Tests Verify** - All tests must pass before proceeding
4. **Run Linting/Type Checking** - Code must be clean and follow standards
5. **NEVER move to next phase with failing tests**`

	testEpic.Requirements = `**Core Stories:**
- Replace in-memory school loading with database pagination
- Add pagination controls with page navigation
- Maintain URL state for bookmarkable paginated views
- Preserve existing filtering (status) and search functionality
- Display pagination metadata (showing X of Y schools)

**Technical Requirements:**
- Database-level pagination to handle hundreds of schools
- URL State Management - Page numbers, filters, and search terms in URL
- LiveView Integration - Real-time pagination without page reloads
- Mobile Responsive - Simplified pagination controls on mobile devices
- QuickCrud Integration - Leverage existing paginate() functionality`

	testEpic.Dependencies = `- Epic 1: Database schema (crm_schools table) and QuickCrud system (required)
- Epic 3: School management LiveView pages and existing filtering (required)
- Epic 4: Contact management for preloading optimization (optional)`

	epicPath := filepath.Join(tempDir, "test-epic.xml")
	err = storage.SaveEpic(testEpic, epicPath)
	require.NoError(t, err)

	loadedEpic, err := storage.LoadEpic(epicPath)
	require.NoError(t, err)

	assert.Equal(t, testEpic.ID, loadedEpic.ID)
	assert.Equal(t, testEpic.Name, loadedEpic.Name)
	assert.Equal(t, testEpic.Workflow, loadedEpic.Workflow, "Workflow should be preserved")
	assert.Equal(t, testEpic.Requirements, loadedEpic.Requirements, "Requirements should be preserved")
	assert.Equal(t, testEpic.Dependencies, loadedEpic.Dependencies, "Dependencies should be preserved")

	content, err := os.ReadFile(epicPath)
	require.NoError(t, err)

	xmlContent := string(content)
	assert.Contains(t, xmlContent, "<workflow>", "XML should contain workflow element")
	assert.Contains(t, xmlContent, "<requirements>", "XML should contain requirements element")
	assert.Contains(t, xmlContent, "<dependencies>", "XML should contain dependencies element")
	assert.Contains(t, xmlContent, "Test-Driven Development", "Workflow content should be in XML")
	assert.Contains(t, xmlContent, "Core Stories", "Requirements content should be in XML")
	assert.Contains(t, xmlContent, "Epic 1: Database schema", "Dependencies content should be in XML")
}
