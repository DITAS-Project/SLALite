package main

import (
	"SLALite/model"
	"SLALite/repositories"
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
)

var a App
var p1 = model.Provider{Id: "01", Name: "Provider01"}
var dbName = "test.db"
var prefix = "pf_" + strconv.Itoa(rand.Int())

// TestMain runs the tests
func TestMain(m *testing.M) {
	repo := repositories.MemRepository{}
	var err error = nil
	//repo,err := repositories.CreateBBoltRepository()
	//repo, err := repositories.CreateMongoDBRepository()
	if err == nil {
		//BBolt test database
		//repo.SetDatabase(dbName)

		//MongoDB test database
		//repo.SetDatabase("slaliteTest", true)
		repo.CreateProvider(&p1)
		a = App{}
		a.Initialize(repo)
	} else {
		log.Fatal(err)
	}

	result := m.Run()

	//BBolt clear database
	os.Remove(dbName)

	os.Exit(result)
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
	checkError(t, res, http.StatusNotFound, res.Code)

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

func TestCreateProvider(t *testing.T) {
	posted := model.Provider{Id: "new", Name: "New provider"}
	body, err := json.Marshal(posted)
	if err != nil {
		t.Error("Unexpected marshalling error")
	}
	req, _ := http.NewRequest("POST", "/providers", bytes.NewBuffer(body))
	res := request(req)

	checkStatus(t, http.StatusCreated, res.Code)

	var created model.Provider
	_ = json.NewDecoder(res.Body).Decode(&created)
	if created != posted {
		t.Errorf("Expected: %v. Actual: %v", posted, created)
	}
}

func TestDeleteProviderThatNotExists(t *testing.T) {
	req, _ := http.NewRequest("DELETE", "/providers/doesnotexist", nil)
	res := request(req)

	checkStatus(t, http.StatusNotFound, res.Code)
	// TODO Check body
}

func TestDeleteProvider(t *testing.T) {
	req, _ := http.NewRequest("DELETE", "/providers/01", nil)
	res := request(req)

	checkStatus(t, http.StatusNoContent, res.Code)
	body, _ := ioutil.ReadAll(res.Body)

	if len(body) > 0 {
		t.Errorf("Expected empty body. Actual: %s", body)
	}
}

func request(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	a.Router.ServeHTTP(rr, req)

	return rr
}

func checkStatus(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected status %d. Actual %d\n", expected, actual)
	}
}

func checkError(t *testing.T, res *httptest.ResponseRecorder, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected status %d. Actual %d\n", expected, actual)
	}
	var body ApiError
	err := json.NewDecoder(res.Body).Decode(&body)
	if err != nil {
		t.Errorf("Error unmarshalling error: %s", err.Error())
	} else {
		if code, _ := strconv.Atoi(body.Code); code != expected {
			t.Errorf("Unmarshalled error. Expected code %d. Actual: %s", expected, body.Code)
		}
	}

}

func BenchmarkDatabase(b *testing.B) {
	executeCreate(b)
	executeGetAll(b)
	executeGet(b)
	executeDelete(b)
}

func executeCreate(b *testing.B) {
	//log.Print("Running create test " + b.Name())
	for i := 0; i < b.N; i++ {
		key := getProviderId(i)
		provider := model.Provider{key, "provider_" + key}
		body, err := json.Marshal(provider)
		if err != nil {
			b.Error("Unexpected marshalling error")
		}
		req, _ := http.NewRequest("POST", "/providers", bytes.NewBuffer(body))
		res := request(req)
		if http.StatusCreated != res.Code && http.StatusConflict != res.Code {
			b.Error("Error creating provider: " + res.Body.String())
		}
	}
}

func executeGetAll(b *testing.B) {
	req, _ := http.NewRequest("GET", "/providers", nil)
	res := request(req)
	if http.StatusOK != res.Code {
		b.Error("Error getting list of providers")
	}

	var providers model.Providers
	_ = json.NewDecoder(res.Body).Decode(&providers)
	if len(providers) != b.N+1 {
		b.Error("Expected "+strconv.Itoa(b.N+1)+" providers. Received: %v", providers)
	}
}

func executeGet(b *testing.B) {
	for i := 0; i < b.N; i++ {
		key := getProviderId(i)
		req, _ := http.NewRequest("GET", "/providers/"+key, nil)
		res := request(req)
		if http.StatusOK != res.Code {
			b.Error("Provider " + key + " not found: " + strconv.Itoa(res.Code))
		}
	}
}

func executeDelete(b *testing.B) {
	for i := 0; i < b.N; i++ {
		key := getProviderId(i)
		req, _ := http.NewRequest("DELETE", "/providers/"+key, nil)
		res := request(req)
		if http.StatusNoContent != res.Code {
			b.Error("Provider " + key + " not found: " + strconv.Itoa(res.Code))
		}
	}
}

func getProviderId(i int) string {
	return prefix + "_" + strconv.Itoa(i)
}
