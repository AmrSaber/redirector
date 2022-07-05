package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path"

	"github.com/AmrSaber/redirector/src/config"
	"github.com/joho/godotenv"
)

// TODO: Get configuration with redirection rules -- read configuration from (file, URL, environment variable)

func main() {
	var configs *config.Config
	var file string
	var url string

	godotenv.Load()

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

	if configs == nil {
		log.Fatal("No configuration provided")
	}

	attachRedirectionHandler(configs)

	fmt.Println("Starting server...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("Could not start server: ", err)
	}
}

func attachRedirectionHandler(configs *config.Config) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Print matched path
		log.Println(r.URL.Path)

		// Print domain
		log.Println(r.Host)

		redirectInfo := configs.GetRedirect(r.Host)
		if redirectInfo == nil {
			// Report 404
			w.WriteHeader(http.StatusNotFound)
			w.Header().Add("Content-Type", "application/json")
			w.Write([]byte(`{ "message": "could not match host to any redirect rule" }`))
			return
		}

		redirectPath := redirectInfo.To

		if redirectInfo.PreservePath {
			redirectPath = path.Join(redirectPath, r.URL.Path)
		}

		http.Redirect(w, r, redirectPath, http.StatusPermanentRedirect)
	})
}
