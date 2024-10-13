package servers

import (
	"context"
	"fmt"
	"net/http"
	"path"

	"github.com/AmrSaber/redirector/src/config"
	"github.com/AmrSaber/redirector/src/lib/logger"
)

func StartHttpServer(ctx context.Context, configManager *config.ConfigManager) <-chan error {
	doneChan := make(chan error)

	go func() {
		defer close(doneChan)

		server := http.Server{
			Addr:    fmt.Sprintf(":%d", configManager.GetPort()),
			Handler: getRedirectionMux(configManager),
		}

		// Close server on end of context
		go func() {
			<-ctx.Done()

			logger.Std.Println("Stopping HTTP server...")
			_ = server.Shutdown(context.Background())
			logger.Std.Println("HTTP server stopped")
		}()

		logger.Std.Printf("Server listening on http://localhost:%d\n", configManager.GetPort())
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			doneChan <- fmt.Errorf("could not start http server: %w", err)
		}
	}()

	return doneChan
}

func getRedirectionMux(configs *config.ConfigManager) http.Handler {
	handler := http.NewServeMux()

	handler.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		redirectInfo := configs.GetRedirect(req.Host)

		requestPath := path.Join(req.Host, req.URL.Path)

		if redirectInfo == nil {
			logger.Std.Printf("Received request for unknown host: %s", requestPath)

			// No redirects found, report 404
			res.WriteHeader(http.StatusNotFound)
			res.Header().Add("Content-Type", "application/json")
			res.Write([]byte(`{ "message": "could not match host to any redirection rule" }`))
			return
		}

		// If user is not authorized, prompt for basic auth and return UNAUTHORIZED
		if !redirectInfo.IsAuthorized(req) {
			res.Header().Set("WWW-Authenticate", fmt.Sprintf(`Basic realm="%s", charset="UTF-8"`, redirectInfo.GetBasicAuthRealm()))
			http.Error(res, "Unauthorized", http.StatusUnauthorized)

			logger.Std.Printf("Received unauthorized request for host: %s", requestPath)
			return
		}

		redirectPath := redirectInfo.ResolvePath(req)

		logger.Std.Printf("Redirecting %q to %q", requestPath, redirectPath)

		status := http.StatusPermanentRedirect
		if *redirectInfo.TempRedirect {
			status = http.StatusTemporaryRedirect
		}

		http.Redirect(res, req, redirectPath, status)
	})

	return handler
}
