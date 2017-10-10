package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"sla/model"
	"testing"
)

var a App
var p1 = model.Provider{Id: "01", Name: "Provider01"}

// TestMain runs the tests
func TestMain(m *testing.M) {
	repo := model.MemRepository{}
	repo.CreateProvider(&p1)
	a = App{}
	a.Initialize(repo)

	os.Exit(m.Run())
}

func TestGetProviders(t *testing.T) {
	req, _ := http.NewRequest("GET", "/providers", nil)
	res := request(req)
	checkStatus(t, http.StatusOK, res.Code)

	var providers model.Providers
	_ = json.NewDecoder(res.Body).Decode(&providers)
	if len(providers) != 1 {
		t.Errorf("Expected 1 provider. Received: %v", providers)
	}
}
func TestGetProviderExists(t *testing.T) {
	req, _ := http.NewRequest("GET", "/providers/01", nil)
	res := request(req)
	checkStatus(t, http.StatusOK, res.Code)
	/*
	 * Check body
	 */
	var provider model.Provider
	_ = json.NewDecoder(res.Body).Decode(&provider)
	if provider != p1 {
		t.Errorf("Expected: %v. Actual: %v", p1, provider)
	}
}

func TestGetProviderNotExists(t *testing.T) {
	req, _ := http.NewRequest("GET", "/providers/doesnotexist", nil)
	res := request(req)
	checkStatus(t, http.StatusNotFound, res.Code)

}

func TestCreateProviderThatExists(t *testing.T) {
	body, err := json.Marshal(p1)
	if err != nil {
		t.Error("Unexpected marshalling error")
	}
	req, _ := http.NewRequest("POST", "/providers", bytes.NewBuffer(body))
	res := request(req)

	checkStatus(t, http.StatusConflict, res.Code)
}

func request(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	a.Router.ServeHTTP(rr, req)

	return rr
}

func checkStatus(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected status %d. Got %d\n", expected, actual)
	}
}
