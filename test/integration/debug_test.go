package integration

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBasicRouting(t *testing.T) {
	mux := http.NewServeMux()
	
	// Test basic patterns similar to what we use
	mux.HandleFunc("GET /{id}/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("GET by ID"))
	})
	
	mux.HandleFunc("PUT /{id}/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("PUT by ID"))
	})
	
	server := httptest.NewServer(mux)
	defer server.Close()
	
	// Test GET
	resp, err := http.Get(server.URL + "/123/")
	if err != nil {
		t.Fatalf("GET request failed: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		t.Errorf("GET: Expected status 200, got %d", resp.StatusCode)
	}
	
	// Test PUT
	req, _ := http.NewRequest("PUT", server.URL + "/123/", nil)
	client := &http.Client{}
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("PUT request failed: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		t.Errorf("PUT: Expected status 200, got %d", resp.StatusCode)
	}
}

func TestStripPrefixRouting(t *testing.T) {
	mainMux := http.NewServeMux()
	
	// Create a sub-mux similar to our handlers
	subMux := http.NewServeMux()
	subMux.HandleFunc("GET /{id}/", func(w http.ResponseWriter, r *http.Request) {
		t.Logf("Sub-mux received: %s %s", r.Method, r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("GET by ID"))
	})
	
	subMux.HandleFunc("PUT /{id}/", func(w http.ResponseWriter, r *http.Request) {
		t.Logf("Sub-mux received: %s %s", r.Method, r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("PUT by ID"))
	})
	
	// Mount with StripPrefix like we do
	mainMux.Handle("/api/v1/users/", http.StripPrefix("/api/v1/users", subMux))
	
	server := httptest.NewServer(mainMux)
	defer server.Close()
	
	// Test GET
	resp, err := http.Get(server.URL + "/api/v1/users/123/")
	if err != nil {
		t.Fatalf("GET request failed: %v", err)
	}
	defer resp.Body.Close()
	
	t.Logf("GET Response status: %d", resp.StatusCode)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("GET: Expected status 200, got %d", resp.StatusCode)
	}
	
	// Test PUT
	req, _ := http.NewRequest("PUT", server.URL + "/api/v1/users/123/", nil)
	client := &http.Client{}
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("PUT request failed: %v", err)
	}
	defer resp.Body.Close()
	
	t.Logf("PUT Response status: %d", resp.StatusCode)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("PUT: Expected status 200, got %d", resp.StatusCode)
	}
}