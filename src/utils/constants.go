package utils

import (
	"os"
	"path"
	"regexp"
)

const DEFAULT_REALM = "Restricted"

var SOCKET_PATH = path.Join(os.TempDir(), "redirector.sock")

const SOCKET_MESSAGE_PING = "@redirector:PING"
const SOCKET_MESSAGE_CLOSE = "@redirector:CLOSE"

// Regex
var DomainRegex = regexp.MustCompile(`^(?:[a-zA-Z0-9-_]+|\*)(?:\.(?:[a-zA-Z0-9-_]+|\*))+$`)
var UrlRegex = regexp.MustCompile(`^\w+://(?:[a-zA-Z0-9-_]+|\*)(?:\.(?:[a-zA-Z0-9-_]+|\*))+(?:/[^/]*)*$`)
var HasPathRegex = regexp.MustCompile(`^.+//.+(?:/[^/]*)+$`)
