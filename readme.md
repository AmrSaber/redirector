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
./redirector --file config.yaml
```

### Docker Image

You can use the application's docker image `ghcr.io/amrsaber/redirector` You can also download a specific version with image tag, available tags can be found [here]().

For example

```bash
docker run -p 8080:8080 ghcr.io/amrsaber/redirector:latest --url https://url-to-config.com
```

Or using a file

```bash
docker run -p 8080:8080 -v /path/to/config.yaml:/config.yaml ghcr.io/amrsaber/redirector:latest --file /config.yaml
```

## Usage

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

### Configuration File

The configuration file is a yaml file having the following structure:

```yaml
# The port for the application to listen on
# Default: 8080
port: 3000

# Cache time to live, if the configuration is loaded from a URL, will attempt to refresh the configuration after that time
# Available time units are ("ns" for nanosecond, "us" for microsecond, "ms" for millisecond, "s" for second, "m" for minute, "h" for hour)
# You can use fractions like 2h30m10s
# Default: "4h"
cache-ttl: 6h

# Whether or not the application should send temp redirection, the application will send permanent redirection status if set to false
# The browser will cache the result if the status is permanent redirect, resulting in faster redirection, but slower invalidation in case you changed redirection target
# Default: false
temp-redirect: true

# The list of redirection rules
redirects:
  - # Will redirect traffic from this domain
    # You can use * in place of domain sections, e.g. *.amr-saber.io, *.*.io, *.amr-saber.*, *.*.* will all match (subdomain.amr-saber.io)
    # Required field, and must not contain protocol or ports
    from: subdomain.amr-saber.io

    # Will redirect traffic to this URL
    # Required field, and must be a valid URL; cannot contain a path if `preserve-path` option is set
    to: https://google.com

    # Whether or not to preserve the path when redirecting
    # e.g. if set to true: subdomain.amr-saber.io/hello-world will redirect to https://google.com/hello-world
    # if set to false: subdomain.amr-saber.io/hello-world will redirect to https://google.com
    # Default: false
    preserve-path: true

    # You can specify if this is a temp redirect or not per each redirect, this will overwrite the global temp-redirect option
    # Default: same value as global `temp-redirect` field
    temp-redirect: true
```

### Redirection Notes

In case the request comes from a domain that matches several redirection rules, redirector will redirect to the first exact match if it's found, otherwise, it will redirect ot the first match with wildcard.

## Logging

Redirector logs different events (like starting server, configuration parsing and update, received requests) to STDOUT, and logs errors and warnings to STDERR.

## Bugs and Feature Requests

If you find any bug, or want to request any feature, feel free to [open a ticket](https://github.com/AmrSaber/redirector/issues).
