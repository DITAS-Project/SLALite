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

package repositories

import (
	"SLALite/model"
	"bytes"
	"fmt"
	"runtime/debug"
	"testing"
	"time"
)

// TestData contains the data type to be used in these tests
type TestData struct {
	P01        model.Provider
	P02        model.Provider
	Pnotexists model.Provider
	A01        model.Agreement
	A02        model.Agreement
	A03        model.Agreement
	Anotexists model.Agreement
	V01        model.Violation
	Vnotexists model.Violation
	T01        model.Template
}

// Data contains the data to be used in these tests. It can be overwritten if needed.
var Data = TestData{
	P01: model.Provider{
		Id:   "p01",
		Name: "Provider01",
	},
	P02: model.Provider{
		Id:   "p02",
		Name: "Provider02",
	},
	Pnotexists: model.Provider{
		Id:   "notexists",
		Name: "ProviderNotExists",
	},
	A01: model.Agreement{
		Id:         "a01",
		Name:       "Agreement01",
		State:      model.STOPPED,
		Assessment: model.Assessment{},
	},
	A02: model.Agreement{
		Id:         "a02",
		Name:       "Agreement02",
		State:      model.STARTED,
		Assessment: model.Assessment{},
	},
	A03: model.Agreement{
		Id:         "a03",
		Name:       "Agreement03",
		State:      model.TERMINATED,
		Assessment: model.Assessment{},
	},
	Anotexists: model.Agreement{
		Id:         "notexists",
		Name:       "AgreementNotExists",
		State:      model.STOPPED,
		Assessment: model.Assessment{},
	},
	V01: model.Violation{
		Id:          "v01",
		AgreementId: "a01",
		Datetime:    time.Now(),
		Constraint:  "t < 100",
		Guarantee:   "gt1",
		Values: []model.MetricValue{
			model.MetricValue{DateTime: time.Now(), Key: "t", Value: 101},
		},
	},
	Vnotexists: model.Violation{
		Id:          "vnotexists",
		AgreementId: "a01",
	},
	T01: model.Template{
		Id:   "t01",
		Name: "Template01",
	},
}

// CheckSetup checks that the entities to be created on this test do not exist in the
// repository (this would make the test fail). To be called from TestMain method.
func CheckSetup(repo model.IRepository) error {

	providers := []string{Data.P01.Id, Data.P02.Id}
	agreements := []string{Data.A01.Id, Data.A02.Id}

	var id string
	var err error
	for _, id = range providers {
		if _, err = repo.GetProvider(id); err != model.ErrNotFound {
			return fmt.Errorf("Provider[%s] exists or err[%v]", id, err)
		}
	}
	for _, id = range agreements {
		if _, err = repo.GetAgreement(id); err != model.ErrNotFound {
			return fmt.Errorf("Agreement[%s] exists or err[%v]", id, err)
		}
	}
	if _, err = repo.GetViolation(Data.V01.Id); err != model.ErrNotFound {
		return fmt.Errorf("Violation[%s] exists or err[%v]", Data.V01.Id, err)
	}
	return nil
}

/*
TestContext contains parameters needed in a repository test when calling from
a t.Run() statement

This way, t.Run() can be called like t.Run("Test", ctx.testSomething).
*/
type TestContext struct {
	Repo model.IRepository
}

// TestCreateProvider executes this test
func (r *TestContext) TestCreateProvider(t *testing.T) {
	_, err := r.Repo.CreateProvider(&Data.P01)
	assertEquals(t, "Unexpected error. Expected: %v; Actual: %v", nil, err)

	r.Repo.CreateProvider(&Data.P02)
	assertEquals(t, "Unexpected error. Expected: %v; Actual: %v", nil, err)
}

// TestCreateProviderExists executes this test
func (r *TestContext) TestCreateProviderExists(t *testing.T) {
	_, err := r.Repo.CreateProvider(&Data.P01)
	assertEquals(t, "Expected error: %v; actual: %v", model.ErrAlreadyExist, err)
}

// TestGetAllProviders executes this test
func (r *TestContext) TestGetAllProviders(t *testing.T) {
	actual, err := r.Repo.GetAllProviders()
	assertEquals(t, "Unexpected error. Expected: %v; Actual: %v", nil, err)
	assertEquals(t, "Unexpected len(providers). Expected: %d; Actual: %d", 2, len(actual))
}

// TestGetProvider executes this test
func (r *TestContext) TestGetProvider(t *testing.T) {
	actual, err := r.Repo.GetProvider(Data.P01.Id)
	assertEquals(t, "Unexpected error. Expected: %v; Actual: %v", nil, err)
	assertEquals(t, "Unexpected result. Expected: %v; Actual: %v", Data.P01.Id, actual.Id)
}

// TestGetProviderNotExists executes this test
func (r *TestContext) TestGetProviderNotExists(t *testing.T) {
	_, err := r.Repo.GetProvider("notexists")
	assertEquals(t, "Unexpected error. Expected: %v; Actual: %v", model.ErrNotFound, err)
}

// TestDeleteProvider executes this test
func (r *TestContext) TestDeleteProvider(t *testing.T) {
	err := r.Repo.DeleteProvider(&Data.P02)
	assertEquals(t, "Unexpected error. Expected: %v; Actual: %v", nil, err)
}

// TestDeleteProviderNotExists executes this test
func (r *TestContext) TestDeleteProviderNotExists(t *testing.T) {
	err := r.Repo.DeleteProvider(&Data.Pnotexists)
	assertEquals(t, "Expected error: %v; actual: %v", model.ErrNotFound, err)
}

// TestCreateAgreement executes this test
func (r *TestContext) TestCreateAgreement(t *testing.T) {
	var a *model.Agreement
	var err error

	a, err = r.Repo.CreateAgreement(&Data.A01)
	assertEquals(t, "Unexpected error. Expected: %v; Actual: %v", nil, err)
	Data.A01 = *a

	a, err = r.Repo.CreateAgreement(&Data.A02)
	assertEquals(t, "Unexpected error. Expected: %v; Actual: %v", nil, err)
	Data.A02 = *a

	a, err = r.Repo.CreateAgreement(&Data.A03)
	assertEquals(t, "Unexpected error. Expected: %v; Actual: %v", nil, err)
	Data.A03 = *a
}

// TestCreateAgreementExists executes this test
func (r *TestContext) TestCreateAgreementExists(t *testing.T) {
	_, err := r.Repo.CreateAgreement(&Data.A01)
	assertEquals(t, "Expected error: %v; actual: %v", model.ErrAlreadyExist, err)
}

// TestGetAllAgreements executes this test
func (r *TestContext) TestGetAllAgreements(t *testing.T) {
	actual, err := r.Repo.GetAllAgreements()
	assertEquals(t, "Unexpected error. Expected: %v; Actual: %v", nil, err)
	assertEquals(t, "Unexpected len(Agreements). Expected: %d; Actual: %d", 3, len(actual))
}

// TestGetAgreement executes this test
func (r *TestContext) TestGetAgreement(t *testing.T) {
	result, err := r.Repo.GetAgreement(Data.A01.Id)
	assertEquals(t, "Unexpected error. Expected: %v; Actual: %v", nil, err)
	assertEquals(t, "Unexpected result. Expected: %v; Actual: %v", Data.A01.Id, result.Id)
}

// TestGetAgreementNotExists executes this test
func (r *TestContext) TestGetAgreementNotExists(t *testing.T) {
	_, err := r.Repo.GetAgreement("notexists")
	assertEquals(t, "Expected error: %v; actual: %v", model.ErrNotFound, err)
}

// TestGetAgreementsByState executes this test
func (r *TestContext) TestGetAgreementsByState(t *testing.T) {
	actual, err := r.Repo.GetAgreementsByState(model.STARTED)
	assertEquals(t, "Unexpected error. Expected: %v; Actual: %v", nil, err)
	assertEquals(t, "Get(STARTED). Unexpected len(agreements). Expected: %v; Actual: %v", 1, len(actual))
	assertEquals(t, "Unexpected state. Expected: %v; Actual: %v", model.STARTED, actual[0].State)

	actual, err = r.Repo.GetAgreementsByState(model.STARTED, model.STOPPED)
	assertEquals(t, "Unexpected error. Expected: %v; Actual: %v", nil, err)
	assertEquals(t, "Get(STARTED, STOPPED). Unexpected len(agreements). Expected: %v; Actual: %v", 2, len(actual))
	for _, a := range actual {
		if a.State != model.STOPPED && a.State != model.STARTED {
			t.Errorf("Unexpected state. Expected: %v; Actual: %v", "STARTED||STOPPED", a.State)
		}
	}

	actual, err = r.Repo.GetAgreementsByState(model.STARTED, model.STOPPED, model.TERMINATED)
	assertEquals(t, "Unexpected error. Expected: %v; Actual: %v", nil, err)
	assertEquals(t, "Get(STARTED, STOPPED, TERMINATED). Unexpected len(agreements). Expected: %v; Actual: %v", 3, len(actual))
}

// TestUpdateAgreementState executes this test
func (r *TestContext) TestUpdateAgreementState(t *testing.T) {
	a, err := r.Repo.UpdateAgreementState(Data.A02.Id, model.STOPPED)
	assertEquals(t, "Unexpected error. Expected: %v; Actual: %v", nil, err)
	a, err = r.Repo.GetAgreement(Data.A02.Id)
	assertEquals(t, "Unexpected state. Expected: %v; Actual: %v", model.STOPPED, a.State)

	a, err = r.Repo.UpdateAgreementState(Data.A02.Id, model.TERMINATED)
	assertEquals(t, "Unexpected error. Expected: %v; Actual: %v", nil, err)
	a, err = r.Repo.GetAgreement(Data.A02.Id)
	assertEquals(t, "Unexpected state. Expected: %v; Actual: %v", model.TERMINATED, a.State)

	a, err = r.Repo.UpdateAgreementState(Data.A02.Id, model.STARTED)
	assertEquals(t, "Unexpected error. Expected: %v; Actual: %v", nil, err)
	a, err = r.Repo.GetAgreement(Data.A02.Id)
	assertEquals(t, "Unexpected state. Expected: %v; Actual: %v", model.STARTED, a.State)
}

// TestUpdateAgreementStateNotExists executes this test
func (r *TestContext) TestUpdateAgreementStateNotExists(t *testing.T) {
	_, err := r.Repo.UpdateAgreementState(Data.Anotexists.Id, model.STOPPED)
	assertEquals(t, "Unexpected error. Expected: %v; Actual: %v", model.ErrNotFound, err)
}

// TestUpdateAgreement executes this test
func (r *TestContext) TestUpdateAgreement(t *testing.T) {
	now := time.Now()
	Data.A02.State = model.STOPPED
	Data.A02.Assessment.FirstExecution = now

	a, err := r.Repo.UpdateAgreement(&Data.A02)
	assertEquals(t, "Unexpected error. Expected: %v; Actual: %v", nil, err)

	a, err = r.Repo.GetAgreement(Data.A02.Id)
	assertEquals(t, "Unexpected error. Expected: %v; Actual: %v", nil, err)
	assertEquals(t, "Unexpected state. Expected: %v; Actual: %v", model.STOPPED, a.State)
	assertEquals(t, "Unexpected Assessment.FirstExecution. Expected: %v; Actual: %v",
		now.Unix(), a.Assessment.FirstExecution.Unix())
}

// TestUpdateAgreementNotExists executes this test
func (r *TestContext) TestUpdateAgreementNotExists(t *testing.T) {
	Data.Anotexists.State = model.STOPPED

	_, err := r.Repo.UpdateAgreement(&Data.Anotexists)
	assertEquals(t, "Unexpected error. Expected: %v; Actual: %v", model.ErrNotFound, err)
}

// TestDeleteAgreement executes this test
func (r *TestContext) TestDeleteAgreement(t *testing.T) {
	err := r.Repo.DeleteAgreement(&Data.A02)
	assertEquals(t, "Unexpected error. Expected: %v; Actual: %v", nil, err)
}

// TestDeleteAgreementNotExists executes this test
func (r *TestContext) TestDeleteAgreementNotExists(t *testing.T) {
	err := r.Repo.DeleteAgreement(&Data.Anotexists)
	assertEquals(t, "Unexpected error. Expected: %v; Actual: %v", model.ErrNotFound, err)
}

// TestCreateViolation executes this test
func (r *TestContext) TestCreateViolation(t *testing.T) {
	// When on externalId repo, we have to sync v.AgreementId
	Data.V01.AgreementId = Data.A01.Id
	v, err := r.Repo.CreateViolation(&Data.V01)
	Data.V01 = *v
	assertEquals(t, "Unexpected error. Expected: %v; Actual: %v", nil, err)

	v, err = r.Repo.GetViolation(Data.V01.Id)
	assertEquals(t, "Unexpected error. Expected: %v; Actual: %v", nil, err)
	assertEquals(t, "Unexpected violation. Expected: %v; Actual: %v", Data.V01.Id, v.Id)
}

// TestCreateViolationExists executes this test
func (r *TestContext) TestCreateViolationExists(t *testing.T) {
	_, err := r.Repo.CreateViolation(&Data.V01)
	assertEquals(t, "Unexpected error. Expected: %v; Actual: %v", model.ErrAlreadyExist, err)
}

// TestGetViolation executes this test
func (r *TestContext) TestGetViolation(t *testing.T) {
	v, err := r.Repo.GetViolation(Data.V01.Id)
	assertEquals(t, "Unexpected error. Expected: %v; Actual: %v", nil, err)
	assertEquals(t, "Unexpected violation. Expected: %v; Actual: %v", Data.V01.Id, v.Id)
}

// TestGetViolationNotExists executes this test
func (r *TestContext) TestGetViolationNotExists(t *testing.T) {
	_, err := r.Repo.GetViolation(Data.Vnotexists.Id)
	assertEquals(t, "Unexpected error. Expected: %v; Actual: %v", model.ErrNotFound, err)
}

// TestCreateTemplate executes this test
func (r *TestContext) TestCreateTemplate(t *testing.T) {
	var tpl *model.Template
	var err error

	tpl, err = r.Repo.CreateTemplate(&Data.T01)
	assertEquals(t, "Unexpected error. Expected: %v; Actual: %v", nil, err)
	Data.T01 = *tpl

	tpl, err = r.Repo.GetTemplate(Data.T01.Id)
	assertEquals(t, "Unexpected error. Expected: %v; Actual: %v", nil, err)
	assertEquals(t, "Unexpected violation. Expected: %v; Actual: %v", Data.T01.Id, tpl.Id)
}

// TestCreateTemplateExists executes this test
func (r *TestContext) TestCreateTemplateExists(t *testing.T) {
	_, err := r.Repo.CreateTemplate(&Data.T01)
	assertEquals(t, "Expected error: %v; actual: %v", model.ErrAlreadyExist, err)
}

// TestGetAllTemplates executes this test
func (r *TestContext) TestGetAllTemplates(t *testing.T) {
	actual, err := r.Repo.GetAllTemplates()
	assertEquals(t, "Unexpected error. Expected: %v; Actual: %v", nil, err)
	assertEquals(t, "Unexpected len(Templates). Expected: %d; Actual: %d", 1, len(actual))
}

// TestGetTemplate executes this test
func (r *TestContext) TestGetTemplate(t *testing.T) {
	result, err := r.Repo.GetTemplate(Data.T01.Id)
	assertEquals(t, "Unexpected error. Expected: %v; Actual: %v", nil, err)
	assertEquals(t, "Unexpected result. Expected: %v; Actual: %v", Data.T01.Id, result.Id)
}

// TestGetTemplateNotExists executes this test
func (r *TestContext) TestGetTemplateNotExists(t *testing.T) {
	_, err := r.Repo.GetAgreement("notexists")
	assertEquals(t, "Expected error: %v; actual: %v", model.ErrNotFound, err)
}

/*
 * The functions below are kept to maintain backwards compatibility, but should
 * be removed at some point
 */

// TestCreateProvider executes this test
// Deprecated: use TestContext function
func TestCreateProvider(t *testing.T, repo model.IRepository) {
	ctx := TestContext{repo}
	ctx.TestCreateProvider(t)
}

// TestCreateProviderExists executes this test
// Deprecated: use TestContext function
func TestCreateProviderExists(t *testing.T, repo model.IRepository) {
	ctx := TestContext{repo}
	ctx.TestCreateProvider(t)
}

// TestGetAllProviders executes this test
// Deprecated: use TestContext function
func TestGetAllProviders(t *testing.T, repo model.IRepository) {
	ctx := TestContext{repo}
	ctx.TestGetAllProviders(t)
}

// TestGetProvider executes this test
// Deprecated: use TestContext function
func TestGetProvider(t *testing.T, repo model.IRepository) {
	ctx := TestContext{repo}
	ctx.TestGetProvider(t)
}

// TestGetProviderNotExists executes this test
// Deprecated: use TestContext function
func TestGetProviderNotExists(t *testing.T, repo model.IRepository) {
	ctx := TestContext{repo}
	ctx.TestGetProviderNotExists(t)
}

// TestDeleteProvider executes this test
// Deprecated: use TestContext function
func TestDeleteProvider(t *testing.T, repo model.IRepository) {
	ctx := TestContext{repo}
	ctx.TestDeleteProvider(t)
}

// TestDeleteProviderNotExists executes this test
// Deprecated: use TestContext function
func TestDeleteProviderNotExists(t *testing.T, repo model.IRepository) {
	ctx := TestContext{repo}
	ctx.TestDeleteProviderNotExists(t)
}

// TestCreateAgreement executes this test
// Deprecated: use TestContext function
func TestCreateAgreement(t *testing.T, repo model.IRepository) {
	ctx := TestContext{repo}
	ctx.TestCreateAgreement(t)
}

// TestCreateAgreementExists executes this test
// Deprecated: use TestContext function
func TestCreateAgreementExists(t *testing.T, repo model.IRepository) {
	ctx := TestContext{repo}
	ctx.TestCreateAgreementExists(t)
}

// TestGetAllAgreements executes this test
// Deprecated: use TestContext function
func TestGetAllAgreements(t *testing.T, repo model.IRepository) {
	ctx := TestContext{repo}
	ctx.TestGetAllAgreements(t)
}

// TestGetAgreement executes this test
// Deprecated: use TestContext function
func TestGetAgreement(t *testing.T, repo model.IRepository) {
	ctx := TestContext{repo}
	ctx.TestGetAgreement(t)
}

// TestGetAgreementNotExists executes this test
// Deprecated: use TestContext function
func TestGetAgreementNotExists(t *testing.T, repo model.IRepository) {
	ctx := TestContext{repo}
	ctx.TestGetAgreementNotExists(t)
}

// TestGetAgreementsByState executes this test
// Deprecated: use TestContext function
func TestGetAgreementsByState(t *testing.T, repo model.IRepository) {
	ctx := TestContext{repo}
	ctx.TestGetAgreementsByState(t)
}

// TestUpdateAgreementState executes this test
// Deprecated: use TestContext function
func TestUpdateAgreementState(t *testing.T, repo model.IRepository) {
	ctx := TestContext{repo}
	ctx.TestUpdateAgreementState(t)
}

// TestUpdateAgreementStateNotExists executes this test
// Deprecated: use TestContext function
func TestUpdateAgreementStateNotExists(t *testing.T, repo model.IRepository) {
	ctx := TestContext{repo}
	ctx.TestUpdateAgreementStateNotExists(t)
}

// TestUpdateAgreement executes this test
// Deprecated: use TestContext function
func TestUpdateAgreement(t *testing.T, repo model.IRepository) {
	ctx := TestContext{repo}
	ctx.TestUpdateAgreement(t)
}

// TestUpdateAgreementNotExists executes this test
// Deprecated: use TestContext function
func TestUpdateAgreementNotExists(t *testing.T, repo model.IRepository) {
	ctx := TestContext{repo}
	ctx.TestUpdateAgreementNotExists(t)
}

// TestDeleteAgreement executes this test
// Deprecated: use TestContext function
func TestDeleteAgreement(t *testing.T, repo model.IRepository) {
	ctx := TestContext{repo}
	ctx.TestDeleteAgreement(t)
}

// TestDeleteAgreementNotExists executes this test
// Deprecated: use TestContext function
func TestDeleteAgreementNotExists(t *testing.T, repo model.IRepository) {
	ctx := TestContext{repo}
	ctx.TestDeleteAgreementNotExists(t)
}

// TestCreateViolation executes this test
// Deprecated: use TestContext function
func TestCreateViolation(t *testing.T, repo model.IRepository) {
	ctx := TestContext{repo}
	ctx.TestCreateViolation(t)
}

// TestCreateViolationExists executes this test
// Deprecated: use TestContext function
func TestCreateViolationExists(t *testing.T, repo model.IRepository) {
	ctx := TestContext{repo}
	ctx.TestCreateViolationExists(t)
}

// TestGetViolation executes this test
// Deprecated: use TestContext function
func TestGetViolation(t *testing.T, repo model.IRepository) {
	ctx := TestContext{repo}
	ctx.TestGetViolation(t)
}

// TestGetViolationNotExists executes this test
// Deprecated: use TestContext function
func TestGetViolationNotExists(t *testing.T, repo model.IRepository) {
	ctx := TestContext{repo}
	ctx.TestGetViolationNotExists(t)
}

func assertEquals(t *testing.T, msg string, expected interface{}, actual interface{}) {
	if expected != actual {
		buf := bytes.Buffer{}
		buf.Write(debug.Stack())
		t.Errorf(msg+"\n%s", expected, actual, buf.String())

	}
}
