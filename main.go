package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/context"
	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
)

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
	LastUpdate        time.Time
}

func (c DoCheck) validate(bodyString string) Result {
	validations := make([]Validation, 0)
	for _, validation := range c.check.Validations {
		if validation.Contain != "" && !strings.Contains(bodyString, validation.Contain) ||
			validation.NotContain != "" && strings.Contains(bodyString, validation.NotContain) {
			validations = append(validations, validation)
		}
	}
	if len(validations) > 0 {
		log.Println("Validation Failed - " + c.check.Name + " " + bodyString)
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
	}
	return "should not contain " + v.NotContain
}

var m map[string]Result

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

func run() {

	var config Config
	err := LoadConfiguration("./checks.json", &config)
	if err != nil {
		log.Fatal(err)
		return
	}

	messages := make(chan Result)

	for _, check := range config.Checks {
		go func(doCheck DoCheck) {

			for {
				bodyString, err := doCheck.DoRequest()

				if err != nil {
					log.Println("ERROR", err)
					time.Sleep(30000 * time.Millisecond)
					continue
				}

				result := doCheck.validate(bodyString)
				result.LastUpdate = time.Now()
				messages <- result
				time.Sleep(30000 * time.Millisecond)
			}

		}(DoCheck{check, request})
	}

	m = make(map[string]Result)

	for {
		log.Println("Waiting message")
		r := <-messages
		m[r.Name] = r
		log.Println("Receive message " + r.Name)
	}
}

func result(w http.ResponseWriter, r *http.Request) {
	result := make([]Result, 0)
	for k := range m {
		result = append(result, m[k])
	}

	sort.Sort(ByName(result))
	json.NewEncoder(w).Encode(result)
}

func recoverHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic: %+v", err)
				http.Error(w, http.StatusText(500), 500)
			}
		}()
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

func xTidHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		xtid := r.Header.Get("X-tid")
		if xtid == "" {
			xtid = strconv.Itoa(rand.Intn(1000))
		}
		context.Set(r, "xtid", xtid)
		w.Header().Add("X-tid", xtid)
		next.ServeHTTP(w, r)

	}
	return http.HandlerFunc(fn)
}

func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "hello")
}

func wrapHandler(h http.Handler) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		context.Set(r, "params", ps)
		h.ServeHTTP(w, r)
	}
}

type router struct {
	*httprouter.Router
}

func (r *router) Get(path string, handler http.Handler) {
	r.GET(path, wrapHandler(handler))
}

func NewRouter() *router {
	return &router{httprouter.New()}
}

func index(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("index.html")
	t.Execute(w, nil)
}

func main() {
	go run()
	router := NewRouter()
	commonHandler := alice.New(context.ClearHandler, xTidHandler, recoverHandler)
	router.Get("/hello", commonHandler.ThenFunc(hello))
	router.Get("/result", commonHandler.ThenFunc(result))
	router.Get("/", http.FileServer(http.Dir(".")))
	http.ListenAndServe(":8000", router)
}
