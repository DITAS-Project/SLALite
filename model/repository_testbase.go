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
	"testing"
	"time"
)

var p01 = Provider{
	Id:   "p01",
	Name: "Provider01",
}

var p02 = Provider{
	Id:   "p02",
	Name: "Provider02",
}

var pnotexists = Provider{
	Id:   "notexists",
	Name: "ProviderNotExists",
}

var a01 = Agreement{
	Id:         "a01",
	Name:       "Agreement01",
	State:      STOPPED,
	Assessment: Assessment{},
}

var a02 = Agreement{
	Id:         "a02",
	Name:       "Agreement02",
	State:      STARTED,
	Assessment: Assessment{},
}

var a03 = Agreement{
	Id:         "a03",
	Name:       "Agreement03",
	State:      TERMINATED,
	Assessment: Assessment{},
}

var anotexists = Agreement{
	Id:         "notexists",
	Name:       "AgreementNotExists",
	State:      STOPPED,
	Assessment: Assessment{},
}

var v01 = Violation{
	Id:          "v01",
	AgreementId: "a01",
}

var vnotexists = Violation{
	Id:          "vnotexists",
	AgreementId: "a01",
}

// TestCreateProvider executes this test
func TestCreateProvider(t *testing.T, repo IRepository) {

	_, err := repo.CreateProvider(&p01)
	assertEquals(t, "Unexpected error. Expected: %v; Actual: %v", nil, err)

	repo.CreateProvider(&p02)
	assertEquals(t, "Unexpected error. Expected: %v; Actual: %v", nil, err)
}

// TestCreateProviderExists executes this test
func TestCreateProviderExists(t *testing.T, repo IRepository) {
	_, err := repo.CreateProvider(&p01)
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
	actual, err := repo.GetProvider(p01.Id)
	assertEquals(t, "Unexpected error. Expected: %v; Actual: %v", nil, err)
	assertEquals(t, "Unexpected result. Expected: %v; Actual: %v", p01.Id, actual.Id)
}

// TestGetProviderNotExists executes this test
func TestGetProviderNotExists(t *testing.T, repo IRepository) {
	_, err := repo.GetProvider("notexists")
	assertEquals(t, "Unexpected error. Expected: %v; Actual: %v", ErrNotFound, err)
}

// TestDeleteProvider executes this test
func TestDeleteProvider(t *testing.T, repo IRepository) {
	err := repo.DeleteProvider(&p02)
	assertEquals(t, "Unexpected error. Expected: %v; Actual: %v", nil, err)
}

// TestDeleteProviderNotExists executes this test
func TestDeleteProviderNotExists(t *testing.T, repo IRepository) {
	err := repo.DeleteProvider(&pnotexists)
	assertEquals(t, "Expected error: %v; actual: %v", ErrNotFound, err)
}

// TestCreateAgreement executes this test
func TestCreateAgreement(t *testing.T, repo IRepository) {

	_, err := repo.CreateAgreement(&a01)
	assertEquals(t, "Unexpected error. Expected: %v; Actual: %v", nil, err)

	repo.CreateAgreement(&a02)
	assertEquals(t, "Unexpected error. Expected: %v; Actual: %v", nil, err)

	repo.CreateAgreement(&a03)
	assertEquals(t, "Unexpected error. Expected: %v; Actual: %v", nil, err)
}

// TestCreateAgreementExists executes this test
func TestCreateAgreementExists(t *testing.T, repo IRepository) {
	_, err := repo.CreateAgreement(&a01)
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
	result, err := repo.GetAgreement(a01.Id)
	assertEquals(t, "Unexpected error. Expected: %v; Actual: %v", nil, err)
	assertEquals(t, "Unexpected result. Expected: %v; Actual: %v", a01.Id, result.Id)
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
	a, err := repo.UpdateAgreementState(a02.Id, STOPPED)
	assertEquals(t, "Unexpected error. Expected: %v; Actual: %v", nil, err)
	a, err = repo.GetAgreement(a02.Id)
	assertEquals(t, "Unexpected state. Expected: %v; Actual: %v", a.State, STOPPED)

	a, err = repo.UpdateAgreementState(a02.Id, TERMINATED)
	assertEquals(t, "Unexpected error. Expected: %v; Actual: %v", nil, err)
	a, err = repo.GetAgreement(a02.Id)
	assertEquals(t, "Unexpected state. Expected: %v; Actual: %v", a.State, TERMINATED)

	a, err = repo.UpdateAgreementState(a02.Id, STARTED)
	assertEquals(t, "Unexpected error. Expected: %v; Actual: %v", nil, err)
	a, err = repo.GetAgreement(a02.Id)
	assertEquals(t, "Unexpected state. Expected: %v; Actual: %v", a.State, STARTED)
}

// TestUpdateAgreementStateNotExists executes this test
func TestUpdateAgreementStateNotExists(t *testing.T, repo IRepository) {
	_, err := repo.UpdateAgreementState(anotexists.Id, STOPPED)
	assertEquals(t, "Unexpected error. Expected: %v; Actual: %v", ErrNotFound, err)
}

// TestUpdateAgreement executes this test
func TestUpdateAgreement(t *testing.T, repo IRepository) {
	now := time.Now()
	a02.State = STOPPED
	a02.Assessment.FirstExecution = now

	a, err := repo.UpdateAgreement(&a02)
	assertEquals(t, "Unexpected error. Expected: %v; Actual: %v", nil, err)

	a, err = repo.GetAgreement(a02.Id)
	assertEquals(t, "Unexpected error. Expected: %v; Actual: %v", nil, err)
	assertEquals(t, "Unexpected state. Expected: %v; Actual: %v", STOPPED, a.State)
	assertEquals(t, "Unexpected Assessment.FirstExceution. Expected: %v; Actual: %v",
		now, a.Assessment.FirstExecution)
}

// TestUpdateAgreementNotExists executes this test
func TestUpdateAgreementNotExists(t *testing.T, repo IRepository) {
	anotexists.State = STOPPED

	_, err := repo.UpdateAgreement(&anotexists)
	assertEquals(t, "Unexpected error. Expected: %v; Actual: %v", ErrNotFound, err)
}

// TestDeleteAgreement executes this test
func TestDeleteAgreement(t *testing.T, repo IRepository) {
	err := repo.DeleteAgreement(&a02)
	assertEquals(t, "Unexpected error. Expected: %v; Actual: %v", nil, err)
}

// TestDeleteAgreementNotExists executes this test
func TestDeleteAgreementNotExists(t *testing.T, repo IRepository) {
	err := repo.DeleteAgreement(&anotexists)
	assertEquals(t, "Unexpected error. Expected: %v; Actual: %v", ErrNotFound, err)
}

// TestCreateViolation executes this test
func TestCreateViolation(t *testing.T, repo IRepository) {
	v, err := repo.CreateViolation(&v01)
	assertEquals(t, "Unexpected error. Expected: %v; Actual: %v", nil, err)

	v, err = repo.GetViolation(v01.Id)
	assertEquals(t, "Unexpected error. Expected: %v; Actual: %v", nil, err)
	assertEquals(t, "Unexpected violation. Expected: %v; Actual: %v", v01.Id, v.Id)
}

// TestCreateViolationExists executes this test
func TestCreateViolationExists(t *testing.T, repo IRepository) {
	_, err := repo.CreateViolation(&v01)
	assertEquals(t, "Unexpected error. Expected: %v; Actual: %v", ErrAlreadyExist, err)
}

// TestGetViolation executes this test
func TestGetViolation(t *testing.T, repo IRepository) {
	v, err := repo.GetViolation(v01.Id)
	assertEquals(t, "Unexpected error. Expected: %v; Actual: %v", nil, err)
	assertEquals(t, "Unexpected violation. Expected: %v; Actual: %v", v01.Id, v.Id)
}

// TestGetViolationNotExists executes this test
func TestGetViolationNotExists(t *testing.T, repo IRepository) {
	_, err := repo.GetViolation(vnotexists.Id)
	assertEquals(t, "Unexpected error. Expected: %v; Actual: %v", ErrNotFound, err)
}

func assertEquals(t *testing.T, msg string, expected interface{}, actual interface{}) {
	if expected != actual {
		t.Errorf(msg, expected, actual)
	}
}
