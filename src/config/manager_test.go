package config

import (
	"sync"
	"testing"

	"github.com/AmrSaber/redirector/src/models"
)

func BenchmarkGetRedirect(b *testing.B) {
	manager := NewConfigManager("MANUAL_UPDATE", "")
	defer manager.Close()
	var wg sync.WaitGroup

	manager.config.Redirects = []models.Redirect{
		{
			From: "a.b.c",
			To:   "http://x.y.z",
		},
		{
			From: "x.y.z",
			To:   "http://a.b.c",
		},
	}

	for n := 0; n < b.N; n++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			manager.GetRedirect("a.b.c")
		}()
	}

	wg.Wait()
}
