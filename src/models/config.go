package models

import (
	"fmt"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/AmrSaber/redirector/src/utils"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Source    string    `yaml:"source"`
	ConfigURI string    `yaml:"config-uri"`
	LoadedAt  time.Time `yaml:"loaded-at"`

	Port         int   `yaml:"port"`
	TempRedirect *bool `yaml:"temp-redirect"`

	Auth             *AuthSchema        `yaml:"auth,omitempty"`
	UrlConfigRefresh *UrlRefreshOptions `yaml:"url-config-refresh,omitempty"` // TODO make into pointer

	Redirects []Redirect `yaml:"redirects"`
}

type AuthSchema struct {
	BasicAuth map[string]*BasicAuthSchema `yaml:"basic-auth,omitempty"`
}

type BasicAuthSchema struct {
	Realm string          `yaml:"realm"`
	Users []BasicAuthUser `yaml:"users"`
}

func (auth BasicAuthSchema) FindMatchingUser(username string) *BasicAuthUser {
	for _, userConfig := range auth.Users {
		if userConfig.Username == username {
			return &userConfig
		}
	}

	return nil
}

type BasicAuthUser struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
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

var _DEFAULT_TEMP_REDIRECT = true

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

	if c.Auth != nil {
		for _, auth := range c.Auth.BasicAuth {
			if auth.Realm == "" {
				auth.Realm = utils.DEFAULT_REALM
			}
		}
	}

	if c.Source == SOURCE_URL {
		if c.UrlConfigRefresh == nil {
			c.UrlConfigRefresh = &UrlRefreshOptions{}
		}

		if c.UrlConfigRefresh.CacheTTL == 0 {
			c.UrlConfigRefresh.CacheTTL, _ = time.ParseDuration("6h")
		}
	}

	if c.Port == 0 {
		c.Port = 80
	}

	if c.TempRedirect == nil {
		c.TempRedirect = &_DEFAULT_TEMP_REDIRECT
	}

	for i, r := range c.Redirects {
		if r.TempRedirect == nil {
			r.TempRedirect = c.TempRedirect
		}

		// Add actual auth objects to redirect for simpler authentication
		if len(r.AuthNames) > 0 {
			r.ActualAuths.BasicAuth = make(map[string]*BasicAuthSchema)

			for _, authName := range r.AuthNames {
				r.ActualAuths.BasicAuth[authName] = c.Auth.BasicAuth[authName]
			}
		}

		c.Redirects[i] = r
	}

	return nil
}

// Validates the config
func (c Config) validate() error {
	errors := []string{}

	// Validate auth
	if c.Auth != nil {
		for key, auth := range c.Auth.BasicAuth {
			// Validate that each object contains a username and a password
			for i, userConfig := range auth.Users {
				if userConfig.Username == "" {
					errors = append(errors, fmt.Sprintf(`Auth "username" must be provided [@basic-auth %q #%d]`, key, i))
				}

				if userConfig.Password == "" {
					errors = append(errors, fmt.Sprintf(`Auth "password" must be provided [@basic-auth %q #%d]`, key, i))
				}
			}
		}

		// Validate that there are no duplicate usernames
		{
			authUsers := make(map[string][]string, 0)
			for key, auth := range c.Auth.BasicAuth {
				for i, userConfig := range auth.Users {
					username := userConfig.Username
					authUsers[username] = append(authUsers[username], fmt.Sprintf("%q [#%d]", key, i))
				}
			}

			for username, instances := range authUsers {
				if len(instances) > 1 {
					instances = utils.MapSlice(instances, func(value string) string { return fmt.Sprintf("  - %s", value) })

					errors = append(
						errors,
						fmt.Sprintf(
							"Found duplicate username %q in basic-auth at:\n%s",
							username,
							strings.Join(instances, "\n"),
						),
					)
				}
			}
		}
	}

	for i, r := range c.Redirects {
		// Trim trailing slash from each domain
		r.From = strings.TrimSuffix(r.From, "/")
		r.To = strings.TrimSuffix(r.To, "/")

		// Validate that each "from" is a valid domain name and each "to" is a valid URL
		if !utils.DomainRegex.MatchString(r.From) {
			errors = append(errors, fmt.Sprintf(`Invalid "from" domain [#%d]: %s`, i, r.From))
		}

		if !utils.UrlRegex.MatchString(r.To) {
			errors = append(errors, fmt.Sprintf(`Invalid "to" URL [#%d]: %s`, i, r.To))
		}

		if r.PreservePath && utils.HasPathRegex.MatchString(r.To) {
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

		if len(r.AuthNames) > 0 {
			realms := make(map[string]any)
			availableAuths := c.GetAvailableAuthNames()
			for _, authName := range r.AuthNames {
				if slices.Contains(availableAuths, authName) {
					realm := c.Auth.BasicAuth[authName].Realm
					if realm == "" {
						realm = utils.DEFAULT_REALM
					}

					realms[realm] = struct{}{}
				} else {
					errors = append(errors, fmt.Sprintf("Auth %q not found [@redirect#%d]", authName, i))
				}
			}

			// Validate that there are no mixed realms
			if len(realms) > 1 {
				foundRealms := utils.MapSlice(utils.GetMapKeys(realms), func(str string) string { return fmt.Sprintf("%q", str) })
				errors = append(
					errors,
					fmt.Sprintf(
						"Found mixed realms (%s) at redirect #%d. All linked auths must have the same realm",
						strings.Join(foundRealms, ", "), i,
					),
				)
			}
		}
	}

	if c.UrlConfigRefresh != nil {
		for i, d := range c.UrlConfigRefresh.RefreshDomains {
			if !utils.DomainRegex.MatchString(d.Domain) {
				errors = append(errors, fmt.Sprintf(`Invalid "domain" for refresh domains [#%d]: %s`, i, d.Domain))
			}

			if d.RefreshOn != REFRESH_ON_HIT && d.RefreshOn != REFRESH_ON_MISS {
				errors = append(errors, fmt.Sprintf(`Invalid "refresh-on" for refresh domains [#%d]: %s`, i, d.RefreshOn))
			}
		}
	}

	if len(errors) != 0 {
		return fmt.Errorf(strings.Join(errors, "\n"))
	} else {
		return nil
	}
}

func (c Config) GetAvailableAuthNames() []string {
	auths := make([]string, 0)
	if c.Auth != nil {
		for key := range c.Auth.BasicAuth {
			auths = append(auths, key)
		}
	}
	return auths
}

func (c Config) IsStale() bool {
	return c.Source == SOURCE_URL && time.Since(c.LoadedAt) >= c.UrlConfigRefresh.CacheTTL
}

// Copy configurations from another config
func (c *Config) copyFrom(other *Config) {
	c.Port = other.Port
	c.TempRedirect = other.TempRedirect

	c.Auth = other.Auth
	c.UrlConfigRefresh = other.UrlConfigRefresh
	c.Redirects = other.Redirects
}

// Prints the config as yaml
func (c Config) String() string {
	out, _ := yaml.Marshal(c)
	return string(out)
}
