package models

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Source    string    `yaml:"source"`
	ConfigURI string    `yaml:"config-uri"`
	LoadedAt  time.Time `yaml:"loaded-at"`

	Port int `yaml:"port"`

	TempRedirect bool `yaml:"temp-redirect"`

	UrlConfigRefresh UrlRefreshOptions `yaml:"url-config-refresh"`

	Redirects []Redirect `yaml:"redirects"`
}

type UrlRefreshOptions struct {
	// How often to refresh the URL
	CacheTTL time.Duration `yaml:"cache-ttl"`

	// Whether or not to refresh on domain hit
	RefreshOnHit bool `yaml:"refresh-on-hit"`

	// Whether or not to refresh on domain miss (domain not found)
	RefreshOnMiss bool `yaml:"refresh-on-miss"`

	// Whether or not to remap the request after refreshing
	RemapAfterRefresh bool `yaml:"remap-after-refresh"`

	// Domains to refresh on
	RefreshDomains []RefreshDomain `yaml:"refresh-domains"`
}

type RefreshDomain struct {
	Domain string `yaml:"domain"`

	// Whether to refresh on domain hit or miss
	RefreshOn string `yaml:"refresh-on"`
}

type RedirectAuth struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Realm    string `yaml:"realm"`
}

func NewConfig(source, uri string) *Config {
	return &Config{
		Source:    source,
		ConfigURI: uri,
	}
}

func (c *Config) Load(yamlBody []byte) error {
	var parsedConfig Config
	if err := yaml.Unmarshal(yamlBody, &parsedConfig); err != nil {
		return fmt.Errorf("could not parse configs from yaml: %s", err)
	}

	if err := parsedConfig.validate(); err != nil {
		return fmt.Errorf("invalid configurations:\n%s", err)
	}

	c.copyFrom(&parsedConfig)
	c.LoadedAt = time.Now()

	if c.UrlConfigRefresh.CacheTTL == 0 {
		c.UrlConfigRefresh.CacheTTL, _ = time.ParseDuration("6h")
	}

	if c.Port == 0 {
		c.Port = 8080
	}

	for i, r := range c.Redirects {
		if r.TempRedirect == nil {
			c.Redirects[i].TempRedirect = &c.TempRedirect
		}

		if r.Auth != nil && r.Auth.Realm == "" {
			c.Redirects[i].Auth.Realm = "Restricted"
		}
	}

	return nil
}

// Validates the config
func (c Config) validate() error {
	errors := []string{}

	// Validate that each "from" is a valid domain name and each "to" is a valid URL

	for i, r := range c.Redirects {
		// Trim trailing slash from each domain
		r.From = strings.TrimSuffix(r.From, "/")
		r.To = strings.TrimSuffix(r.To, "/")

		if !domainRegex.MatchString(r.From) {
			errors = append(errors, fmt.Sprintf(`Invalid "from" domain [#%d]: %s`, i, r.From))
		}

		if !urlRegex.MatchString(r.To) {
			errors = append(errors, fmt.Sprintf(`Invalid "to" URL [#%d]: %s`, i, r.To))
		}

		if r.PreservePath && hasPathRegex.MatchString(r.To) {
			errors = append(errors, fmt.Sprintf(`"To" URL cannot contain path and set preserve path [#%d]: %s`, i, r.To))
		}

		if toWildcardsCount := strings.Count(r.To, "*"); toWildcardsCount > 0 {
			toUrl, _ := url.Parse(r.To)

			toSectionsCount := len(strings.Split(toUrl.Host, "."))
			fromSectionsCount := len(strings.Split(r.From, "."))

			if toSectionsCount != fromSectionsCount {
				errors = append(
					errors,
					fmt.Sprintf(
						`"to" has wildcard(s) but "To" sections (found %d) and "From" sections (found %d) don't match `,
						toSectionsCount,
						fromSectionsCount,
					),
				)
			}

		}

		if r.Auth != nil && r.Auth.Username == "" {
			errors = append(errors, fmt.Sprintf(`Auth "username" must be provided [#%d]`, i))
		}

		if r.Auth != nil && r.Auth.Password == "" {
			errors = append(errors, fmt.Sprintf(`Auth "password" must be provided [#%d]`, i))
		}
	}

	for i, d := range c.UrlConfigRefresh.RefreshDomains {
		if !domainRegex.MatchString(d.Domain) {
			errors = append(errors, fmt.Sprintf(`Invalid "domain" for refresh domains [#%d]: %s`, i, d.Domain))
		}

		if d.RefreshOn != REFRESH_ON_HIT && d.RefreshOn != REFRESH_ON_MISS {
			errors = append(errors, fmt.Sprintf(`Invalid "refresh-on" for refresh domains [#%d]: %s`, i, d.RefreshOn))
		}
	}

	if len(errors) != 0 {
		return fmt.Errorf(strings.Join(errors, "\n"))
	} else {
		return nil
	}
}

func (c Config) IsStale() bool {
	return c.Source == SOURCE_URL && time.Since(c.LoadedAt) >= c.UrlConfigRefresh.CacheTTL
}

// Copy configurations from another config
func (c *Config) copyFrom(other *Config) {
	c.Port = other.Port
	c.TempRedirect = other.TempRedirect

	c.UrlConfigRefresh = other.UrlConfigRefresh
	c.Redirects = other.Redirects
}

// Prints the config as yaml
func (c Config) String() string {
	out, _ := yaml.Marshal(c)
	return string(out)
}
