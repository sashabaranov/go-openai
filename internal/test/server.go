package test

import (
	"log"
	"net/http"
	"net/http/httptest"
)

const testAPI = "this-is-my-secure-token-do-not-steal!!"

func GetTestToken() string {
	return testAPI
}

type ServerTest struct {
	handlers map[string]handler
}
type handler func(w http.ResponseWriter, r *http.Request)

func NewTestServer() *ServerTest {
	return &ServerTest{handlers: make(map[string]handler)}
}

func (ts *ServerTest) RegisterHandler(path string, handler handler) {
	ts.handlers[path] = handler
}

// OpenAITestServer Creates a mocked OpenAI server which can pretend to handle requests during testing.
func (ts *ServerTest) OpenAITestServer() *httptest.Server {
	return httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("received request at path %q\n", r.URL.Path)

		// check auth
		if r.Header.Get("Authorization") != "Bearer "+GetTestToken() && r.Header.Get("api-key") != GetTestToken() {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		handlerCall, ok := ts.handlers[r.URL.Path]
		if !ok {
			http.Error(w, "the resource path doesn't exist", http.StatusNotFound)
			return
		}
		handlerCall(w, r)
	}))
}
