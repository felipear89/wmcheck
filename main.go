package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

type Config struct {
	Checks []struct {
		Name    string `json:"name"`
		Request struct {
			Method  string `json:"method"`
			URL     string `json:"url"`
			Body    string `json:"body"`
			Headers []struct {
				Key   string `json:"key"`
				Value string `json:"Value"`
			} `json:"headers"`
		} `json:"request"`
		Validations []struct {
			Contain    string `json:"contain"`
			NotContain string `json:"notContain"`
		} `json:"validations"`
	} `json:"checks"`
}

// LoadConfiguration open a file and decode to Config struct
func LoadConfiguration(file string) Config {
	var config Config
	configFile, err := os.Open(file)
	defer configFile.Close()
	if err != nil {
		fmt.Println(err.Error())
	}
	json.NewDecoder(configFile).Decode(&config)
	return config
}

func main() {
	config := LoadConfiguration("./checks.json")

	for _, check := range config.Checks {

		log.Printf("Checking " + check.Name)

		req, err := http.NewRequest(check.Request.Method, check.Request.URL, strings.NewReader(check.Request.Body))
		if err != nil {
			log.Fatal("Check: ["+check.Name+"]"+" - Error: ", err)
			return
		}
		for _, header := range check.Request.Headers {
			req.Header.Set(header.Key, header.Value)
		}

		client := http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Fatal("Check: ["+check.Name+"]"+" - Error: ", err)
			return
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal("Check: ["+check.Name+"]"+" - Error: ", err)
			return
		}

		bodyString := string(body)
		errors := 0
		for _, validation := range check.Validations {
			if validation.Contain != "" && !strings.Contains(bodyString, validation.Contain) {
				log.Println("[ValidationError] " + check.Name + " not contain: " + validation.Contain)
				errors++
			}
			if validation.NotContain != "" && strings.Contains(bodyString, validation.NotContain) {
				log.Println("[ValidationError] " + check.Name + " contain: " + validation.NotContain)
				errors++
			}
		}
		if errors == 0 {
			log.Println("OK!")
		}
	}
}
