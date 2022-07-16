package config

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	SOURCE_STDIN = "@source:stdin"
	SOURCE_FILE  = "@source:file"
	SOURCE_URL   = "@source:url"
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

// TODO: figure out a way to test the execution of these options
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

type Redirect struct {
	From         string        `yaml:"from"`
	To           string        `yaml:"to"`
	PreservePath bool          `yaml:"preserve-path"`
	TempRedirect *bool         `yaml:"temp-redirect"`
	Auth         *RedirectAuth `yaml:"auth,omitempty"`
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

func (c *Config) Load() error {
	var yamlBody []byte
	var err error

	switch c.Source {
	case SOURCE_STDIN:
		yamlBody, err = ioutil.ReadAll(os.Stdin)
		if err != nil {
			return err
		}

	case SOURCE_FILE:
		yamlBody, err = ioutil.ReadFile(c.ConfigURI)
		if err != nil {
			return err
		}

	case SOURCE_URL:
		res, err := http.Get(c.ConfigURI)
		if err != nil {
			return nil
		}

		yamlBody, err = ioutil.ReadAll(res.Body)
		if err != nil {
			return nil
		}

		res.Body.Close()
	}

	var parsedConfig Config
	if err := yaml.Unmarshal(yamlBody, &parsedConfig); err != nil {
		return fmt.Errorf("could not parse configs from yaml: %s", err)
	}

	if err := parsedConfig.Validate(); err != nil {
		return fmt.Errorf("invalid configurations:\n%s", err)
	}

	c.CopyFrom(&parsedConfig)
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
func (c *Config) Validate() error {
	errors := []string{}

	// Validate that each "from" is a valid domain name and each "to" is a valid URL
	domainRegex := regexp.MustCompile(`^(?:[a-zA-Z0-9-_]+|\*)(?:\.(?:[a-zA-Z0-9-_]+|\*))+$`)
	urlRegex := regexp.MustCompile(`^\w+://[a-zA-Z0-9-_]+(?:\.[a-zA-Z0-9-_]+)+(?:/[^/]*)*$`)
	hasPathRegex := regexp.MustCompile(`^.+//.+(?:/[^/]*)+$`)

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

		if d.RefreshOn != "hit" && d.RefreshOn != "miss" {
			errors = append(errors, fmt.Sprintf(`Invalid "refresh-on" for refresh domains [#%d]: %s`, i, d.RefreshOn))
		}
	}

	if len(errors) != 0 {
		return fmt.Errorf(strings.Join(errors, "\n"))
	} else {
		return nil
	}
}

// Gets the redirection that matches the given domain
func (c *Config) GetRedirect(domain string) *Redirect {
	// Refresh the config if it's stale
	if c.Source == SOURCE_URL && time.Since(c.LoadedAt) >= c.UrlConfigRefresh.CacheTTL {
		reloadConfig(c)
	}

	matchedRedirect := matchDomain(domain, c.Redirects, func(r Redirect) string { return r.From })

	// Perform a refresh in the background if it's needed
	if c.UrlConfigRefresh.RemapAfterRefresh {
		if refreshed := refreshConfig(c, domain); refreshed {
			matchedRedirect = matchDomain(domain, c.Redirects, func(r Redirect) string { return r.From })
		}
	} else {
		go refreshConfig(c, domain)
	}

	return matchedRedirect
}

// Copy configurations from another config
func (c *Config) CopyFrom(other *Config) {
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
