package rediscomp

import (
	"context"
	"sync"
	"testing"
)

// Verifies that concurrent callers of DB() share one client per DB and do not
// race on the internal map. Run with -race to be meaningful.
func TestDBConcurrent(t *testing.T) {
	r := &Redis{Addr: "localhost:6379"}
	if err := r.Init(context.Background()); err != nil {
		t.Fatalf("Init: %v", err)
	}

	clients := make([]any, 50)
	var wg sync.WaitGroup
	for i := range clients {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			clients[i] = r.DB(i % 4)
		}(i)
	}
	wg.Wait()

	for i, got := range clients {
		if again := r.DB(i % 4); again != got {
			t.Fatalf("DB(%d) returned a different client on the second call", i%4)
		}
	}

	if got := len(r.clients()); got != 4 {
		t.Fatalf("open clients = %d, want 4", got)
	}
}
