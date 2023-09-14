package config

import (
	"fmt"
	"net/http"
	"testing"
)

func TestResolveDomain(t *testing.T) {
	redirect := Redirect{
		From:         "*.amr-saber.com",
		To:           "https://*.amrsaber.io",
		PreservePath: true,
	}

	request, _ := http.NewRequest("GET", "https://abc.amr-saber.com/some-path", nil)
	redirectUrl := redirect.ResolvePath(request)
	fmt.Println("redirect url:", redirectUrl)
}
