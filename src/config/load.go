package config

import (
	"context"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/AmrSaber/redirector/src/watchers"
)

func LoadConfig(ctx context.Context, filePath, url string) *Config {
	if filePath != "" {
		// Read file contents
		yamlFile, err := ioutil.ReadFile(filePath)
		if err != nil {
			log.Fatal(err)
		}

		// Parse config
		configs, err := ConfigFromYaml(yamlFile)
		if err != nil {
			log.Fatal(err)
		}

		// Watch config file for updates
		fileChan, err := watchers.WatchConfigFile(ctx, filePath)
		if err != nil {
			log.Fatal("could not watch config file: ", err)
		}

		// Update config on file change
		go func() {
			for fileContents := range fileChan {
				newConfigs, err := ConfigFromYaml(fileContents)
				if err != nil {
					log.Println("Could not parse updated config file: ", err)
					continue
				}

				configs.CopyFrom(newConfigs)
			}
		}()

		return configs
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

		configs, err := ConfigFromYaml(body)
		if err != nil {
			log.Fatal(err)
		}

		return configs
	}

	return nil
}
