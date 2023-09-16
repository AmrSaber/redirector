package models

import "regexp"

var domainRegex = regexp.MustCompile(`^(?:[a-zA-Z0-9-_]+|\*)(?:\.(?:[a-zA-Z0-9-_]+|\*))+$`)
var urlRegex = regexp.MustCompile(`^\w+://(?:[a-zA-Z0-9-_]+|\*)(?:\.(?:[a-zA-Z0-9-_]+|\*))+(?:/[^/]*)*$`)
var hasPathRegex = regexp.MustCompile(`^.+//.+(?:/[^/]*)+$`)
