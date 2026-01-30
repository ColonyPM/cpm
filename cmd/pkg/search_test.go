package pkgcmd

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSearchPackages(t *testing.T) {
	originalBaseURL := baseURL
	defer func() { baseURL = originalBaseURL }()

	tests := []struct {
		name        string
		args        []string
		handler     func(w http.ResponseWriter, r *http.Request)
		expectedOut string
		errContains string
	}{
		{
			name: "Success: Packages found",
			args: []string{"test-pkg"},
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Query().Get("q") != "test-pkg" {
					http.Error(w, "bad query", http.StatusBadRequest)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`["alpha", "beta"]`))
			},
			expectedOut: "alpha\nbeta\n",
		},
		{
			name: "Success: No packages found",
			args: []string{"ghost-pkg"},
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`[]`))
			},
			expectedOut: "No packages found.\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(tt.handler))
			defer srv.Close()

			baseURL = srv.URL + "/"

			cmd := newPkgSearchCmd()
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()

			if tt.errContains != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedOut, strings.ReplaceAll(buf.String(), "\r\n", "\n"))
			}
		})
	}
}
