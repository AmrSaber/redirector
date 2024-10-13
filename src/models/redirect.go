package models

import (
	"net/http"
	"net/url"
	"strings"
)

type Redirect struct {
	From         string        `yaml:"from"`
	To           string        `yaml:"to"`
	PreservePath bool          `yaml:"preserve-path"`
	TempRedirect *bool         `yaml:"temp-redirect"`
	Auth         *RedirectAuth `yaml:"auth,omitempty"`
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
