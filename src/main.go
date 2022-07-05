package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/AmrSaber/redirector/src/config"
	"github.com/joho/godotenv"
)

// TODO:
// 1. Watch config file for changes
// 2. Cache config file from URL and invalidate the cache after some time (adjustable from config)
// 3. Add tests to github actions workflow
// 4. Dockerize the app
// 5. Add github actions workflow to auto-publish docker image on github

const URL_ENV_NAME = "CONFIG_URL"

func main() {
	godotenv.Load()

	configs := loadConfig()
	if configs == nil {
		log.Fatal("No configuration provided")
	}

	attachRedirectionHandler(configs)

	fmt.Println("Starting server...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("Could not start server: ", err)
	}
}

func loadConfig() *config.Config {
	var configs *config.Config
	var file string
	var url string

	flag.StringVar(&file, "file", "", "YAML file containing configuration")
	flag.StringVar(&url, "url", "", "URL containing configuration yaml file")
	flag.Parse()

	if file != "" {
		// Read file contents
		yamlFile, err := ioutil.ReadFile(file)
		if err != nil {
			log.Fatal(err)
		}

		configs, err = config.ConfigFromYaml(yamlFile)
		if err != nil {
			log.Fatal(err)
		}
	}

	// given flag overwrites env variable
	urlEnvValue := os.Getenv(URL_ENV_NAME)
	if url == "" && urlEnvValue != "" {
		url = urlEnvValue
	} else if urlEnvValue != "" {
		log.Printf("Env variable %s overwritten by provided url flag", URL_ENV_NAME)
	}

	if url != "" {
		// Download url content
		res, err := http.Get(url)
		if err != nil {
			log.Fatal(err)
		}

		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			log.Fatal(err)
		}

		res.Body.Close()

		configs, err = config.ConfigFromYaml(body)
		if err != nil {
			log.Fatal(err)
		}
	}

	return configs
}

func attachRedirectionHandler(configs *config.Config) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		redirectInfo := configs.GetRedirect(r.Host)

		requestPath := path.Join(r.Host, r.URL.Path)

		if redirectInfo == nil {
			log.Printf("Received request for unknown host: %s", requestPath)

			// No redirects found, report 404
			w.WriteHeader(http.StatusNotFound)
			w.Header().Add("Content-Type", "application/json")
			w.Write([]byte(`{ "message": "could not match host to any redirect rule" }`))
			return
		}

		redirectPath := redirectInfo.To

		if redirectInfo.PreservePath {
			redirectPath = path.Join(redirectPath, r.URL.Path)
		}

		log.Printf("Redirecting %s to %s", requestPath, redirectPath)

		http.Redirect(w, r, redirectPath, http.StatusPermanentRedirect)
	})
}
