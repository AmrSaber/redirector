package config

import (
	"fmt"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Redirects []Redirect `yaml:"redirects"`
}

type Redirect struct {
	From         string `yaml:"from"`
	To           string `yaml:"to"`
	PreservePath bool   `yaml:"preserve-path"`
}

func ConfigFromYaml(yamlFile []byte) (*Config, error) {
	var config Config

	if err := yaml.Unmarshal(yamlFile, &config); err != nil {
		return nil, fmt.Errorf("could not parse configs from yaml: %s", err)
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configurations:\n%s", err)
	}

	return &config, nil
}

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

func (c *Config) GetRedirect(host string) *Redirect {
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

func (c Config) String() string {
	out, _ := yaml.Marshal(c)
	return string(out)
}
