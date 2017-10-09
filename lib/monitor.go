package wmcheck

import (
	"encoding/json"
	"log"
	"os"
	"time"
)

func LoadConfiguration(file string, config *Config) error {
	configFile, err := os.Open(file)
	if err != nil {
		return err
	}
	defer configFile.Close()
	return json.NewDecoder(configFile).Decode(config)
}

func StartMonitor(messages chan Result) {

	var config Config
	err := LoadConfiguration("./checks.json", &config)
	if err != nil {
		log.Fatal(err)
		return
	}

	for _, c := range config.Checks {
		go func(check Check) {

			for {
				bodyString, err := Request(check.Request.Method, check.Request.URL, check.Request.Body, check.getHeaders())

				if err != nil {
					log.Println("ERROR", err)
				}

				result := check.validate(bodyString)
				result.LastUpdate = time.Now()
				messages <- result
				time.Sleep(30000 * time.Millisecond)
			}
		}(c)
	}
}
