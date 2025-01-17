package examples

import (
	"github.com/go-gum/unravel"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

type Student struct {
	Name string
	Age  int
}

func TestHttpRequestPathValueMux(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /register/student/{Name}/age/{Age}", func(w http.ResponseWriter, req *http.Request) {
		var student Student
		if err := unravel.Unmarshal(PathValueSource{Request: req}, &student); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// business.RegisterStudent(w, student)
	})

	newServer := httptest.NewServer(mux)
	defer newServer.Client()

	client := newServer.Client()
	_, err := client.Get(newServer.URL + "/register/student/Albert/age/18")
	require.NoError(t, err)
}

func TestPathValueSource(t *testing.T) {
	var req http.Request
	req.SetPathValue("Name", "Albert")
	req.SetPathValue("Age", "18")

	var student Student
	err := unravel.Unmarshal(PathValueSource{Request: &req}, &student)

	require.NoError(t, err)
	require.Equal(t, Student{Name: "Albert", Age: 18}, student)
}

type PathValueSource struct {
	unravel.EmptySource
	Request *http.Request
}

func (p PathValueSource) Get(key string) (unravel.Source, error) {
	v := p.Request.PathValue(key)
	if v == "" {
		return nil, unravel.ErrNoValue
	}

	return unravel.StringSource(v), nil
}
