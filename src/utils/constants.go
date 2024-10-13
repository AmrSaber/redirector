package utils

import (
	"os"
	"path"
)

var SOCKET_PATH = path.Join(os.TempDir(), "redirector.sock")

const SOCKET_MESSAGE_PING = "@redirector:PING"
const SOCKET_MESSAGE_CLOSE = "@redirector:CLOSE"
