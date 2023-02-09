package api

import (
	"log"
	"net/http"
	"net/http/httptest"
)

const testAPI = "this-is-my-secure-token-do-not-steal!!"

func GetTestToken() string {
	return testAPI
}

// OpenAITestServer Creates a mocked OpenAI server which can pretend to handle requests during testing.
func OpenAITestServer() *httptest.Server {
	return httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("received request at path %q\n", r.URL.Path)

		// check auth
		if r.Header.Get("Authorization") != "Bearer "+GetTestToken() {
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

// serverMap.
var serverMap = make(map[string]Handler)

// RegisterHandler Register handler.
func RegisterHandler(path string, handler Handler) {
	serverMap[path] = handler
}
