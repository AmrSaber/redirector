package server

import (
	"context"
	"log"
	"net/http"
	"path"

	"github.com/AmrSaber/redirector/src/config"
)

func SetupServer(ctx context.Context, configs *config.Config) *http.Server {
	server := http.Server{
		Addr:    ":8080",
		Handler: getRedirectionMux(configs),
	}

	// Stop server on context cancel
	go func() {
		<-ctx.Done()

		log.Println("Stopping server...")
		server.Shutdown(context.Background())
		log.Println("Server stopped")
	}()

	return &server
}

func getRedirectionMux(configs *config.Config) http.Handler {
	handler := http.NewServeMux()

	handler.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		redirectInfo := configs.GetRedirect(r.Host)

		requestPath := path.Join(r.Host, r.URL.Path)

		if redirectInfo == nil {
			log.Printf("Received request for unknown host: %s", requestPath)

			// No redirects found, report 404
			w.WriteHeader(http.StatusNotFound)
			w.Header().Add("Content-Type", "application/json")
			w.Write([]byte(`{ "message": "could not match host to any redirect rule" }`))
			return
		}

		redirectPath := redirectInfo.To

		if redirectInfo.PreservePath {
			redirectPath = path.Join(redirectPath, r.URL.Path)
		}

		log.Printf("Redirecting %s to %s", requestPath, redirectPath)

		http.Redirect(w, r, redirectPath, http.StatusPermanentRedirect)
	})

	return handler
}