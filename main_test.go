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
	"SLALite/repositories/memrepository"
	"SLALite/repositories/mongodb"
	"SLALite/repositories/validation"
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
	"strings"
	"testing"
	"time"
	"github.com/spf13/viper"
)

var a App
var repo model.IRepository
var p1 = model.Provider{Id: "p01", Name: "Provider01"}
var p2 = model.Provider{Id: "p02", Name: "Provider02"}
var c2 = model.Provider{Id: "c02", Name: "A client"}
var pdelete = model.Provider{Id: "pdelete", Name: "Removable provider"}
var dbName = "test.db"
var providerPrefix = "pf_" + strconv.Itoa(rand.Int())
var agreementPrefix = "apf_" + strconv.Itoa(rand.Int())

var a1 = createAgreement("a01", p1, c2, "Agreement 01")

// TestMain runs the tests
func TestMain(m *testing.M) {
	envvar := "SLA_" + strings.ToUpper(repositoryTypePropertyName)
	repotype, ok := os.LookupEnv(envvar)
	if !ok {
		repotype = defaultRepositoryType
	}
	repo = createRepository(repotype)
	if repo != nil {
		_, err := repo.CreateProvider(&p1)
		if err == nil {
			_, err = repo.CreateProvider(&pdelete)
		}
		if err == nil {
			_, err = repo.CreateAgreement(&a1)
		}
		if err != nil {
			log.Fatalf("Error creating initial state: %v", err)
		}
		a, _ = NewApp(viper.New(), repo)
	} else {
		log.Fatal("Error initializing repository")
	}

	result := m.Run()

	//BBolt clear database
	os.Remove(dbName)

	os.Exit(result)
}

func createRepository(repoType string) model.IRepository {
	var repo model.IRepository

	switch repoType {
	case defaultRepositoryType:
		memrepo, _ := memrepository.New(nil)
		repo = memrepo
	case "mongodb":
		config, _ := mongodb.NewDefaultConfig()
		config.Set("database", "slaliteTest")
		config.Set("clear_on_boot", true)
		mongoRepo, errMongo := mongodb.New(config)
		if errMongo != nil {
			log.Fatal("Error creating mongo repository: ", errMongo.Error())
		}
		repo = mongoRepo
	}
	repo, _ = validation.New(repo)
	return repo
}

func TestProviders(t *testing.T) {
	t.Run("GetProviders", testGetProviders)
	t.Run("GetProviderExists", testGetProviderExists)
	t.Run("GetProviderNotExists", testGetProviderNotExists)
	t.Run("CreateProviderThatExists", testCreateProviderThatExists)
	t.Run("CreateProvider", testCreateProvider)
	t.Run("DeleteProviderThatNotExists", testDeleteProviderThatNotExists)
	t.Run("DeleteProvider", testDeleteProvider)
	t.Run("Issue7 - Create provider with wrong input", testCreateProviderWithWrongInput)
}

func testGetProviders(t *testing.T) {
	req, _ := http.NewRequest("GET", "/providers", nil)
	res := request(req)
	checkStatus(t, http.StatusOK, res.Code)

	var providers model.Providers
	_ = json.NewDecoder(res.Body).Decode(&providers)
	if len(providers) != 2 {
		t.Errorf("Expected 2 provider. Received: %v", providers)
	}
}

func testGetProviderExists(t *testing.T) {
	req, _ := http.NewRequest("GET", "/providers/p01", nil)
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

func testGetProviderNotExists(t *testing.T) {
	req, _ := http.NewRequest("GET", "/providers/doesnotexist", nil)
	res := request(req)
	checkError(t, res, http.StatusNotFound, res.Code)

}

func testCreateProviderThatExists(t *testing.T) {
	body, err := json.Marshal(p1)
	if err != nil {
		t.Error("Unexpected marshalling error")
	}
	req, _ := http.NewRequest("POST", "/providers", bytes.NewBuffer(body))
	res := request(req)

	checkStatus(t, http.StatusConflict, res.Code)
}

func testCreateProvider(t *testing.T) {
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

func testCreateProviderWithWrongInput(t *testing.T) {
	body := "{\"id\": \"id\" \"name\": \"name\"}"         // note the missing ','
	req, _ := http.NewRequest("POST", "/providers", strings.NewReader(body))
	res := request(req)

	checkStatus(t, http.StatusBadRequest, res.Code)

	data := res.Body.Bytes()

	var restError ApiError

	/*
	 * Decode works! Using Unmarshal
	 */
	 err := json.Unmarshal(data, &restError)
	 //err := json.NewDecoder(res.Body).Decode(&restError)

	if err != nil {
		t.Errorf("Could not deserialize body request: %s", data)
	}
}

func testDeleteProviderThatNotExists(t *testing.T) {
	req, _ := http.NewRequest("DELETE", "/providers/doesnotexist", nil)
	res := request(req)

	checkStatus(t, http.StatusNotFound, res.Code)
	// TODO Check body
}

func testDeleteProvider(t *testing.T) {
	req, _ := http.NewRequest("DELETE", "/providers/pdelete", nil)
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

func TestAgreements(t *testing.T) {
	t.Run("GetAgreements", testGetAgreements)
	t.Run("GetActiveAgreements", testGetActiveAgreements)
	t.Run("GetAgreementExists", testGetAgreementExists)
	t.Run("GetAgreementNotExists", testGetAgreementNotExists)
	t.Run("CreateAgreementThatExists", testCreateAgreementThatExists)
	//t.Run("CreateAgreementWrongProvider", testCreateAgreementWrongProvider)
	t.Run("CreateAgreement", testCreateAgreement)
	t.Run("StartAgreementNotExist", testStartAgreementNotExist)
	t.Run("StartAgreementExist", testStartAgreementExist)
	t.Run("StopAgreementNotExist", testStopAgreementNotExist)
	t.Run("StopAgreementExist", testStopAgreementExist)
	t.Run("DeleteAgreementThatNotExists", testDeleteAgreementThatNotExists)
	t.Run("DeleteAgreement", testDeleteAgreement)
}

func testGetAgreements(t *testing.T) {
	req, _ := http.NewRequest("GET", "/agreements", nil)
	res := request(req)
	checkStatus(t, http.StatusOK, res.Code)

	var agreements model.Agreements
	_ = json.NewDecoder(res.Body).Decode(&agreements)
	if len(agreements) != 1 {
		t.Errorf("Expected 1 agreement. Received: %v", agreements)
	}
}

func testGetActiveAgreements(t *testing.T) {

	repo.StartAgreement("01")

	inactive := createAgreement("in1", p1, c2, "inactive")
	inactive.State = model.STOPPED

	repo.CreateAgreement(&inactive)

	expired := createAgreement("expired", p1, c2, "expired")
	expired.State = model.STARTED
	expired.Text.Expiration = time.Now().Add(-10 * time.Minute)

	repo.CreateAgreement(&expired)

	active := createAgreement("a_active", p1, c2, "active")
	active.State = model.STARTED
	repo.CreateAgreement(&active)

	as, _ := repo.GetAllAgreements()
	if len(as) != 4 {
		t.Fatalf("Cannot create initial conditions for test (wrong number of agreements). "+
			"Expected: %d. Actual: %d", 4, len(as))
	}

	req, _ := http.NewRequest("GET", "/agreements?active=true", nil)
	res := request(req)

	var agreements model.Agreements
	_ = json.NewDecoder(res.Body).Decode(&agreements)
	if len(agreements) != 1 {
		t.Errorf("Expected 1 agreement. Received: %v", agreements)
	}

	for _, agreement := range agreements {
		if !(agreement.Id == p1.Id || agreement.Id == active.Id) {
			t.Errorf("Got unexpected active agreement %s", agreement.Id)
		}
	}
}

func testGetAgreementExists(t *testing.T) {
	req, _ := http.NewRequest("GET", "/agreements/a01", nil)
	res := request(req)
	checkStatus(t, http.StatusOK, res.Code)
	/*
	 * Check body
	 */
	var agreement model.Agreement
	_ = json.NewDecoder(res.Body).Decode(&agreement)
	if reflect.DeepEqual(agreement, a1) {
		t.Errorf("Expected: %v. Actual: %v", a1, agreement)
	}
}

func testGetAgreementNotExists(t *testing.T) {
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

func testCreateAgreementThatExists(t *testing.T) {
	prepareCreateAgreement()
	body, err := json.Marshal(a1)
	if err != nil {
		t.Error("Unexpected marshalling error")
	}
	req, _ := http.NewRequest("POST", "/agreements", bytes.NewBuffer(body))
	res := request(req)

	checkStatus(t, http.StatusConflict, res.Code)
}

func testCreateAgreementWrongProvider(t *testing.T) {
	prepareCreateAgreement()
	posted := createAgreement("a02", p2, c2, "Agreement 02")
	body, err := json.Marshal(posted)
	if err != nil {
		t.Error("Unexpected marshalling error")
	}
	req, _ := http.NewRequest("POST", "/agreements", bytes.NewBuffer(body))
	res := request(req)

	checkStatus(t, http.StatusNotFound, res.Code)
}

func testCreateAgreement(t *testing.T) {
	posted := createAgreement("a02", p1, c2, "Agreement 02")
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

func testStartAgreementNotExist(t *testing.T) {
	req, _ := http.NewRequest("PUT", "/agreements/doesnotexist/start", nil)
	res := request(req)

	checkStatus(t, http.StatusNotFound, res.Code)
}

func testStartAgreementExist(t *testing.T) {
	req, _ := http.NewRequest("PUT", "/agreements/a01/start", nil)
	res := request(req)

	checkStatus(t, http.StatusNoContent, res.Code)

	agreement, _ := repo.GetAgreement("a01")
	if !agreement.IsStarted() {
		t.Error("Expected active agreement but it's not")
	}
}

func testStopAgreementNotExist(t *testing.T) {
	req, _ := http.NewRequest("PUT", "/agreements/doesnotexist/stop", nil)
	res := request(req)

	checkStatus(t, http.StatusNotFound, res.Code)
}

func testStopAgreementExist(t *testing.T) {
	req, _ := http.NewRequest("PUT", "/agreements/a01/stop", nil)
	res := request(req)

	checkStatus(t, http.StatusNoContent, res.Code)

	agreement, _ := repo.GetAgreement("a01")
	if !agreement.IsStopped() {
		t.Error("Expected inactive agreement but it's active")
	}
}

func testDeleteAgreementThatNotExists(t *testing.T) {
	req, _ := http.NewRequest("DELETE", "/agreements/doesnotexist", nil)
	res := request(req)

	checkStatus(t, http.StatusNotFound, res.Code)
	// TODO Check body
}

func testDeleteAgreement(t *testing.T) {
	req, _ := http.NewRequest("DELETE", "/agreements/a01", nil)
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

func createAgreement(aid string, provider, client model.Provider, name string) model.Agreement {
	return model.Agreement{
		Id:    aid,
		Name:  name,
		State: model.STOPPED,
		Text: model.AgreementText{
			Id:       aid,
			Name:     name,
			Type:     model.AGREEMENT,
			Provider: provider, Client: client,
			Creation:   time.Now(),
			Expiration: time.Now().Add(24 * time.Hour),
			Guarantees: []model.Guarantee{
				model.Guarantee{Name: "TestGuarantee", Constraint: "[test_value] > 10"},
			},
		},
	}
}
