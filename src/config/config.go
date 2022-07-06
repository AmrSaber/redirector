package config

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/AmrSaber/redirector/src/logger"
	"gopkg.in/yaml.v3"
)

const (
	SOURCE_FILE = "@source:file"
	SOURCE_URL  = "@source:url"
)

type Config struct {
	Source    string `yaml:"source"`
	ConfigURI string `yaml:"config-uri"`

	CacheTTL time.Duration `yaml:"cache-ttl"`
	LoadedAt time.Time     `yaml:"loaded-at"`

	Port int `yaml:"port"`

	Redirects []Redirect `yaml:"redirects"`
}

type Redirect struct {
	From         string `yaml:"from"`
	To           string `yaml:"to"`
	PreservePath bool   `yaml:"preserve-path"`
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

	if c.CacheTTL == 0 {
		c.CacheTTL, _ = time.ParseDuration("4h")
	}

	if c.Port == 0 {
		c.Port = 8080
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
	}

	if len(errors) != 0 {
		return fmt.Errorf(strings.Join(errors, "\n"))
	} else {
		return nil
	}
}

// Gets the redirection that matches the given domain
func (c *Config) GetRedirect(host string) *Redirect {
	// Refresh the config if it's stale
	if c.Source == SOURCE_URL && time.Since(c.LoadedAt) >= c.CacheTTL {
		if err := c.Load(); err != nil {
			logger.Err.Printf("Could not refresh config from URL: %s", err)
		} else {
			logger.Std.Printf("Refreshed config from URL, new config:\n\n%s\n", c)
		}
	}

	// Remove port from host if it's present
	host = regexp.MustCompile(`(:\d+)?$`).ReplaceAllString(host, "")
	hostParts := strings.Split(host, ".")

	// Try to find exact match
	for _, r := range c.Redirects {
		if r.From == host {
			return &r
		}
	}

	// Try to find wildcard match
	for _, r := range c.Redirects {
		if !strings.Contains(r.From, "*") {
			continue
		}

		fromParts := strings.Split(r.From, ".")
		if len(fromParts) != len(hostParts) {
			continue
		}

		isMatch := true
		for i, part := range fromParts {
			if part == "*" {
				continue
			}

			if part != hostParts[i] {
				isMatch = false
				break
			}
		}

		if isMatch {
			return &r
		}
	}

	return nil
}

// Copy configurations from another config
func (c *Config) CopyFrom(other *Config) {
	c.CacheTTL = other.CacheTTL
	c.Port = other.Port
	c.Redirects = other.Redirects
}

// Prints the config as yaml
func (c Config) String() string {
	out, _ := yaml.Marshal(c)
	return string(out)
}
