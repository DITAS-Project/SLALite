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
package ditas

import (
	"SLALite/assessment"
	"SLALite/assessment/monitor"
	"SLALite/model"
	"os"
	"testing"
	"time"
)

type TestMonitoring struct {
	Metrics  map[string]map[string]monitor.MetricValue
	consumed map[string]bool
}

func (t *TestMonitoring) Initialize(a *model.Agreement) {
	t.consumed = make(map[string]bool)
}

func (t *TestMonitoring) NextValues(gt model.Guarantee) map[string]monitor.MetricValue {
	_, ok := t.consumed[gt.Name]
	if !ok {
		t.consumed[gt.Name] = true
		return t.Metrics[gt.Name]
	}

	return nil
}

var t0 = time.Now()
var notifier DitasNotifier

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func checkMethodList(t *testing.T, methodList *MethodListType, name string) {
	if methodList == nil {
		t.Fatalf("Can't find method list %s", name)
	}

	if len(methodList.Methods) == 0 {
		t.Fatalf("No methods found in %s section", name)
	}

	if *methodList.Methods[0].Name != "patient-details" {
		t.Fatalf("Unexpected name for method in %s. Found %s", name, *methodList.Methods[0].Name)
	}

}

func TestReader(t *testing.T) {
	blueprint := ReadBlueprint("resources/vdc_blueprint_example_1.json")

	checkMethodList(t, blueprint.DataManagement, "Data Management")
	checkMethodList(t, blueprint.AbstractProperties, "Abstract Properties")

	slas := CreateAgreements(blueprint)

	if slas == nil || len(slas) == 0 {
		t.Fatalf("Did not get any SLA from the blueprint")
	}

	sla := slas[0]
	if sla.Id != "patient-details" {
		t.Fatalf("Did not find patient-details SLA")
	}

	guarantees := sla.Details.Guarantees
	if len(guarantees) != 3 {
		t.Fatalf("Unexpected number of guarantees. Expected 3 but found %d", len(guarantees))
	}

}

func TestNotifier(t *testing.T) {
	blueprint := ReadBlueprint("resources/vdc_blueprint_example_1.json")
	slas := CreateAgreements(blueprint)
	slas[0].State = model.STARTED
	slas[0].Details.Expiration = time.Now().Add(24 * time.Hour)

	var m1 = map[string]map[string]monitor.MetricValue{
		"1 or 4": {
			"Availability": monitor.MetricValue{Key: "Availability", Value: 90, DateTime: t_(0)},
			"Timeliness":   monitor.MetricValue{Key: "Timeliness", Value: 1, DateTime: t_(0)},
		},
		"2": {
			"ResponseTime": monitor.MetricValue{Key: "ResponseTime", Value: 1.5, DateTime: t_(0)},
		},
		"3 or 5": {
			"volume":               monitor.MetricValue{Key: "volume", Value: 10000, DateTime: t_(0)},
			"Process_completeness": monitor.MetricValue{Key: "Process_completeness", Value: 95, DateTime: t_(0)},
		},
	}

	adapter := TestMonitoring{
		Metrics: m1,
	}

	result := assessment.AssessAgreement(&slas[0], &adapter, time.Now())
	notifier.NotifyViolations(&slas[0], &result)

	if len(notifier.Result.GetViolations()) != 2 {
		t.Fatalf("Violation number don't match. Expected %d, found %d", 2, len(notifier.Result.GetViolations()))
	}

	for _, violation := range notifier.Result.GetViolations() {
		var expected []string
		switch violation.Guarantee {
		case "1 or 4":
			expected = []string{"Availability", "Timeliness"}
		case "2":
			expected = []string{"ResponseTime"}
		}
		checkValues(t, violation.Guarantee, violation.Values, expected)
	}
}

func checkValues(t *testing.T, gt string, values map[string]interface{}, expected []string) {

	if len(values) != len(expected) {
		t.Fatalf("Different number of values found for violation of guarantee %s. Expected %d, found %d", gt, len(values), len(expected))
	}

	for _, value := range expected {
		_, ok := values[value]
		if !ok {
			t.Errorf("Expected %s metric not found in violation", value)
		}
	}
}

func t_(second time.Duration) time.Time {
	return t0.Add(time.Second * second)
}
