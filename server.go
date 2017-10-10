package wmcheck

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"sort"

	"github.com/gorilla/context"
	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
)

type appContext struct {
	results    map[string]Result
	resultChan chan Result
}

func (ac appContext) result(w http.ResponseWriter, r *http.Request) {
	mapResult := ac.results
	results := make([]Result, 0)
	for k := range mapResult {
		results = append(results, mapResult[k])
	}

	sort.Sort(ByName(results))
	json.NewEncoder(w).Encode(results)
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

func newRouter() *router {
	return &router{httprouter.New()}
}

func getStatusMessage(failedValidations []Validation) string {
	if len(failedValidations) == 0 {
		return "OK"
	}
	return "FAIL"
}

func resultChanListener(ac appContext) {
	for {
		log.Println("Waiting for a result...")
		r := <-ac.resultChan
		if len(r.FailedValidations) > 0 {
			log.Println("Validation Failed - " + r.Name + " " + r.Body)
		}
		if len(ac.results[r.Name].FailedValidations) != len(r.FailedValidations) {
			token := os.Getenv("SLACK_TOKEN")
			channel := os.Getenv("SLACK_CHANNEL")
			statusMessage := getStatusMessage(r.FailedValidations)
			text := url.QueryEscape(r.Name + " updated! The status is: " + statusMessage)
			url := "https://slack.com/api/chat.postMessage?token=%s&channel=%s&text=%s&as_user=true"
			url = fmt.Sprintf(url, token, channel, text)
			Request("GET", url, "", make(map[string]string))
		}
		ac.results[r.Name] = r
		log.Println("Result " + r.Name + " updated")
	}
}

func loadConfig() Config {
	var config Config
	configPath := os.Getenv("CONFIG_PATH")
	err := loadConfiguration(configPath, &config)
	if err != nil {
		log.Fatalf("Failed to load CONFIG_PATH=%s %s\n", configPath, err)
	}
	return config
}

func StartServer() {

	appContext := appContext{results: make(map[string]Result), resultChan: make(chan Result)}

	StartMonitor(appContext.resultChan, loadConfig())

	go resultChanListener(appContext)

	router := newRouter()
	commonHandler := alice.New(context.ClearHandler, recoverHandler)
	router.Get("/result", commonHandler.ThenFunc(appContext.result))
	router.Get("/", http.FileServer(http.Dir(".")))
	http.ListenAndServe(":8000", router)

}
