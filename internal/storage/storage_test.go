package storage

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mindreframer/agentpm/internal/epic"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileStorage_LoadEpic(t *testing.T) {
	storage := NewFileStorage()

	t.Run("load valid epic", func(t *testing.T) {
		epicPath := filepath.Join("..", "..", "testdata", "epic-valid.xml")

		e, err := storage.LoadEpic(epicPath)
		require.NoError(t, err)

		assert.Equal(t, "8", e.ID)
		assert.Equal(t, "Epic Name", e.Name)
		assert.Equal(t, epic.StatusPending, e.Status)
		assert.Equal(t, "agent_claude", e.Assignee)
		assert.Equal(t, "Epic description", e.Description)

		// Check phases
		require.Len(t, e.Phases, 2)
		assert.Equal(t, "1A", e.Phases[0].ID)
		assert.Equal(t, "Setup", e.Phases[0].Name)
		assert.Equal(t, epic.StatusPending, e.Phases[0].Status)

		// Check tasks
		require.Len(t, e.Tasks, 2)
		assert.Equal(t, "1A_1", e.Tasks[0].ID)
		assert.Equal(t, "1A", e.Tasks[0].PhaseID)
		assert.Equal(t, "Initialize Project", e.Tasks[0].Name)
		assert.Equal(t, "agent_claude", e.Tasks[0].Assignee)

		// Check tests
		require.Len(t, e.Tests, 1)
		assert.Equal(t, "T1A_1", e.Tests[0].ID)
		assert.Equal(t, "1A_1", e.Tests[0].TaskID)
		assert.Equal(t, "Test Project Init", e.Tests[0].Name)
	})

	t.Run("load non-existent epic", func(t *testing.T) {
		_, err := storage.LoadEpic("non-existent.xml")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read epic file")
	})

	t.Run("load malformed XML", func(t *testing.T) {
		tempDir := t.TempDir()
		invalidPath := filepath.Join(tempDir, "invalid.xml")

		// Write malformed XML
		err := os.WriteFile(invalidPath, []byte(`<?xml version="1.0"?><epic>malformed xml`), 0644)
		require.NoError(t, err)

		_, err = storage.LoadEpic(invalidPath)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read epic file")
	})

	t.Run("missing epic root element", func(t *testing.T) {
		tempDir := t.TempDir()
		badPath := filepath.Join(tempDir, "bad.xml")

		// Create file with wrong root element
		err := os.WriteFile(badPath, []byte(`<?xml version="1.0"?><notepic></notepic>`), 0644)
		require.NoError(t, err)

		_, err = storage.LoadEpic(badPath)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing <epic> root element")
	})
}

func TestFileStorage_SaveEpic(t *testing.T) {
	storage := NewFileStorage()

	t.Run("save valid epic", func(t *testing.T) {
		tempDir := t.TempDir()
		epicPath := filepath.Join(tempDir, "test-epic.xml")

		e := &epic.Epic{
			ID:          "test-1",
			Name:        "Test Epic",
			Status:      epic.StatusPending,
			CreatedAt:   time.Date(2025, 8, 16, 9, 0, 0, 0, time.UTC),
			Assignee:    "test_agent",
			Description: "Test description",
			Phases: []epic.Phase{
				{
					ID:          "P1",
					Name:        "Phase 1",
					Status:      epic.StatusPending,
					Description: "First phase",
				},
			},
			Tasks: []epic.Task{
				{
					ID:          "T1",
					PhaseID:     "P1",
					Name:        "Task 1",
					Status:      epic.StatusPending,
					Assignee:    "test_agent",
					Description: "First task",
				},
			},
			Tests: []epic.Test{
				{
					ID:          "TEST1",
					TaskID:      "T1",
					Name:        "Test 1",
					Status:      epic.StatusPending,
					Description: "First test",
				},
			},
		}

		err := storage.SaveEpic(e, epicPath)
		assert.NoError(t, err)

		// Verify file exists
		assert.True(t, storage.EpicExists(epicPath))

		// Load it back and verify
		loaded, err := storage.LoadEpic(epicPath)
		require.NoError(t, err)

		assert.Equal(t, e.ID, loaded.ID)
		assert.Equal(t, e.Name, loaded.Name)
		assert.Equal(t, e.Status, loaded.Status)
		assert.Equal(t, e.Assignee, loaded.Assignee)
		assert.Equal(t, e.Description, loaded.Description)

		require.Len(t, loaded.Phases, 1)
		assert.Equal(t, "P1", loaded.Phases[0].ID)

		require.Len(t, loaded.Tasks, 1)
		assert.Equal(t, "T1", loaded.Tasks[0].ID)

		require.Len(t, loaded.Tests, 1)
		assert.Equal(t, "TEST1", loaded.Tests[0].ID)
	})

	t.Run("save creates directory", func(t *testing.T) {
		tempDir := t.TempDir()
		epicPath := filepath.Join(tempDir, "subdir", "epic.xml")

		e := &epic.Epic{
			ID:        "test-1",
			Name:      "Test Epic",
			Status:    epic.StatusPending,
			CreatedAt: time.Now(),
		}

		err := storage.SaveEpic(e, epicPath)
		assert.NoError(t, err)

		assert.True(t, storage.EpicExists(epicPath))
	})

	t.Run("atomic save operation", func(t *testing.T) {
		tempDir := t.TempDir()
		epicPath := filepath.Join(tempDir, "atomic.xml")

		// Create initial epic
		e1 := &epic.Epic{
			ID:        "test-1",
			Name:      "Original",
			Status:    epic.StatusPending,
			CreatedAt: time.Now(),
		}

		err := storage.SaveEpic(e1, epicPath)
		require.NoError(t, err)

		// Update epic
		e2 := &epic.Epic{
			ID:        "test-1",
			Name:      "Updated",
			Status:    epic.StatusActive,
			CreatedAt: time.Now(),
		}

		err = storage.SaveEpic(e2, epicPath)
		assert.NoError(t, err)

		// Verify final state
		loaded, err := storage.LoadEpic(epicPath)
		require.NoError(t, err)
		assert.Equal(t, "Updated", loaded.Name)
		assert.Equal(t, epic.StatusActive, loaded.Status)

		// Verify no temp file left behind
		tempFile := epicPath + ".tmp"
		_, err = os.Stat(tempFile)
		assert.True(t, os.IsNotExist(err))
	})
}

func TestFileStorage_EpicExists(t *testing.T) {
	storage := NewFileStorage()

	t.Run("existing file", func(t *testing.T) {
		epicPath := filepath.Join("..", "..", "testdata", "epic-valid.xml")
		assert.True(t, storage.EpicExists(epicPath))
	})

	t.Run("non-existing file", func(t *testing.T) {
		assert.False(t, storage.EpicExists("non-existent.xml"))
	})
}

func TestMemoryStorage(t *testing.T) {
	storage := NewMemoryStorage()

	t.Run("save and load epic", func(t *testing.T) {
		e := &epic.Epic{
			ID:        "mem-1",
			Name:      "Memory Epic",
			Status:    epic.StatusPending,
			CreatedAt: time.Now(),
		}

		err := storage.SaveEpic(e, "memory-epic.xml")
		assert.NoError(t, err)

		assert.True(t, storage.EpicExists("memory-epic.xml"))

		loaded, err := storage.LoadEpic("memory-epic.xml")
		require.NoError(t, err)

		assert.Equal(t, e.ID, loaded.ID)
		assert.Equal(t, e.Name, loaded.Name)
		assert.Equal(t, e.Status, loaded.Status)
	})

	t.Run("load non-existent epic", func(t *testing.T) {
		_, err := storage.LoadEpic("missing.xml")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "epic file not found")
	})

	t.Run("save nil epic", func(t *testing.T) {
		err := storage.SaveEpic(nil, "nil-epic.xml")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "epic cannot be nil")
	})

	t.Run("store epic directly", func(t *testing.T) {
		e := &epic.Epic{
			ID:   "direct-1",
			Name: "Direct Epic",
		}

		storage.StoreEpic("direct.xml", e)

		assert.True(t, storage.EpicExists("direct.xml"))

		loaded, err := storage.LoadEpic("direct.xml")
		require.NoError(t, err)
		assert.Equal(t, "direct-1", loaded.ID)
	})
}

func TestStorageFactory(t *testing.T) {
	t.Run("create memory storage", func(t *testing.T) {
		factory := NewFactory(true)
		storage := factory.CreateStorage()

		_, ok := storage.(*MemoryStorage)
		assert.True(t, ok, "should create MemoryStorage")
	})

	t.Run("create file storage", func(t *testing.T) {
		factory := NewFactory(false)
		storage := factory.CreateStorage()

		_, ok := storage.(*FileStorage)
		assert.True(t, ok, "should create FileStorage")
	})
}

func TestXMLRoundTrip(t *testing.T) {
	storage := NewFileStorage()
	tempDir := t.TempDir()

	original := &epic.Epic{
		ID:          "rt-1",
		Name:        "Round Trip Epic",
		Status:      epic.StatusActive,
		CreatedAt:   time.Date(2025, 8, 16, 9, 0, 0, 0, time.UTC),
		Assignee:    "test_agent",
		Description: "Round trip test",
		Phases: []epic.Phase{
			{
				ID:          "RT_P1",
				Name:        "RT Phase 1",
				Status:      epic.StatusActive,
				Description: "Round trip phase",
			},
		},
		Tasks: []epic.Task{
			{
				ID:          "RT_T1",
				PhaseID:     "RT_P1",
				Name:        "RT Task 1",
				Status:      epic.StatusCompleted,
				Assignee:    "task_agent",
				Description: "Round trip task",
			},
		},
		Tests: []epic.Test{
			{
				ID:          "RT_TEST1",
				TaskID:      "RT_T1",
				Name:        "RT Test 1",
				Status:      epic.StatusCompleted,
				Description: "Round trip test",
			},
		},
	}

	// Save original
	epicPath := filepath.Join(tempDir, "roundtrip.xml")
	err := storage.SaveEpic(original, epicPath)
	require.NoError(t, err)

	// Load it back
	loaded, err := storage.LoadEpic(epicPath)
	require.NoError(t, err)

	// Save it again
	epicPath2 := filepath.Join(tempDir, "roundtrip2.xml")
	err = storage.SaveEpic(loaded, epicPath2)
	require.NoError(t, err)

	// Load second version
	loaded2, err := storage.LoadEpic(epicPath2)
	require.NoError(t, err)

	// Verify all data is preserved
	assert.Equal(t, original.ID, loaded2.ID)
	assert.Equal(t, original.Name, loaded2.Name)
	assert.Equal(t, original.Status, loaded2.Status)
	assert.Equal(t, original.Assignee, loaded2.Assignee)
	assert.Equal(t, original.Description, loaded2.Description)

	require.Len(t, loaded2.Phases, len(original.Phases))
	for i, phase := range original.Phases {
		assert.Equal(t, phase.ID, loaded2.Phases[i].ID)
		assert.Equal(t, phase.Name, loaded2.Phases[i].Name)
		assert.Equal(t, phase.Status, loaded2.Phases[i].Status)
		assert.Equal(t, phase.Description, loaded2.Phases[i].Description)
	}

	require.Len(t, loaded2.Tasks, len(original.Tasks))
	for i, task := range original.Tasks {
		assert.Equal(t, task.ID, loaded2.Tasks[i].ID)
		assert.Equal(t, task.PhaseID, loaded2.Tasks[i].PhaseID)
		assert.Equal(t, task.Name, loaded2.Tasks[i].Name)
		assert.Equal(t, task.Status, loaded2.Tasks[i].Status)
		assert.Equal(t, task.Assignee, loaded2.Tasks[i].Assignee)
		assert.Equal(t, task.Description, loaded2.Tasks[i].Description)
	}

	require.Len(t, loaded2.Tests, len(original.Tests))
	for i, test := range original.Tests {
		assert.Equal(t, test.ID, loaded2.Tests[i].ID)
		assert.Equal(t, test.TaskID, loaded2.Tests[i].TaskID)
		assert.Equal(t, test.Name, loaded2.Tests[i].Name)
		assert.Equal(t, test.Status, loaded2.Tests[i].Status)
		assert.Equal(t, test.Description, loaded2.Tests[i].Description)
	}
}
