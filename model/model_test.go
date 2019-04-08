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
package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

var pr = Provider{Id: "id", Name: "name"}
var cl = Client{Id: "id", Name: "name"}
var val = NewDefaultValidator(false, true)

func TestProviders(t *testing.T) {
	p := Provider{Id: "id", Name: "name"}
	checkNumber(t, &p, 0)

	if p.GetId() != p.Id {
		t.Errorf("Provider.Id and Provider.GetId() do not match")
	}

	p = Provider{Id: "", Name: "name"}
	checkNumber(t, &p, 1)

	p = Provider{Id: "id", Name: ""}
	checkNumber(t, &p, 1)

	p = Provider{Id: "", Name: ""}
	checkNumber(t, &p, 2)
}

func TestClient(t *testing.T) {
	checkNumber(t, &cl, 0)
	if cl.GetId() != cl.Id {
		t.Errorf("Provider.Id and Provider.GetId() do not match")
	}
}

func TestAssessment(t *testing.T) {
	a := Assessment{FirstExecution: time.Now(), LastExecution: time.Now()}
	checkNumber(t, &a, 0)
}

func TestGuarantee(t *testing.T) {
	g := Guarantee{Name: "name", Constraint: "a LT 10"}
	checkNumber(t, &g, 0)

	g = Guarantee{Name: "", Constraint: "a LT 10"}
	checkNumber(t, &g, 1)

	g = Guarantee{Name: "name", Constraint: ""}
	checkNumber(t, &g, 1)

}

func TestDetails(t *testing.T) {
	at := Details{Id: "id", Name: "name", Provider: pr, Client: cl}
	checkNumber(t, &at, 0)

	at = Details{Id: "", Name: "name", Provider: pr, Client: cl}
	checkNumber(t, &at, 1)

	at = Details{Id: "id", Name: "", Provider: pr, Client: cl}
	checkNumber(t, &at, 1)

	at = Details{Id: "id", Name: "name", Client: cl}
	checkNumber(t, &at, 2)

	at = Details{Id: "id", Name: "name", Provider: pr}
	checkNumber(t, &at, 2)

	at = Details{
		Id:       "id",
		Name:     "name",
		Provider: pr,
		Client:   cl,
		Guarantees: []Guarantee{
			Guarantee{Name: ""},
		},
	}
	checkNumber(t, &at, 2)
}

func TestAgreement(t *testing.T) {

	a := Agreement{
		Id:         "id",
		Name:       "name",
		State:      STOPPED,
		Assessment: Assessment{},
		Details: Details{
			Id:       "id",
			Name:     "name",
			Provider: pr,
			Client:   cl,
		},
	}
	checkNumber(t, &a, 0)

	if a.GetId() != a.Id {
		t.Errorf("Agreement.Id and Agreement.GetId() do not match")
	}

	a.Id = ""
	a.Details.Id = ""
	a.Name = "name"
	a.Details.Name = "name"
	checkNumber(t, &a, 2) // one error per empty id

	a.Id = "id"
	a.Details.Id = "id"
	a.Name = ""
	a.Details.Name = ""
	checkNumber(t, &a, 2) // one error per empty name

	a.Id = "id1"
	a.Details.Id = "id2"
	a.Name = "name"
	a.Details.Name = "name"
	checkNumber(t, &a, 1)

	a.Id = "id"
	a.Details.Id = "id"
	a.Name = "name1"
	a.Details.Name = "name2"
	checkNumber(t, &a, 1)
}

func TestTemplate(test *testing.T) {
	t, err := ReadTemplate("testdata/template.json")
	if err != nil {
		test.Error(err)
	}
	checkNumber(test, &t, 0)

	t.Id = "does-not-match-details-id"
	checkNumber(test, &t, 1)

	t.Id = ""
	t.Details.Id = ""
	checkNumber(test, &t, 2) // one error per empty id

	t.Id = "id"
	t.Details.Id = "id"
	t.Name = ""
	checkNumber(test, &t, 1)

	t.Name = "name"
	t.Details.Type = AGREEMENT
	checkNumber(test, &t, 1)
}

func TestStates(t *testing.T) {
	a := Agreement{State: STOPPED}
	if !a.IsStopped() {
		t.Error("Agreement should be stopped")
	}

	a = Agreement{State: STARTED}
	if !a.IsStarted() {
		t.Error("Agreement should be started")
	}
	a = Agreement{State: TERMINATED}
	if !a.IsTerminated() {
		t.Error("Agreement should be terminated")
	}
}

func TestIsValidTransition(t *testing.T) {

	type transition struct {
		old State
		new State
	}

	a := Agreement{}
	transitions := map[transition]bool{
		{STOPPED, STOPPED}:       true,
		{STOPPED, STARTED}:       true,
		{STOPPED, TERMINATED}:    true,
		{STARTED, STOPPED}:       true,
		{STARTED, STARTED}:       true,
		{STARTED, TERMINATED}:    true,
		{TERMINATED, STOPPED}:    false,
		{TERMINATED, STARTED}:    false,
		{TERMINATED, TERMINATED}: false,
	}

	for transition, valid := range transitions {
		a.State = transition.old
		if a.IsValidTransition(transition.new) != valid {
			t.Errorf("IsValidTransition from %s to %s. Expected: %v. Actual: %v",
				a.State, transition.new, valid, !valid)
		}
	}
}

func TestProviderSerialization(t *testing.T) {
	var p Provider

	s := `{"id": "id", "name": "name" }`
	json.NewDecoder(strings.NewReader(s)).Decode(&p)
	checkNumber(t, &p, 0)
}

func TestAgreementSerialization(t *testing.T) {
	var a1, a2 Agreement
	var err error

	s := `{
		"id": "id", 
		"name": "name", 
		"details": {
			"id": "id",
			"name": "name",
			"provider": { "id": "pr-id", "name": "pr-name" },
			"client": { "id": "cl-id", "name": "cl-name" }
		},
		"state": "stopped"
	}`
	err = json.NewDecoder(strings.NewReader(s)).Decode(&a1)
	if err != nil {
		t.Fatalf("Error decoding %v", err)
	}
	checkNumber(t, &a1, 0)

	// state is empty. Validate should normalize to STOPPED
	s = `{
		"id": "id", 
		"name": "name", 
		"details": {
			"id": "id",
			"name": "name",
			"provider": { "id": "pr-id", "name": "pr-name" },
			"client": { "id": "cl-id", "name": "cl-name" }
		}
	}`
	err = json.NewDecoder(strings.NewReader(s)).Decode(&a2)
	if err != nil {
		t.Fatalf("Error decoding %v", err)
	}
	checkNumber(t, &a2, 0)
	if a2.State != STOPPED {
		t.Errorf("State=%s is not STOPPED", a2.State)
	}

}

func TestViolation(t *testing.T) {
	var v = Violation{}
	checkNumber(t, &v, 6)
	if v.GetId() != v.Id {
		t.Errorf("Violation.Id and Violation.GetId() do not match")
	}
}

func TestViolationSerialization(t *testing.T) {
	var v Violation
	s := `{
		"id": "v-id",
		"agreement_id": "a-id",
		"datetime": "2018-05-15T14:15:00Z",
		"guarantee": "gt-name",
		"constraint": "var1 < 100 and var2 > 100",
		"values": [{ "key": "var1", "value": 101, "datetime": "2018-05-15T14:15:01Z"}, { "key": "var2", "value": 100, "datetime": "2018-05-15T14:15:02Z"}]
	}`
	err := json.NewDecoder(strings.NewReader(s)).Decode(&v)
	if err != nil {
		t.Fatalf("Error decoding %v", err)
	}
	checkNumber(t, &v, 0)
}

type valError string

func (e valError) Error() string {
	return string(e)
}
func (e valError) IsErrValidation() bool {
	return true
}

func TestValidationError(t *testing.T) {
	var e error
	if e = valError("error"); !IsErrValidation(e) {
		t.Errorf("Error %v should be a validation error", e)
	}
	if e = errors.New("error"); IsErrValidation(e) {
		t.Errorf("Error %v should not be a validation error", e)
	}
}

func TestMetricValue(t *testing.T) {
	m := MetricValue{
		DateTime: time.Now(),
		Key:      "m1",
		Value:    100,
	}
	if s := fmt.Sprintf("%s", m); s == "" {
		t.Errorf("MetricValue.String() should not be an empty string")
	}
}

func checkNumber(t *testing.T, v Validable, expected int) {
	if errs := v.Validate(val, CREATE); len(errs) != expected {
		t.Errorf("Error validating %s%v. Errors = %v; Expected: %d", reflect.TypeOf(v), v, errs, expected)
	}
}
