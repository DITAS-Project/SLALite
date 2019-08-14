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
	"SLALite/assessment/monitor/simpleadapter"
	"SLALite/model"
	"flag"
	"os"
	"testing"
	"time"

	blueprint "github.com/DITAS-Project/blueprint-go"
)

const (
	DS4MUrl = "http://31.171.247.162:50003/NotifyViolation"
)

var (
	integrationNotifier = flag.Bool("notifier", false, "run DS4M integration tests")
	integrationElastic  = flag.Bool("elastic", false, "run ElasticSearch integration tests")
)

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

	if len(slas) > 1 {
		t.Fatalf("Expected one SLA but found %d", len(slas))
	}

	sla := slas[0]

	_, ok := methodsInfo[sla.Id]
	if !ok {
		t.Fatalf("Can't find method information for SLA %s", sla.Id)
	}

	if len(sla.Details.Guarantees) != 4 {
		t.Fatalf("Wrong number of guarantees for SLA %s: expected 4 but found %d", sla.Id, len(sla.Details.Guarantees))
	}

	for _, guarantee := range sla.Details.Guarantees {
		var expected string
		switch guarantee.Name {
		case "serviceAvailable":
			expected = "availability >= 90.000000"
		case "fastProcess":
			expected = "responseTime <= 2.000000"
		case "freshData":
			expected = "timeliness <= 99.000000"
		case "EnoughData":
			expected = "volume >= 1200.000000"
		}
		if guarantee.Constraint != expected {
			t.Fatalf("Invalid guarantee %s found in SLA %s: Expected %s but found %s", guarantee.Name, sla.Id, expected, guarantee.Constraint)
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

	adapter := simpleadapter.New(assessment_model.GuaranteeData{
		m1,
	})

	result := assessment.AssessAgreement(&slas[0], adapter, time.Now())
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
