# Redirector

Simple, light weight, configurable, special-purpose reverse proxy for handling redirection based on domain name, written in Go.

Redirector enables you to redirect your traffic from a given domain to a URL (or domain) using simple configuration. You can configure whether or not you want the redirection to be permanent (for browser caching) and whether or not to preserve the path after redirection.

Redirector also watches the configuration file for updates, so you do not need to restart the application on each configuration update (see [configuration watching](#configuration-watching) section below).

## Install

### From binaries
You can download the binaries from the [latest release](https://github.com/AmrSaber/redirector/releases) manually, or using bash

```bash
# Download binary from release
curl -o redirector -sSL https://github.com/amrsaber/redirector/releases/download/<version>/redirector_<os>_<arch>

# Give execute permission
chmod +x redirector

# Execute
./redirector start --file config.yaml
```

### Docker Image
You can use the application's docker image `ghcr.io/amrsaber/redirector` You can also download a specific version with image tag, available tags can be found [here](https://github.com/AmrSaber/redirector/pkgs/container/redirector).

For example

```bash
docker run -p 8080:8080 ghcr.io/amrsaber/redirector:latest start --url https://url-to-config.com
```

Or using a file

```bash
docker run -p 8080:8080 -v /path/to/config.yaml:/config.yaml ghcr.io/amrsaber/redirector:latest start --file /config.yaml
```

## Usage

### Commands
- `start`: starts the server, see more details below
- `stop`: stops the server if it's running and returns "OK", otherwise returns error
- `ping`: pings the server to make sure it's running and healthy, returns "PONG" if server is running, otherwise returns error

To view commands and their documentation and flags, start the application with `--help`, `-h`, `help`, `h`, or without any commands. And you can use `--help` or `-h` with any command to view more details about it.

### Loading Configuration
You can load configuration from different sources:

- From file using `--file` flag then file name, e.g. `redirector --file config.yaml`
- From URL using `--url` flag then URL, e.g. `redirector --url https://some-url.com/path/to/config`
- From URL using `CONFIG_URL` environment variable: just set the env variable and start the application, e.g. `CONFIG_URL=https//some-url.com/some/path redirector`
- From STDIN using `--stdin` flag, e.g. `cat config.yaml | redirector --stdin`

The application will print the final form of the parsed configuration after parsing them, the configuration will be validated for the right type and schema, any additional fields will be ignored.

To only print the parsed configuration provide the flag `--dry-run`, e.g. `redirector --file config.yaml --dry-run`.

In case you provide more that 1 source, the precedence is as follows: stdin, file, url, env variable.

#### Configuration Watching
In case of providing the configuration from a file, the application will attempt to watch the file for changes and update the configuration automatically with after each change, if the file became invalid after an update, the application will keep the last valid parsed configuration.

In case of providing the configuration from a URL (using `--url` flag or the env variable), the application will attempt to refresh the configuration from the URL after `cache-ttl` time specified in the configuration file. If the application fails to refresh after the specified cache-ttl time (invalid config format, network problem, etc...) it will keep the last valid parsed configuration and attempt to refresh with each new request.

#### HTTP Basic Auth
You can protect a redirect behind http basic auth using the `auth` field as described in the schema below.

Notes:
- if the username and password of the basic auth were sent through http, they could be intercepted by a hacker.
- Browsers cache username and passwords for http basic auth, so in case it's intended for personal use only, make sure that you use incognito browsing if you are opening the url from someone else's device.

### Configuration File
The configuration file is a yaml file having the following structure:

```yaml
# The port for the application to listen on
# Default: 80
port: 3000

# Options for managing the cached configurations in case it's loaded from a URL
url-config-refresh:
  # Cache time to live, will attempt to refresh the configuration after that time
  # Available time units are ("ns" for nanosecond, "us" for microsecond, "ms" for millisecond, "s" for second, "m" for minute, "h" for hour)
  # You can use fractions like 2h30m10s
  # Default: "6h"
  cache-ttl: 12h

  # Whether to re-perform the mapping (from received request to target url) again after refresh
  # Default: false
  remap-after-refresh: true

  # A list of domains to refresh when there is redirect miss/hit
  # Any matching domain from this list, will overwrite refresh-on-hit and refresh-on-miss options
  refresh-domains:
    - # Domain name, can include wild cards just like redirects
      # Required field, must be a valid domain name
      domain: '*.amr-saber.io'

      # Whether to refresh when there is a miss or a hit
      # When set to "hit" and the domain name matches the refresh domain, acts as if refresh-on-hit is set for this request
      # When set to "miss" and the domain name matches the refresh domain, acts as if refresh-on-miss is set for this request
      # Required field, must be one of (hit, miss)
      refresh-on: miss

  # Whether to refresh or not when there was a matched redirect
  # Default: false
  refresh-on-hit: true

  # Whether to refresh or not when there was no matched redirect
  # Default: false
  refresh-on-miss: true

# Auth schemas to be used with redirects
auth:
  # Basic auth is the only support type of auth for now
  basic-auth:
    # Schema name. Can be anything
    some-auth:
      # Basic auth realm. Can be anything. 
      # Default: "Restricted"
      realm: "MyRealm"

      # Users defined within this schema. Each of them must have a non-empty username and password
      # username cannot repeat across all defined basic-auth schemas
      users:
        - username: user-1
          password: 1234
        - username: user-2
          password: 5678

    # You can have as many auth schemas as you want
    some-other-auth:
      realm: "MyRealm"
      users:
        - username: user-4
          password: 1234
        - username: user-5
          password: 5678

# Whether or not the application should send temp redirection, the application will send permanent redirection status if set to false
# The browser will cache the result if the status is permanent redirect, resulting in faster redirection,
# but slower invalidation in case you changed redirection target
# Default: true
temp-redirect: false

# The list of redirection rules
redirects:
  - # Will redirect traffic from this domain
    # You can use * in place of domain sections, e.g. *.amr-saber.io, *.*.io, *.amr-saber.*, *.*.* will all match (subdomain.amr-saber.io)
    # Required field, and must not contain protocol or port
    from: subdomain.amr-saber.io

    # Will redirect traffic to this URL
    # Required field, and must be a valid URL; cannot contain a path if `preserve-path` option is true
    to: https://google.com

    # Whether or not to preserve the path when redirecting
    # e.g. if set to true: subdomain.amr-saber.io/hello-world will redirect to https://google.com/hello-world
    # if set to false: subdomain.amr-saber.io/hello-world will redirect to https://google.com
    # Default: false
    preserve-path: true

    # You can specify if this is a temp redirect or not per each redirect, this will overwrite the global temp-redirect option
    # Default: value of global `temp-redirect` field
    temp-redirect: true

    # Auth configuration, must be one of the schemas defined in `auth` global block
    # If several basic-auth schemas are used, all of them must have the same realm
    auth:
      - some-auth
      - some-other-auth

  - from: *.amr-saber.io

    # "to" can also contain *, in this case both "to" and "from" must have the same structure
    # and will substitute the wildcard(s) part(s) from the corresponding parts from the "from" domain
    # i.e. if "to" is *.x.z, "from" can be (a.b.c, a.*.c, a.b.*, *.*.*) but not b.c or a.b.c.d
    # This is useful if you wish to redirect all subdomains to a new domain for example
    # e.g. "to": (*.x.z) "from": (*.b.c) => request from (a.b.c) will be redirected to (a.x.z)
    # This can -of course- be combined with all the other options (preserve-path, temp-redirect, auth, ...)
    to: https://*.amrsaber.io
```

### Redirection Notes
In case the request comes from a domain that matches several redirection rules, redirector will redirect to the first exact match if it's found, otherwise, it will redirect ot the first match with wildcard.

## Logging
Redirector logs different events (like starting server, configuration parsing and update, received requests) to STDOUT, and logs errors and warnings to STDERR.

## Bugs and Feature Requests
If you find any bug, or want to request any feature, feel free to [open a ticket](https://github.com/AmrSaber/redirector/issues).
