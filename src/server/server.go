package server

import (
	"context"
	"fmt"
	"net/http"
	"path"

	"github.com/AmrSaber/redirector/src/config"
	"github.com/AmrSaber/redirector/src/lib/logger"
)

func SetupServer(ctx context.Context, configs *config.Config) *http.Server {
	server := http.Server{
		Addr:    fmt.Sprintf(":%d", configs.Port),
		Handler: getRedirectionMux(configs),
	}

	// Stop server on context cancel
	go func() {
		<-ctx.Done()

		logger.Std.Println("Stopping server...")
		server.Shutdown(context.Background())
		logger.Std.Println("Server stopped")
	}()

	return &server
}

func getRedirectionMux(configs *config.Config) http.Handler {
	handler := http.NewServeMux()

	handler.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		redirectInfo := configs.GetRedirect(r.Host)

		requestPath := path.Join(r.Host, r.URL.Path)

		if redirectInfo == nil {
			logger.Std.Printf("Received request for unknown host: %s", requestPath)

			// No redirects found, report 404
			w.WriteHeader(http.StatusNotFound)
			w.Header().Add("Content-Type", "application/json")
			w.Write([]byte(`{ "message": "could not match host to any redirection rule" }`))
			return
		}

		redirectPath := redirectInfo.To

		if redirectInfo.PreservePath {
			redirectPath = path.Join(redirectPath, r.URL.Path)
		}

		logger.Std.Printf("Redirecting %q to %q", requestPath, redirectPath)

		status := http.StatusPermanentRedirect
		if *redirectInfo.TempRedirect {
			status = http.StatusTemporaryRedirect
		}

		http.Redirect(w, r, redirectPath, status)
	})

	return handler
}
