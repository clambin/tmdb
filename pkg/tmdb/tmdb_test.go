package tmdb_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
)

func makeTestServer(path string, f func(*http.Request) string) *httptest.Server {
	m := http.NewServeMux()
	m.Handle(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		input, err := os.Open(filepath.Join("testdata", f(r)))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		defer func(f *os.File) { _ = f.Close() }(input)
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.Copy(w, input)

	}))
	return httptest.NewServer(m)
}
