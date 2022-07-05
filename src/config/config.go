package config

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	SOURCE_FILE = "source-file"
	SOURCE_URL  = "source-url"
)

type Config struct {
	source    string `yaml:"-"`
	configURI string `yaml:"-"`

	CacheTTL time.Duration `yaml:"cache-ttl"`
	LoadedAt time.Time     `yaml:"-"`

	Redirects []Redirect `yaml:"redirects"`
}

type Redirect struct {
	From         string `yaml:"from"`
	To           string `yaml:"to"`
	PreservePath bool   `yaml:"preserve-path"`
}

func NewConfig(source, uri string) *Config {
	return &Config{
		source:    source,
		configURI: uri,
	}
}

func (c *Config) Load() error {
	var yamlBody []byte
	var err error

	switch c.source {
	case SOURCE_FILE:
		yamlBody, err = ioutil.ReadFile(c.configURI)
		if err != nil {
			return err
		}

	case SOURCE_URL:
		res, err := http.Get(c.configURI)
		if err != nil {
			log.Fatal(err)
		}

		yamlBody, err = ioutil.ReadAll(res.Body)
		if err != nil {
			log.Fatal(err)
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

	if c.CacheTTL == 0 {
		c.CacheTTL, _ = time.ParseDuration("4h")
	}

	c.CopyFrom(&parsedConfig)
	c.LoadedAt = time.Now()

	return nil
}

// Validates the config
func (c *Config) Validate() error {
	errors := []string{}

	// Validate that each "from" is a valid domain name and each "to" is a valid URL
	domainRegex := regexp.MustCompile(`^(?:[a-zA-Z0-9-_]+|\*)(?:\.(?:[a-zA-Z0-9-_]+|\*))+(?::\d+)?$`)
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
	if c.source == SOURCE_URL && time.Since(c.LoadedAt) >= c.CacheTTL {
		if err := c.Load(); err != nil {
			log.Printf("Could not refresh config from URL: %s", err)
		}
	}

	hostParts := strings.Split(host, ".")

	for _, r := range c.Redirects {
		isMatch := false

		if r.From == host {
			isMatch = true
		} else if strings.Contains(r.From, "*") {
			fromParts := strings.Split(r.From, ".")

			if len(fromParts) == len(hostParts) {
				isMatch = true
				for i, part := range fromParts {
					if part != "*" && part != hostParts[i] {
						isMatch = false
						break
					}
				}
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
	c.Redirects = other.Redirects
}

// Prints the config as yaml
func (c Config) String() string {
	out, _ := yaml.Marshal(c)
	return string(out)
}
