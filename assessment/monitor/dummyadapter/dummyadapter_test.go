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

package dummyadapter

import (
	"SLALite/model"
	"os"
	"testing"
	"time"

	"github.com/Knetic/govaluate"
)

var a1 = createAgreement("a01", p1, c2, "Agreement 01", "m >= 0.5 && n >= 0.5", nil)
var p1 = model.Provider{Id: "p01", Name: "Provider01"}
var c2 = model.Client{Id: "c02", Name: "A client"}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func TestDummyAdapter(t *testing.T) {
	ma := New(1)
	ma.Initialize(&a1)

	gt := a1.Details.Guarantees[0]
	exp, err := govaluate.NewEvaluableExpression(gt.Constraint)
	if err != nil {
		t.Fatalf("Invalid expression %s", gt.Constraint)
	}
	values := ma.GetValues(gt, exp.Vars(), time.Now())
	if values == nil {
		t.Fatalf("GetValues(). Expected: []map[string]monitor.MetricValue. Actual: nil")
	}
	if len(values) != 1 {
		t.Fatalf("len(GetValues()). Expected: 1. Actual: %v", len(values))
	}
	value := values[0]
	if _, ok := value["m"]; !ok {
		t.Fatalf("GetValues()[0]['m'] does not exist")
	}
	if _, ok := value["n"]; !ok {
		t.Errorf("GetValues()[0]['n'] does not exist")
	}
}

func createAgreement(aid string, provider model.Provider, client model.Client, name string, constraint string, expiration *time.Time) model.Agreement {
	return model.Agreement{
		Id:    aid,
		Name:  name,
		State: model.STOPPED,
		Details: model.Details{
			Id:       aid,
			Name:     name,
			Type:     model.AGREEMENT,
			Provider: provider, Client: client,
			Creation:   time.Now(),
			Expiration: expiration,
			Guarantees: []model.Guarantee{
				model.Guarantee{Name: "TestGuarantee", Constraint: constraint},
			},
		},
	}
}
