package wmcheck

import (
	"strings"
	"time"
)

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

type Env struct {
	App struct {
		Port string `json:"port"`
	} `json:"app"`
	Slack struct {
		SLACKCHANNEL string `json:"SLACK_CHANNEL"`
		SLACKTOKEN   string `json:"SLACK_TOKEN"`
	} `json:"slack"`
}

func (c Check) getHeaders() map[string]string {
	headers := make(map[string]string)
	for _, h := range c.Request.Headers {
		headers[h.Key] = h.Value
	}
	return headers
}

func (c Check) validate(bodyString string) []Validation {
	validations := make([]Validation, 0)
	for _, validation := range c.Validations {
		if validation.Contain != "" && !strings.Contains(bodyString, validation.Contain) ||
			validation.NotContain != "" && strings.Contains(bodyString, validation.NotContain) {
			validations = append(validations, validation)
		}
	}
	return validations
}

type Result struct {
	Name              string
	FailedValidations []Validation
	LastUpdate        time.Time
	Body              string
}

type Validation struct {
	Contain    string `json:"contain"`
	NotContain string `json:"notContain"`
}

func (v Validation) String() string {
	if v.Contain != "" {
		return "should contain " + v.Contain
	}
	return "should not contain " + v.NotContain
}

type ByName []Result

func (s ByName) Len() int {
	return len(s)
}
func (s ByName) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s ByName) Less(i, j int) bool {
	return s[i].Name < s[j].Name
}
