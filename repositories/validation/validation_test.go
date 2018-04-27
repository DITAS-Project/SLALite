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

package validation

import (
	"SLALite/model"
	"SLALite/repositories/memrepository"
	"encoding/json"
	"os"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func TestCallThroughMethods(t *testing.T) {
	r, _ := memrepository.New(nil)
	v, _ := New(r, true)

	v.GetProvider("id")
	v.GetAllProviders()
	v.GetAgreement("id")
	v.GetAllAgreements()
	v.GetActiveAgreements()
	v.GetViolation("id")
	v.StartAgreement("id")
	v.StopAgreement("id")
}

func TestRepositoryWithExternalIds(t *testing.T) {
	var err error
	a, _ := readAgreement("testdata/a.json")

	as := map[string]model.Agreement{
		"id": *a,
	}

	r := memrepository.NewMemRepository(nil, as, nil, nil)
	v, _ := New(r, true)

	p := &model.Provider{Id: "", Name: "Name"}
	p, err = v.CreateProvider(p)
	if err != nil {
		t.Errorf("No errors expected. Found %v", err)
		return
	}
	v.DeleteProvider(p)

	p.Id = "id"
	p, err = v.CreateProvider(p)
	if err == nil {
		t.Errorf("Errors expected. Found %v", err)
		return
	}

	a.Id = ""
	a, err = v.CreateAgreement(a)
	if err != nil {
		t.Errorf("No errors expected. Found %v", err)
		return
	}
	_, err = v.UpdateAgreement(a)
	if err == nil {
		t.Errorf("Errors expected. Found %v", err)
		return
	}
	v.DeleteAgreement(a)

	a.Id = "id"
	a.Details.Id = "id2"
	a, err = v.CreateAgreement(a)
	if err == nil {
		t.Errorf("Errors expected. Found %v", err)
		return
	}

	a, err = v.UpdateAgreement(a)
	if err != nil {
		t.Errorf("No errors expected. Found %v", err)
		return
	}

	vi := &model.Violation{Id: "", AgreementId: "id", Guarantee: "gt", Datetime: time.Now()}
	vi, err = v.CreateViolation(vi)
	if err != nil {
		t.Errorf("No errors expected. Found %v", err)
		return
	}

	vi.Id = "id"
	vi, err = v.CreateViolation(vi)
	if err == nil {
		t.Errorf("Errors expected. Found %v", err)
		return
	}

}

func TestRepositoryWithoutExternalIds(t *testing.T) {
	var err error
	r, _ := memrepository.New(nil)
	v, _ := New(r, false)

	p := &model.Provider{Id: "Id", Name: "Name"}
	p, err = v.CreateProvider(p)
	if err != nil {
		t.Errorf("No errors expected. Found %v", err)
		return
	}
	v.DeleteProvider(p)

	p.Id = ""
	p, err = v.CreateProvider(p)
	if err == nil {
		t.Errorf("Errors expected. Found %v", err)
		return
	}

	a, _ := readAgreement("testdata/a.json")
	a, err = v.CreateAgreement(a)
	if err != nil {
		t.Errorf("No errors expected. Found %v", err)
		return
	}
	v.UpdateAgreement(a)
	v.DeleteAgreement(a)

	a.Id = ""
	a, err = v.CreateAgreement(a)
	if err == nil {
		t.Errorf("Errors expected. Found %v", err)
		return
	}

	vi := &model.Violation{Id: "", AgreementId: "id", Guarantee: "gt", Datetime: time.Now()}
	vi, err = v.CreateViolation(vi)
	if err == nil {
		t.Errorf("No errors expected. Found %v", err)
		return
	}

	vi.Id = "id"
	vi, err = v.CreateViolation(vi)
	if err != nil {
		t.Errorf("Errors expected. Found %v", err)
		return
	}

}

func readAgreement(path string) (*model.Agreement, error) {
	var a model.Agreement

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	json.NewDecoder(f).Decode(&a)
	f.Close()
	return &a, nil
}
