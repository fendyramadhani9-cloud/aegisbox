package sandbox_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/aegisbox/aegisbox/internal/sandbox"
)

func TestCleanupTracker_DirectoryRemoval(t *testing.T) {
	tracker := sandbox.NewCleanupTracker()

	tmpDir, err := os.MkdirTemp("", "aegisbox-cleanup-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("hello"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	tracker.TrackDir(tmpDir)

	errs := tracker.CleanupAll()
	if len(errs) > 0 {
		t.Fatalf("cleanup returned errors: %v", errs)
	}

	if _, err := os.Stat(tmpDir); !os.IsNotExist(err) {
		t.Fatalf("expected directory %s to be deleted by cleanup tracker", tmpDir)
	}
}

func TestCleanupTracker_CustomFunctions(t *testing.T) {
	tracker := sandbox.NewCleanupTracker()

	executed := false
	tracker.TrackFunc(func() error {
		executed = true
		return nil
	})

	errs := tracker.CleanupAll()
	if len(errs) > 0 {
		t.Fatalf("unexpected cleanup errors: %v", errs)
	}

	if !executed {
		t.Fatalf("expected custom cleanup function to be executed")
	}
}
