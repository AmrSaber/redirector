package config

import (
	"io"
	"net/http"
	"os"

	"github.com/AmrSaber/redirector/src/lib/active"
	"github.com/AmrSaber/redirector/src/lib/logger"
	"github.com/AmrSaber/redirector/src/models"
)

type ConfigManager struct {
	config models.Config
	active *active.ActiveObject
}

func NewConfigManager(source, uri string) *ConfigManager {
	manager := &ConfigManager{
		config: *models.NewConfig(source, uri),
		active: active.NewActiveObject(1024),
	}

	manager.active.Start()

	return manager
}

func (manager *ConfigManager) Close() {
	manager.active.Close()
}

func (manager *ConfigManager) LoadConfig() error {
	return active.RunCommandSync(
		manager.active,
		func() error { return manager.loadConfigUnsafe() },
	)
}

// Gets the redirection that matches the given domain
func (manager *ConfigManager) GetRedirect(domain string) *models.Redirect {
	return active.RunCommandSync(
		manager.active,
		func() *models.Redirect {
			if manager.config.Source == models.SOURCE_URL {
				if manager.config.IsStale() {
					manager.loadConfigUnsafe()
				}

				if manager.config.UrlConfigRefresh.RemapAfterRefresh {
					manager.refreshConfigUnsafe(domain)
				} else {
					manager.active.DispatchCommand(func() { manager.refreshConfigUnsafe(domain) })
				}
			}

			return manager.matchRedirect(domain)
		},
	)
}

func (manager *ConfigManager) matchRedirect(domain string) *models.Redirect {
	return matchDomain(domain, manager.config.Redirects, func(r models.Redirect) string { return r.From })
}

func (manager *ConfigManager) matchRefreshDomain(domain string) *models.RefreshDomain {
	return matchDomain(
		domain,
		manager.config.UrlConfigRefresh.RefreshDomains,
		func(d models.RefreshDomain) string { return d.Domain },
	)
}

func (manager *ConfigManager) refreshConfigUnsafe(domain string) {
	if manager.config.Source != models.SOURCE_URL {
		return
	}

	matchedRefreshDomain := manager.matchRefreshDomain(domain)
	matchedRedirect := manager.matchRedirect(domain)

	if matchedRefreshDomain != nil {
		if matchedRefreshDomain.RefreshOn == models.REFRESH_ON_HIT && matchedRedirect != nil {
			logger.Std.Printf("Refreshing config due to match with refresh domain %q and a redirect was found", domain)
			manager.loadConfigUnsafe()
			return
		}

		if matchedRefreshDomain.RefreshOn == models.REFRESH_ON_MISS && matchedRedirect == nil {
			logger.Std.Printf("Refreshing config due to match with refresh domain %q and no redirect was found", domain)
			manager.loadConfigUnsafe()
			return
		}
	}

	// Refresh config if refresh-on-hit is set and a redirect was found
	if manager.config.UrlConfigRefresh.RefreshOnHit && matchedRedirect != nil {
		logger.Std.Printf("Refreshing config due to refresh-on-hit and a redirect was found")
		manager.loadConfigUnsafe()
		return
	}

	// Refresh config if refresh-on-miss is set and no redirect was found
	if manager.config.UrlConfigRefresh.RefreshOnMiss && matchedRedirect == nil {
		logger.Std.Printf("Refreshing config due to refresh-on-miss and no redirect was found")
		manager.loadConfigUnsafe()
		return
	}
}

func (manager *ConfigManager) loadConfigUnsafe() error {
	var yamlBody []byte
	var err error

	switch manager.config.Source {
	case models.SOURCE_STDIN:
		yamlBody, err = io.ReadAll(os.Stdin)
		if err != nil {
			return err
		}

	case models.SOURCE_FILE:
		yamlBody, err = os.ReadFile(manager.config.ConfigURI)
		if err != nil {
			return err
		}

	case models.SOURCE_URL:
		res, err := http.Get(manager.config.ConfigURI)
		if err != nil {
			return err
		}

		yamlBody, err = io.ReadAll(res.Body)
		if err != nil {
			return err
		}

		res.Body.Close()
	}

	err = manager.config.Load(yamlBody)
	return err
}

func (manager *ConfigManager) GetPort() int {
	return active.RunCommandSync(
		manager.active,
		func() int { return manager.config.Port },
	)
}

func (manager *ConfigManager) GetStringConfig() string {
	return active.RunCommandSync(
		manager.active,
		func() string { return manager.config.String() },
	)
}
