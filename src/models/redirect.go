package models

import (
	"crypto/sha256"
	"crypto/subtle"
	"net/http"
	"net/url"
	"strings"
)

type Redirect struct {
	From         string     `yaml:"from"`
	To           string     `yaml:"to"`
	PreservePath bool       `yaml:"preserve-path"`
	TempRedirect *bool      `yaml:"temp-redirect"`
	AuthNames    []string   `yaml:"auth,omitempty"`
	ActualAuths  AuthSchema `yaml:"-"`
}

func (redirect Redirect) ResolvePath(request *http.Request) string {
	toUrl, _ := url.Parse(redirect.To)

	if strings.Contains(toUrl.Host, "*") {
		toSections := strings.Split(toUrl.Host, ".")
		requestSections := strings.Split(request.Host, ".")

		// Substitute every * in to with corresponding section in request
		for i, section := range toSections {
			if section != "*" {
				continue
			}

			toSections[i] = requestSections[i]
		}

		toUrl.Host = strings.Join(toSections, ".")
	}

	if redirect.PreservePath {
		toUrl.Path = request.URL.Path
	}

	return toUrl.String()
}

func (redirect Redirect) GetBasicAuthRealm() string {
	if len(redirect.AuthNames) == 0 {
		return ""
	}

	authName := redirect.AuthNames[0]
	return redirect.ActualAuths.BasicAuth[authName].Realm
}

func (redirect Redirect) IsAuthorized(req *http.Request) bool {
	// Validate basic auth
	if len(redirect.ActualAuths.BasicAuth) > 0 {
		return redirect.authorizeBasicAuth(req)
	}

	// More kinds of auth to be added here...

	return true
}

func (redirect Redirect) authorizeBasicAuth(req *http.Request) bool {
	reqUsername, reqPassword, ok := req.BasicAuth()
	if !ok {
		return false
	}

	// Find the matching auth block based on user
	matchingBasicAuth := redirect.findMatchingBasicAuth(req)
	if matchingBasicAuth == nil {
		return false
	}

	userConfig := matchingBasicAuth.FindMatchingUser(reqUsername)

	// Hashes are used to perform const-time password check
	passwordHash := sha256.Sum256([]byte(reqPassword))
	expectedPasswordHash := sha256.Sum256([]byte(userConfig.Password))

	return (subtle.ConstantTimeCompare(passwordHash[:], expectedPasswordHash[:]) == 1)
}

func (redirect Redirect) findMatchingBasicAuth(req *http.Request) *BasicAuthSchema {
	reqUsername, _, ok := req.BasicAuth()
	if !ok {
		return nil
	}

	for _, auth := range redirect.ActualAuths.BasicAuth {
		for _, userConfig := range auth.Users {
			if userConfig.Username == reqUsername {
				return auth
			}
		}
	}

	return nil
}
