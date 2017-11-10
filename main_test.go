/*
   Copyright 2017 Atos

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/
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
	"reflect"
	"strconv"
	"testing"
	"time"
)

var a App
var repo model.IRepository
var p1 = model.Provider{Id: "01", Name: "Provider01"}
var dbName = "test.db"
var providerPrefix = "pf_" + strconv.Itoa(rand.Int())
var agreementPrefix = "apf_" + strconv.Itoa(rand.Int())

var a1 = createAgreement("01", "01", "02", "Agreement 01")

func createRepository(repoType string) model.IRepository {
	var repo model.IRepository

	switch repoType {
	case defaultRepositoryType:
		repo = repositories.MemRepository{}
	case "bbolt":
		boltRepo, errRepo := repositories.CreateBBoltRepository()
		if errRepo != nil {
			log.Fatal("Error creating bbolt repository: ", errRepo.Error())
		}
		boltRepo.SetDatabase(dbName)
		repo = boltRepo
	case "mongodb":
		mongoRepo, errMongo := repositories.CreateMongoDBRepository()
		if errMongo != nil {
			log.Fatal("Error creating mongo repository: ", errMongo.Error())
		}
		mongoRepo.SetDatabase("slaliteTest", true)
		repo = mongoRepo
	}
	return repo
}

// TestMain runs the tests
func TestMain(m *testing.M) {
	repo = createRepository("mongodb")
	if repo != nil {
		repo.CreateProvider(&p1)
		repo.CreateAgreement(&a1)
		a = App{}
		a.Initialize(repo)
	} else {
		log.Fatal("Error initializing repository")
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

/********************************************************************
*****************AGREEMENTS******************************************
********************************************************************/
func TestGetAgreements(t *testing.T) {
	req, _ := http.NewRequest("GET", "/agreements", nil)
	res := request(req)
	checkStatus(t, http.StatusOK, res.Code)

	var agreements model.Agreements
	_ = json.NewDecoder(res.Body).Decode(&agreements)
	if len(agreements) != 1 {
		t.Errorf("Expected 1 agreement. Received: %v", agreements)
	}
}

func TestGetActiveAgreements(t *testing.T) {

	repo.StartAgreement("01")

	inactive := createAgreement("in1", "01", "02", "inactive")
	inactive.Active = false

	repo.CreateAgreement(&inactive)

	expired := createAgreement("expired", "01", "02", "expired")
	expired.Active = true
	expired.Expiration = time.Now().Add(-10 * time.Minute)

	repo.CreateAgreement(&expired)

	active := createAgreement("active", "01", "02", "active")
	active.Active = true
	repo.CreateAgreement(&active)

	req, _ := http.NewRequest("GET", "/agreements?active=true", nil)
	res := request(req)

	var agreements model.Agreements
	_ = json.NewDecoder(res.Body).Decode(&agreements)
	if len(agreements) != 2 {
		t.Errorf("Expected 2 agreement. Received: %v", agreements)
	}

	for _, agreement := range agreements {
		if !(agreement.Id == p1.Id || agreement.Id == active.Id) {
			t.Errorf("Got unexpected active agreement %s", agreement.Id)
		}
	}
}

func TestGetAgreementExists(t *testing.T) {
	req, _ := http.NewRequest("GET", "/agreements/01", nil)
	res := request(req)
	checkStatus(t, http.StatusOK, res.Code)
	/*
	 * Check body
	 */
	var agreement model.Agreement
	_ = json.NewDecoder(res.Body).Decode(&agreement)
	if reflect.DeepEqual(agreement, a1) {
		t.Errorf("Expected: %v. Actual: %v", p1, agreement)
	}
}

func TestGetAgreementNotExists(t *testing.T) {
	req, _ := http.NewRequest("GET", "/agreements/doesnotexist", nil)
	res := request(req)
	checkError(t, res, http.StatusNotFound, res.Code)
}

func prepareCreateAgreement() {
	_, err := repo.GetProvider("01")
	if err != nil {
		repo.CreateProvider(&p1)
	}
}

func TestCreateAgreementThatExists(t *testing.T) {
	prepareCreateAgreement()
	body, err := json.Marshal(a1)
	if err != nil {
		t.Error("Unexpected marshalling error")
	}
	req, _ := http.NewRequest("POST", "/agreements", bytes.NewBuffer(body))
	res := request(req)

	checkStatus(t, http.StatusConflict, res.Code)
}

func TestCreateAgreementWrongProvider(t *testing.T) {
	prepareCreateAgreement()
	posted := createAgreement("02", "02", "02", "Agreement 02")
	body, err := json.Marshal(posted)
	if err != nil {
		t.Error("Unexpected marshalling error")
	}
	req, _ := http.NewRequest("POST", "/agreements", bytes.NewBuffer(body))
	res := request(req)

	checkStatus(t, http.StatusNotFound, res.Code)
}

func TestCreateAgreement(t *testing.T) {
	posted := createAgreement("02", "01", "02", "Agreement 02")
	body, err := json.Marshal(posted)
	if err != nil {
		t.Error("Unexpected marshalling error")
	}
	req, _ := http.NewRequest("POST", "/agreements", bytes.NewBuffer(body))
	res := request(req)

	checkStatus(t, http.StatusCreated, res.Code)

	var created model.Agreement
	_ = json.NewDecoder(res.Body).Decode(&created)
	if reflect.DeepEqual(created, posted) {
		t.Errorf("Expected: %v. Actual: %v", posted, created)
	}
}

func TestStartAgreementNotExist(t *testing.T) {
	req, _ := http.NewRequest("PUT", "/agreements/doesnotexist/start", nil)
	res := request(req)

	checkStatus(t, http.StatusNotFound, res.Code)
}

func TestStartAgreementExist(t *testing.T) {
	agreement, _ := repo.GetAgreement("01")

	req, _ := http.NewRequest("PUT", "/agreements/01/start", nil)
	res := request(req)

	checkStatus(t, http.StatusNoContent, res.Code)

	agreement, _ = repo.GetAgreement("01")
	if !agreement.Active {
		t.Error("Expected active agreement but it's not")
	}
}

func TestStopAgreementNotExist(t *testing.T) {
	req, _ := http.NewRequest("PUT", "/agreements/doesnotexist/stop", nil)
	res := request(req)

	checkStatus(t, http.StatusNotFound, res.Code)
}

func TestStopAgreementExist(t *testing.T) {
	req, _ := http.NewRequest("PUT", "/agreements/01/stop", nil)
	res := request(req)

	checkStatus(t, http.StatusNoContent, res.Code)

	agreement, _ := repo.GetAgreement("01")
	if agreement.Active {
		t.Error("Expected inactive agreement but it's active")
	}
}

func TestDeleteAgreementThatNotExists(t *testing.T) {
	req, _ := http.NewRequest("DELETE", "/agreements/doesnotexist", nil)
	res := request(req)

	checkStatus(t, http.StatusNotFound, res.Code)
	// TODO Check body
}

func TestDeleteAgreement(t *testing.T) {
	req, _ := http.NewRequest("DELETE", "/agreements/01", nil)
	res := request(req)

	checkStatus(t, http.StatusNoContent, res.Code)
	body, _ := ioutil.ReadAll(res.Body)

	if len(body) > 0 {
		t.Errorf("Expected empty body. Actual: %s", body)
	}
}

func TestEvaluationSuccess(t *testing.T) {

	data := map[string]map[string]interface{}{
		"TestGuarantee": map[string]interface{}{
			"test_value": 11,
		},
	}

	failed, err := evaluateAgreement(a1, data)
	if err != nil {
		t.Errorf("Error evaluating agreement: %s", err.Error())
	}

	if len(failed) > 0 {
		t.Errorf("Found penalties but none were expected")
	}
}

func TestEvaluationFailure(t *testing.T) {

	data := map[string]map[string]interface{}{
		"TestGuarantee": map[string]interface{}{
			"test_value": 9,
		},
	}

	failed, err := evaluateAgreement(a1, data)
	if err != nil {
		t.Errorf("Error evaluating agreement: %s", err.Error())
	}

	if len(failed) != 1 {
		t.Errorf("Penalty expected but none found")
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
	return providerPrefix + "_" + strconv.Itoa(i)
}

func createAgreement(aid, pid, cid, name string) model.Agreement {
	return model.Agreement{Id: aid, Name: name, Type: "Agreement", Active: false,
		Provider: model.Provider{Id: pid}, Client: model.Provider{Id: cid},
		Creation:   time.Now(),
		Expiration: time.Now().Add(24 * time.Hour),
		Guarantees: []model.Guarantee{
			model.Guarantee{Name: "TestGuarantee", Constraint: "[test_value] > 10"},
		},
	}
}
