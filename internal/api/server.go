package api

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
)

const (
	TestAPIToken = "this-is-my-secure-token-do-not-steal!!"
)

func init() {
	if serverMap == nil {
		serverMap = make(map[string]Handler)
	}
}

// OpenAITestServer Creates a mocked OpenAI server which can pretend to handle requests during testing.
func OpenAITestServer() *httptest.Server {
	return httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("received request at path %q\n", r.URL.Path)

		// check auth
		if r.Header.Get("Authorization") != "Bearer "+TestAPIToken {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		handler, ok := serverMap[r.URL.Path]
		if !ok {
			http.Error(w, "the resource path doesn't exist", http.StatusNotFound)
			return
		}
		handler(w, r)
	}))
}

type Handler func(w http.ResponseWriter, r *http.Request)

// serverMap
var serverMap map[string]Handler

// RegisterHandler Register handler
func RegisterHandler(path string, handler Handler) {
	if _, ok := serverMap[path]; ok {
		fmt.Println("This path already has a processing function. Skip this registration")
	}
	serverMap[path] = handler
}
