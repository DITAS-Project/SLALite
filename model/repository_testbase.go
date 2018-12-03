/*
Copyright 2018 Atos

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

package model

import (
	"bytes"
	"fmt"
	"runtime/debug"
	"testing"
	"time"
)

type TestData struct {
	P01        Provider
	P02        Provider
	Pnotexists Provider
	A01        Agreement
	A02        Agreement
	A03        Agreement
	Anotexists Agreement
	V01        Violation
	Vnotexists Violation
}

var Data = TestData{
	P01: Provider{
		Id:   "p01",
		Name: "Provider01",
	},
	P02: Provider{
		Id:   "p02",
		Name: "Provider02",
	},
	Pnotexists: Provider{
		Id:   "notexists",
		Name: "ProviderNotExists",
	},
	A01: Agreement{
		Id:         "a01",
		Name:       "Agreement01",
		State:      STOPPED,
		Assessment: &Assessment{},
	},
	A02: Agreement{
		Id:         "a02",
		Name:       "Agreement02",
		State:      STARTED,
		Assessment: &Assessment{},
	},
	A03: Agreement{
		Id:         "a03",
		Name:       "Agreement03",
		State:      TERMINATED,
		Assessment: &Assessment{},
	},
	Anotexists: Agreement{
		Id:         "notexists",
		Name:       "AgreementNotExists",
		State:      STOPPED,
		Assessment: &Assessment{},
	},
	V01: Violation{
		Id:          "v01",
		AgreementId: "a01",
		Datetime:    time.Now(),
		Constraint:  "t < 100",
		Guarantee:   "gt1",
		Values: []MetricValue{
			MetricValue{DateTime: time.Now(), Key: "t", Value: 101},
		},
	},
	Vnotexists: Violation{
		Id:          "vnotexists",
		AgreementId: "a01",
	},
}

// CheckSetup checks that the entities to be created on this test do not exist in the
// repository (this would make the test fail). To be called from TestMain method.
func CheckSetup(repo IRepository) error {

	providers := []string{Data.P01.Id, Data.P02.Id}
	agreements := []string{Data.A01.Id, Data.A02.Id}

	var id string
	var err error
	for _, id = range providers {
		if _, err = repo.GetProvider(id); err != ErrNotFound {
			return fmt.Errorf("Provider[%s] exists or err[%v]", id, err)
		}
	}
	for _, id = range agreements {
		if _, err = repo.GetAgreement(id); err != ErrNotFound {
			return fmt.Errorf("Agreement[%s] exists or err[%v]", id, err)
		}
	}
	if _, err = repo.GetViolation(Data.V01.Id); err != ErrNotFound {
		return fmt.Errorf("Violation[%s] exists or err[%v]", Data.V01.Id, err)
	}
	return nil
}

// TestCreateProvider executes this test
func TestCreateProvider(t *testing.T, repo IRepository) {

	_, err := repo.CreateProvider(&Data.P01)
	assertEquals(t, "Unexpected error. Expected: %v; Actual: %v", nil, err)

	repo.CreateProvider(&Data.P02)
	assertEquals(t, "Unexpected error. Expected: %v; Actual: %v", nil, err)
}

// TestCreateProviderExists executes this test
func TestCreateProviderExists(t *testing.T, repo IRepository) {
	_, err := repo.CreateProvider(&Data.P01)
	assertEquals(t, "Expected error: %v; actual: %v", ErrAlreadyExist, err)
}

// TestGetAllProviders executes this test
func TestGetAllProviders(t *testing.T, repo IRepository) {
	actual, err := repo.GetAllProviders()
	assertEquals(t, "Unexpected error. Expected: %v; Actual: %v", nil, err)
	assertEquals(t, "Unexpected len(providers). Expected: %d; Actual: %d", 2, len(actual))
}

// TestGetProvider executes this test
func TestGetProvider(t *testing.T, repo IRepository) {
	actual, err := repo.GetProvider(Data.P01.Id)
	assertEquals(t, "Unexpected error. Expected: %v; Actual: %v", nil, err)
	assertEquals(t, "Unexpected result. Expected: %v; Actual: %v", Data.P01.Id, actual.Id)
}

// TestGetProviderNotExists executes this test
func TestGetProviderNotExists(t *testing.T, repo IRepository) {
	_, err := repo.GetProvider("notexists")
	assertEquals(t, "Unexpected error. Expected: %v; Actual: %v", ErrNotFound, err)
}

// TestDeleteProvider executes this test
func TestDeleteProvider(t *testing.T, repo IRepository) {
	err := repo.DeleteProvider(&Data.P02)
	assertEquals(t, "Unexpected error. Expected: %v; Actual: %v", nil, err)
}

// TestDeleteProviderNotExists executes this test
func TestDeleteProviderNotExists(t *testing.T, repo IRepository) {
	err := repo.DeleteProvider(&Data.Pnotexists)
	assertEquals(t, "Expected error: %v; actual: %v", ErrNotFound, err)
}

// TestCreateAgreement executes this test
func TestCreateAgreement(t *testing.T, repo IRepository) {
	var a *Agreement
	var err error

	a, err = repo.CreateAgreement(&Data.A01)
	assertEquals(t, "Unexpected error. Expected: %v; Actual: %v", nil, err)
	Data.A01 = *a

	a, err = repo.CreateAgreement(&Data.A02)
	assertEquals(t, "Unexpected error. Expected: %v; Actual: %v", nil, err)
	Data.A02 = *a

	a, err = repo.CreateAgreement(&Data.A03)
	assertEquals(t, "Unexpected error. Expected: %v; Actual: %v", nil, err)
	Data.A03 = *a
}

// TestCreateAgreementExists executes this test
func TestCreateAgreementExists(t *testing.T, repo IRepository) {
	_, err := repo.CreateAgreement(&Data.A01)
	assertEquals(t, "Expected error: %v; actual: %v", ErrAlreadyExist, err)
}

// TestGetAllAgreements executes this test
func TestGetAllAgreements(t *testing.T, repo IRepository) {
	actual, err := repo.GetAllAgreements()
	assertEquals(t, "Unexpected error. Expected: %v; Actual: %v", nil, err)
	assertEquals(t, "Unexpected len(Agreements). Expected: %d; Actual: %d", 3, len(actual))
}

// TestGetAgreement executes this test
func TestGetAgreement(t *testing.T, repo IRepository) {
	result, err := repo.GetAgreement(Data.A01.Id)
	assertEquals(t, "Unexpected error. Expected: %v; Actual: %v", nil, err)
	assertEquals(t, "Unexpected result. Expected: %v; Actual: %v", Data.A01.Id, result.Id)
}

// TestGetAgreementNotExists executes this test
func TestGetAgreementNotExists(t *testing.T, repo IRepository) {
	_, err := repo.GetAgreement("notexists")
	assertEquals(t, "Expected error: %v; actual: %v", ErrNotFound, err)
}

// TestGetAgreementsByState executes this test
func TestGetAgreementsByState(t *testing.T, repo IRepository) {
	actual, err := repo.GetAgreementsByState(STARTED)
	assertEquals(t, "Unexpected error. Expected: %v; Actual: %v", nil, err)
	assertEquals(t, "Get(STARTED). Unexpected len(agreements). Expected: %v; Actual: %v", 1, len(actual))
	assertEquals(t, "Unexpected state. Expected: %v; Actual: %v", STARTED, actual[0].State)

	actual, err = repo.GetAgreementsByState(STARTED, STOPPED)
	assertEquals(t, "Unexpected error. Expected: %v; Actual: %v", nil, err)
	assertEquals(t, "Get(STARTED, STOPPED). Unexpected len(agreements). Expected: %v; Actual: %v", 2, len(actual))
	for _, a := range actual {
		if a.State != STOPPED && a.State != STARTED {
			t.Errorf("Unexpected state. Expected: %v; Actual: %v", "STARTED||STOPPED", a.State)
		}
	}

	actual, err = repo.GetAgreementsByState(STARTED, STOPPED, TERMINATED)
	assertEquals(t, "Unexpected error. Expected: %v; Actual: %v", nil, err)
	assertEquals(t, "Get(STARTED, STOPPED, TERMINATED). Unexpected len(agreements). Expected: %v; Actual: %v", 3, len(actual))
}

// TestUpdateAgreementState executes this test
func TestUpdateAgreementState(t *testing.T, repo IRepository) {
	a, err := repo.UpdateAgreementState(Data.A02.Id, STOPPED)
	assertEquals(t, "Unexpected error. Expected: %v; Actual: %v", nil, err)
	a, err = repo.GetAgreement(Data.A02.Id)
	assertEquals(t, "Unexpected state. Expected: %v; Actual: %v", STOPPED, a.State)

	a, err = repo.UpdateAgreementState(Data.A02.Id, TERMINATED)
	assertEquals(t, "Unexpected error. Expected: %v; Actual: %v", nil, err)
	a, err = repo.GetAgreement(Data.A02.Id)
	assertEquals(t, "Unexpected state. Expected: %v; Actual: %v", TERMINATED, a.State)

	a, err = repo.UpdateAgreementState(Data.A02.Id, STARTED)
	assertEquals(t, "Unexpected error. Expected: %v; Actual: %v", nil, err)
	a, err = repo.GetAgreement(Data.A02.Id)
	assertEquals(t, "Unexpected state. Expected: %v; Actual: %v", STARTED, a.State)
}

// TestUpdateAgreementStateNotExists executes this test
func TestUpdateAgreementStateNotExists(t *testing.T, repo IRepository) {
	_, err := repo.UpdateAgreementState(Data.Anotexists.Id, STOPPED)
	assertEquals(t, "Unexpected error. Expected: %v; Actual: %v", ErrNotFound, err)
}

// TestUpdateAgreement executes this test
func TestUpdateAgreement(t *testing.T, repo IRepository) {
	now := time.Now()
	Data.A02.State = STOPPED
	Data.A02.Assessment.FirstExecution = now

	a, err := repo.UpdateAgreement(&Data.A02)
	assertEquals(t, "Unexpected error. Expected: %v; Actual: %v", nil, err)

	a, err = repo.GetAgreement(Data.A02.Id)
	assertEquals(t, "Unexpected error. Expected: %v; Actual: %v", nil, err)
	assertEquals(t, "Unexpected state. Expected: %v; Actual: %v", STOPPED, a.State)
	assertEquals(t, "Unexpected Assessment.FirstExecution. Expected: %v; Actual: %v",
		now.Unix(), a.Assessment.FirstExecution.Unix())
}

// TestUpdateAgreementNotExists executes this test
func TestUpdateAgreementNotExists(t *testing.T, repo IRepository) {
	Data.Anotexists.State = STOPPED

	_, err := repo.UpdateAgreement(&Data.Anotexists)
	assertEquals(t, "Unexpected error. Expected: %v; Actual: %v", ErrNotFound, err)
}

// TestDeleteAgreement executes this test
func TestDeleteAgreement(t *testing.T, repo IRepository) {
	err := repo.DeleteAgreement(&Data.A02)
	assertEquals(t, "Unexpected error. Expected: %v; Actual: %v", nil, err)
}

// TestDeleteAgreementNotExists executes this test
func TestDeleteAgreementNotExists(t *testing.T, repo IRepository) {
	err := repo.DeleteAgreement(&Data.Anotexists)
	assertEquals(t, "Unexpected error. Expected: %v; Actual: %v", ErrNotFound, err)
}

// TestCreateViolation executes this test
func TestCreateViolation(t *testing.T, repo IRepository) {
	// When on externalId repo, we have to sync v.AgreementId
	Data.V01.AgreementId = Data.A01.Id
	v, err := repo.CreateViolation(&Data.V01)
	Data.V01 = *v
	assertEquals(t, "Unexpected error. Expected: %v; Actual: %v", nil, err)

	v, err = repo.GetViolation(Data.V01.Id)
	assertEquals(t, "Unexpected error. Expected: %v; Actual: %v", nil, err)
	assertEquals(t, "Unexpected violation. Expected: %v; Actual: %v", Data.V01.Id, v.Id)
}

// TestCreateViolationExists executes this test
func TestCreateViolationExists(t *testing.T, repo IRepository) {
	_, err := repo.CreateViolation(&Data.V01)
	assertEquals(t, "Unexpected error. Expected: %v; Actual: %v", ErrAlreadyExist, err)
}

// TestGetViolation executes this test
func TestGetViolation(t *testing.T, repo IRepository) {
	v, err := repo.GetViolation(Data.V01.Id)
	assertEquals(t, "Unexpected error. Expected: %v; Actual: %v", nil, err)
	assertEquals(t, "Unexpected violation. Expected: %v; Actual: %v", Data.V01.Id, v.Id)
}

// TestGetViolationNotExists executes this test
func TestGetViolationNotExists(t *testing.T, repo IRepository) {
	_, err := repo.GetViolation(Data.Vnotexists.Id)
	assertEquals(t, "Unexpected error. Expected: %v; Actual: %v", ErrNotFound, err)
}

func assertEquals(t *testing.T, msg string, expected interface{}, actual interface{}) {
	if expected != actual {
		buf := bytes.Buffer{}
		buf.Write(debug.Stack())
		t.Errorf(msg+"\n%s", expected, actual, buf.String())

	}
}
