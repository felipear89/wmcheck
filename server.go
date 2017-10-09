package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sort"
	lib "wmcheck/lib"

	"github.com/gorilla/context"
	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
)

type appContext struct {
	results map[string]lib.Result
}

func (ac appContext) result(w http.ResponseWriter, r *http.Request) {
	mapResult := ac.results
	results := make([]lib.Result, 0)
	for k := range mapResult {
		results = append(results, mapResult[k])
	}

	sort.Sort(lib.ByName(results))
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

func main() {
	appContext := appContext{}
	appContext.results = make(map[string]lib.Result)
	messages := make(chan lib.Result)

	lib.StartMonitor(messages)

	go func() {
		for {
			log.Println("Waiting message")
			r := <-messages
			appContext.results[r.Name] = r
			log.Println("Receive message " + r.Name)
		}
	}()

	router := newRouter()
	commonHandler := alice.New(context.ClearHandler, recoverHandler)
	router.Get("/result", commonHandler.ThenFunc(appContext.result))
	router.Get("/", http.FileServer(http.Dir(".")))
	http.ListenAndServe(":8000", router)

}
