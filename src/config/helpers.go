package config

import (
	"regexp"
	"strings"

	"github.com/AmrSaber/redirector/src/logger"
)

// Returns a pointer to the element of the list that matched the domain after mapping it with the given mapper
func matchDomain[T any](domain string, list []T, mapper func(T) string) *T {
	domain = regexp.MustCompile(`(:\d+)?$`).ReplaceAllString(domain, "")
	domainParts := strings.Split(domain, ".")

	// Try to find exact match
	for i, item := range list {
		if mapper(item) == domain {
			return &list[i]
		}
	}

	// Try to find wildcard match
	for i, item := range list {
		if !strings.Contains(mapper(item), "*") {
			continue
		}

		fromParts := strings.Split(mapper(item), ".")
		if len(fromParts) != len(domainParts) {
			continue
		}

		isMatch := true
		for i, part := range fromParts {
			if part == "*" {
				continue
			}

			if part != domainParts[i] {
				isMatch = false
				break
			}
		}

		if isMatch {
			return &list[i]
		}
	}

	return nil
}

func refreshConfig(c *Config) {
	if err := c.Load(); err != nil {
		logger.Err.Printf("Could not refresh config from URL: %s", err)
	} else {
		logger.Std.Printf("Refreshed config from URL, new config:\n\n%s\n", c)
	}
}
