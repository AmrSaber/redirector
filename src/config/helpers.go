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

func reloadConfig(c *Config) {
	if err := c.Load(); err != nil {
		logger.Err.Printf("Could not refresh config from URL: %s", err)
	} else {
		logger.Std.Printf("Refreshed config from URL, new config:\n\n%s\n", c)
	}
}

// Reloads the config if a reload is needed, and returns a boolean whether or not a reload happened
func refreshConfig(c *Config, domain string) bool {
	if c.Source != SOURCE_URL {
		return false
	}

	matchedRefreshDomain := matchDomain(domain, c.UrlConfigRefresh.RefreshDomains, func(d RefreshDomain) string { return d.Domain })
	matchedRedirect := matchDomain(domain, c.Redirects, func(r Redirect) string { return r.From })

	if matchedRefreshDomain != nil {
		if matchedRefreshDomain.RefreshOn == "hit" && matchedRedirect != nil {
			logger.Std.Printf("Refreshing config due to match with refresh domain %q and a redirect was found", matchedRefreshDomain.Domain)
			reloadConfig(c)
			return true
		}

		if matchedRefreshDomain.RefreshOn == "miss" && matchedRedirect == nil {
			logger.Std.Printf("Refreshing config due to match with refresh domain %q and no redirect was found", matchedRefreshDomain.Domain)
			reloadConfig(c)
			return true
		}
	}

	// Refresh config if refresh-on-hit is set and a redirect was found
	if c.UrlConfigRefresh.RefreshOnHit && matchedRedirect != nil {
		logger.Std.Printf("Refreshing config due to refresh-on-hit and a redirect was found")
		reloadConfig(c)
		return true
	}

	// Refresh config if refresh-on-miss is set and no redirect was found
	if c.UrlConfigRefresh.RefreshOnMiss && matchedRedirect == nil {
		logger.Std.Printf("Refreshing config due to refresh-on-miss and no redirect was found")
		reloadConfig(c)
		return true
	}

	return false
}
