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
)

var a1 = createAgreement("a01", p1, c2, "Agreement 01", "m >= 0.5 && n >= 0.5")
var p1 = model.Provider{Id: "p01", Name: "Provider01"}
var c2 = model.Client{Id: "c02", Name: "A client"}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func TestDummyAdapter(t *testing.T) {
	ma := New()
	ma.Initialize(&a1)

	gt := a1.Details.Guarantees[0]
	values := ma.NextValues(gt)
	if values == nil {
		t.Errorf("NextValues(). Expected: map[string]monitor.MetricValue. Actual: nil")
		return
	}
	if len(values) != 2 {
		t.Errorf("len(NextValues()). Expected: 2. Actual: %v", len(values))
		return
	}
	if _, ok := values["m"]; !ok {
		t.Errorf("NextValues()['m'] does not exist")
		return
	}
	if _, ok := values["n"]; !ok {
		t.Errorf("NextValues()['n'] does not exist")
	}
	values = ma.NextValues(gt)
	if values != nil {
		t.Errorf("NextValues(). Expected: nil. Actual: %v", values)
	}
}

func createAgreement(aid string, provider model.Provider, client model.Client, name string, constraint string) model.Agreement {
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
			Expiration: time.Now().Add(24 * time.Hour),
			Guarantees: []model.Guarantee{
				model.Guarantee{Name: "TestGuarantee", Constraint: constraint},
			},
		},
	}
}
