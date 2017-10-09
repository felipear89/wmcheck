package wmcheck

import (
	"encoding/json"
	"log"
	"os"
	"time"
)

func loadConfiguration(file string, config *Config) error {
	configFile, err := os.Open(file)
	if err != nil {
		return err
	}
	defer configFile.Close()
	return json.NewDecoder(configFile).Decode(config)
}

func StartMonitor(resultChan chan Result, config Config) {

	for _, c := range config.Checks {
		go func(check Check) {

			for {
				bodyString, err := Request(check.Request.Method, check.Request.URL, check.Request.Body, check.getHeaders())

				if err != nil {
					log.Println("ERROR ", err)
				}

				validations := check.validate(bodyString)
				result := Result{Name: check.Name, FailedValidations: validations, LastUpdate: time.Now(), Body: bodyString}
				resultChan <- result
				time.Sleep(30000 * time.Millisecond)
			}
		}(c)
	}
}
