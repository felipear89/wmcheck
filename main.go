package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

// DoCheck control the check execution
type DoCheck struct {
	check   Check
	request func(Check) (string, error)
}

func (d DoCheck) DoRequest() (string, error) {
	return d.request(d.check)
}

func request(check Check) (string, error) {

	req, err := NewRequest(check)
	if err != nil {
		return "", err
	}

	bodyString, err := DoRequest(req)
	if err != nil {
		return "", err
	}
	return bodyString, nil

}

type Result struct {
	Name              string
	FailedValidations []Validation
}

func (c DoCheck) validate(bodyString string) Result {
	validations := make([]Validation, 0)
	for _, validation := range c.check.Validations {
		if validation.Contain != "" && !strings.Contains(bodyString, validation.Contain) ||
			validation.NotContain != "" && strings.Contains(bodyString, validation.NotContain) {
			validations = append(validations, validation)
		}
	}
	return Result{Name: c.check.Name, FailedValidations: validations}
}

type Config struct {
	Checks []Check `json:"checks"`
}

type Check struct {
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
	Validations []Validation `json:"validations"`
}

type Validation struct {
	Contain    string `json:"contain"`
	NotContain string `json:"notContain"`
}

func (v Validation) String() string {
	if v.Contain != "" {
		return "should contain " + v.Contain
	} else {
		return "should not contain " + v.NotContain
	}
}

// LoadConfiguration open a file and decode json to Config struct
func LoadConfiguration(file string, config *Config) error {
	configFile, err := os.Open(file)
	if err != nil {
		return err
	}
	defer configFile.Close()
	return json.NewDecoder(configFile).Decode(config)
}

func NewRequest(check Check) (*http.Request, error) {
	req, err := http.NewRequest(check.Request.Method, check.Request.URL, strings.NewReader(check.Request.Body))
	if err != nil {
		return nil, err
	}
	for _, header := range check.Request.Headers {
		req.Header.Set(header.Key, header.Value)
	}
	return req, nil
}

func DoRequest(req *http.Request) (string, error) {
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func main() {
	var config Config
	err := LoadConfiguration("./checks.json", &config)
	if err != nil {
		log.Fatal(err)
		return
	}

	messages := make(chan Result)

	for _, check := range config.Checks {
		go func(doCheck DoCheck) {
			bodyString, err := doCheck.DoRequest()

			if err != nil {
				log.Fatal(err)
			}

			result := doCheck.validate(bodyString)
			messages <- result

		}(DoCheck{check, request})
	}

	for i := 0; i < len(config.Checks); i++ {
		result := <-messages
		if len(result.FailedValidations) == 0 {
			log.Printf(result.Name + " is OK")
		} else {
			log.Printf(result.Name+" %v", result.FailedValidations)
		}

	}
	close(messages)
}
