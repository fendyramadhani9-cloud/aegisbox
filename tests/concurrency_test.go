package tests

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/aegisbox/aegisbox/internal/executor"
)

func TestExecution_ConcurrentRequests(t *testing.T) {
	mgr := newTestExecutionManager(t)

	concurrency := 6
	var wg sync.WaitGroup
	results := make([]*executor.ExecutionResult, concurrency)
	seenIDs := make(map[string]bool)
	var mu sync.Mutex

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			req := &executor.ExecutionRequest{
				Language:  "python",
				Code:      fmt.Sprintf("print('concurrency-test-%d')", idx),
				TimeoutMs: 3000,
				MaxMemMB:  64,
			}
			res := mgr.Execute(context.Background(), req)
			results[idx] = res

			mu.Lock()
			seenIDs[res.ExecutionID] = true
			mu.Unlock()
		}(i)
	}

	wg.Wait()

	for i, res := range results {
		if res == nil {
			t.Fatalf("result %d was nil", i)
		}
		if res.Status != executor.StatusCompleted {
			t.Errorf("worker %d failed with status: %s (err: %s)", i, res.Status, res.ErrorMessage)
		}
		expectedOut := fmt.Sprintf("concurrency-test-%d\n", i)
		if res.Stdout != expectedOut {
			t.Errorf("worker %d expected '%s', got '%s'", i, expectedOut, res.Stdout)
		}
	}

	if len(seenIDs) != concurrency {
		t.Errorf("expected %d unique execution IDs, got %d", concurrency, len(seenIDs))
	}
}
