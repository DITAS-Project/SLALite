/**
 * Copyright 2018 Atos
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not
 * use this file except in compliance with the License. You may obtain a copy of
 * the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
 * WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
 * License for the specific language governing permissions and limitations under
 * the License.
 *
 * This is being developed for the DITAS Project: https://www.ditas-project.eu/
 */

package ditas

import (
	"SLALite/assessment"
	assessment_model "SLALite/assessment/model"
	"SLALite/model"
	"flag"
	"os"
	"testing"
	"time"

	"github.com/Knetic/govaluate"

	blueprint "github.com/DITAS-Project/blueprint-go"
)

const (
	DS4MUrl = "http://31.171.247.162:50003/NotifyViolation"
)

var (
	integrationNotifier = flag.Bool("notifier", false, "run DS4M integration tests")
	integrationElastic  = flag.Bool("elastic", false, "run ElasticSearch integration tests")
)

type TestMonitoring struct {
	Metrics assessment_model.ExpressionData
}

func (t *TestMonitoring) Initialize(a *model.Agreement) {
}

func (t *TestMonitoring) GetValues(gt model.Guarantee, vars []string) assessment_model.GuaranteeData {
	return assessment_model.GuaranteeData{t.Metrics}
}

var t0 = time.Now()
var testNotifier = DitasNotifier{
	VDCId: "VDC_2",
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func TestReader(t *testing.T) {
	bp, err := blueprint.ReadBlueprint("resources/concrete_blueprint_doctor.json")

	if err != nil {
		t.Fatalf("Error reading blueprint: %s", err.Error())
	}

	slas, methodsInfo := CreateAgreements(bp)

	if slas == nil || len(slas) == 0 {
		t.Fatalf("Did not get any SLA from the blueprint")
	}

	for _, sla := range slas {
		_, ok := methodsInfo[sla.Id]
		if !ok {
			t.Fatalf("Can't find method information for SLA %s", sla.Id)
		}

		if !(len(sla.Details.Guarantees) > 0) {
			t.Fatalf("Guarantees were not generated for SLA %s", sla.Id)
		}

		for _, guarantee := range sla.Details.Guarantees {
			if guarantee.Name == "" {
				t.Fatalf("Empty guarantee name for SLA %s", sla.Id)
			}

			if guarantee.Constraint == "" {
				t.Fatalf("Empty constraint for guarantee %s of SLA %s", guarantee.Name, sla.Id)
			}
		}
	}
}

func TestNotifier(t *testing.T) {
	if *integrationNotifier {
		testNotifier.NotifyUrl = DS4MUrl
	}
	bp, err := blueprint.ReadBlueprint("resources/concrete_blueprint_doctor.json")
	if err != nil {
		t.Fatalf("Error reading blueprint: %s", err.Error())
	}

	testNotifier.VDCId = *bp.InternalStructure.Overview.Name

	slas, _ := CreateAgreements(bp)
	slas[0].State = model.STARTED

	var m1 = assessment_model.ExpressionData{
		"availability": model.MetricValue{Key: "availability", Value: 90, DateTime: t_(0)},
		"responseTime": model.MetricValue{Key: "responseTime", Value: 1.5, DateTime: t_(0)},
		"timeliness":   model.MetricValue{Key: "timeliness", Value: 0.5, DateTime: t_(0)},
	}

	adapter := TestMonitoring{
		Metrics: m1,
	}

	result := assessment.AssessAgreement(&slas[0], &adapter, time.Now())
	testNotifier.NotifyViolations(&slas[0], &result)

	notViolations := testNotifier.Violations
	if len(notViolations) == 1 {
		violation := notViolations[0]
		if violation.VDCId != testNotifier.VDCId {
			t.Errorf("Unexpected VDCId: %s. Expected %s", violation.VDCId, testNotifier.VDCId)
		}

		if violation.Method != slas[0].Id {
			t.Errorf("Unexpected method name: %s. Expected %s", violation.Method, slas[0].Id)
		}

		if len(violation.Metrics) != 2 {
			t.Errorf("Unexpected number of metrics: %d. Expected %d", len(violation.Metrics), 3)
		}

		expectedMetrics := make(map[string]bool)
		expectedMetrics["availability"] = false
		expectedMetrics["responseTime"] = false

		for _, metric := range violation.Metrics {
			found, ok := expectedMetrics[metric.Key]
			if !ok {
				t.Errorf("Unexpected metric found: %s.", metric)
			}

			if found {
				t.Errorf("Unexpected duplicate metric found: %s.", metric)
			}

			expectedMetrics[metric.Key] = true
		}

		for metric, found := range expectedMetrics {
			if !found {
				t.Errorf("Expected metric %s not found in results", metric)
			}
		}

	} else {
		t.Errorf("Unexpected number of violations: %d. Expected %d", len(notViolations), 1)
	}

}

func TestElastic(t *testing.T) {
	if *integrationElastic {
		t.Log("Testing elasticsearch integration")
		bp, err := blueprint.ReadBlueprint("resources/concrete_blueprint_doctor.json")

		if err != nil {
			t.Fatalf("Error reading blueprint: %s", err.Error())
		}

		slas, methods := CreateAgreements(bp)

		sla := slas[0]

		monitor := NewAdapter("http://localhost:9200", methods)

		monitor.Initialize(&sla)

		for _, guarantee := range sla.Details.Guarantees {
			constraint := guarantee.Constraint
			exp, err := govaluate.NewEvaluableExpression(constraint)
			if err != nil {
				t.Fatalf("Invalid constraint %s found: %s", constraint, err.Error())
			}
			vars := exp.Vars()
			values := monitor.GetValues(guarantee, vars)

			if len(values) == 0 {
				t.Errorf("Can't find values for constraint %s", constraint)
			}

			for _, metrics := range values {
				if len(metrics) == 0 {
					t.Errorf("Found empty metrics map for constraint %s", constraint)
				}
				for key, value := range metrics {
					if !contains(key, vars) {
						t.Fatalf("Found metric not requested %s", key)
					}
					if value.Key != key {
						t.Fatalf("Found not matching key in map. Expected: %s, found: %s", key, value.Key)
					}
				}
			}
		}

	} else {
		t.Log("Skipping elasticsearch integration test")
	}
}

func contains(key string, values []string) bool {
	for _, value := range values {
		if value == key {
			return true
		}
	}
	return false
}

func t_(second time.Duration) time.Time {
	return t0.Add(time.Second * second)
}
