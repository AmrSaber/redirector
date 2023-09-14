package server

import (
	"crypto/sha256"
	"crypto/subtle"
	"fmt"
	"net/http"
	"path"

	"github.com/AmrSaber/redirector/src/config"
	"github.com/AmrSaber/redirector/src/lib/logger"
)

func SetupServer(configs *config.Config) *http.Server {
	server := http.Server{
		Addr:    fmt.Sprintf(":%d", configs.Port),
		Handler: getRedirectionMux(configs),
	}

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

		// If user is not authorized, prompt for basic auth and return UNAUTHORIZED
		if !isAuthorized(r, redirectInfo) {
			w.Header().Set("WWW-Authenticate", fmt.Sprintf(`Basic realm="%s", charset="UTF-8"`, redirectInfo.Auth.Realm))
			http.Error(w, "Unauthorized", http.StatusUnauthorized)

			logger.Std.Printf("Received unauthorized request for host: %s", requestPath)
			return
		}

		redirectPath := redirectInfo.ResolvePath(r)

		logger.Std.Printf("Redirecting %q to %q", requestPath, redirectPath)

		status := http.StatusPermanentRedirect
		if *redirectInfo.TempRedirect {
			status = http.StatusTemporaryRedirect
		}

		http.Redirect(w, r, redirectPath, status)
	})

	return handler
}

func isAuthorized(r *http.Request, redirectInfo *config.Redirect) bool {
	if redirectInfo.Auth == nil {
		return true
	}

	username, password, ok := r.BasicAuth()
	if !ok {
		return false
	}

	// Hashes are used to perform const-time password check
	usernameHash := sha256.Sum256([]byte(username))
	passwordHash := sha256.Sum256([]byte(password))
	expectedUsernameHash := sha256.Sum256([]byte(redirectInfo.Auth.Username))
	expectedPasswordHash := sha256.Sum256([]byte(redirectInfo.Auth.Password))

	usernameMatch := (subtle.ConstantTimeCompare(usernameHash[:], expectedUsernameHash[:]) == 1)
	passwordMatch := (subtle.ConstantTimeCompare(passwordHash[:], expectedPasswordHash[:]) == 1)

	return usernameMatch && passwordMatch
}
