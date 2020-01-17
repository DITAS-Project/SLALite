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
	"SLALite/assessment/monitor"
	"SLALite/assessment/monitor/genericadapter"
	"SLALite/assessment/monitor/simpleadapter"
	"SLALite/model"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"testing"
	"time"

	blueprint "github.com/DITAS-Project/blueprint-go"
	"github.com/jarcoal/httpmock"
)

const (
	DS4MUrl          = "http://ds4m"
	dataAnalyticsURL = "http://data-analytics/data-analytics"
)

var t0 = time.Now()

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

func getRamdomTestData(sla model.Agreement, ranges map[string]struct {
	Min float64
	Max float64
}, from time.Time) (map[string][]model.MetricValue, time.Time, error) {
	result := make(map[string][]model.MetricValue)
	currentTime := from
	for _, variable := range sla.Details.Variables {
		values := make([]model.MetricValue, 10)
		currentRange, ok := ranges[variable.Metric]
		if !ok {
			return result, currentTime, fmt.Errorf("Can't find range for variable %s", variable.Metric)
		}
		if currentRange.Min > currentRange.Max {
			return result, currentTime, fmt.Errorf("Invalid range for variable %s: Min %f is greater than max %f", variable.Metric, currentRange.Min, currentRange.Max)
		}
		max := currentRange.Max - currentRange.Min
		for i := range values {
			currentTime := currentTime.Add(time.Minute * time.Duration(1))
			values[i] = model.MetricValue{
				Key:      variable.Metric,
				Value:    currentRange.Min + (rand.Float64() * max),
				DateTime: currentTime,
			}
		}
		result[variable.Metric] = values
	}
	return result, currentTime, nil
}

func getMonitoringItems(sla model.Agreement, from time.Time, to time.Time) []monitor.RetrievalItem {
	vars := sla.Details.Variables
	result := make([]monitor.RetrievalItem, len(vars))
	for i, variable := range vars {
		result[i] = monitor.RetrievalItem{
			From: from,
			To:   to,
			Var:  variable,
		}
	}
	return result
}

func mockDataRetrieval(data map[string][]model.MetricValue) {
	httpmock.RegisterResponder("GET", dataAnalyticsURL+"/infra1", func(req *http.Request) (*http.Response, error) {
		query := req.URL.Query()
		metric := query.Get("name")
		if metric == "" {
			return httpmock.NewStringResponse(http.StatusBadRequest, "Can't find metric name in request"), nil
		}
		operation := query.Get("operationID")
		if operation == "" {
			return httpmock.NewStringResponse(http.StatusBadRequest, "Can't find operation identifier in request"), nil
		}
		metricData, ok := data[metric]
		if !ok {
			return httpmock.NewStringResponse(http.StatusNotFound, fmt.Sprintf("Can't find data for metric %s", metric)), nil
		}
		result := make([]DataAnalyticsMeter, len(metricData))
		for i := range result {
			value := metricData[i]
			result[i] = DataAnalyticsMeter{
				OperationID: operation,
				Name:        metric,
				Value:       value.Value.(float64),
				Unit:        "",
				Timestamp:   value.DateTime.Format(time.RFC3339),
			}
		}
		return httpmock.NewJsonResponse(http.StatusOK, result)
	})
}

func TestDitasMonitoringAdapter(t *testing.T) {
	vdcID := "vdc1"
	infraID := "infra1"
	bp, err := blueprint.ReadBlueprint("resources/concrete_blueprint_doctor.json")
	if err != nil {
		t.Fatalf("Error reading blueprint: %s", err.Error())
	}

	slas, _ := CreateAgreements(bp)
	sla := slas[0]
	sla.State = model.STARTED

	da := NewDataAnalyticsAdapter(dataAnalyticsURL, vdcID, infraID, TestingConfiguration{
		Enabled: false,
	}, false)

	httpmock.ActivateNonDefault(da.Client.GetClient())
	defer httpmock.DeactivateAndReset()

	testData := map[string]struct {
		Min float64
		Max float64
	}{
		"availability": {0.0, 100.0},
		"responseTime": {0.1, 10},
		"timeliness":   {0.0, 100.0},
		"volume":       {1000.0, 10000.0},
	}
	from := time.Now()
	data, to, err := getRamdomTestData(sla, testData, from)
	if err != nil {
		t.Fatalf("Error getting random data: %s", err.Error())
	}

	mockDataRetrieval(data)

	responseValues := da.Retrieve(sla, getMonitoringItems(sla, from, to))
	if len(data) != len(responseValues) {
		t.Fatalf("Retrieved %d variables but expected %d", len(responseValues), len(data))
	}
	for variable, values := range responseValues {
		varSourceData, ok := data[variable.Metric]
		if !ok {
			t.Fatalf("Can't find original data for returned metric %s", variable.Metric)
		}
		if len(values) != len(varSourceData) {
			t.Fatalf("There are %d returned metrics for variable %s while it was expected to be %d", len(values), variable.Metric, len(varSourceData))
		}
	}

	adapter := genericadapter.New(da.Retrieve, da.Process).Initialize(&sla)
	for _, guarantee := range sla.Details.Guarantees {
		var vars []string
		switch guarantee.Name {
		case "serviceAvailable":
			vars = []string{"availability"}
		case "fastProcess":
			vars = []string{"responseTime"}
		case "freshData":
			vars = []string{"timeliness"}
		case "EnoughData":
			vars = []string{"volume"}
		}
		vals := adapter.GetValues(guarantee, vars, to)
		if len(vals) != 1 {
			t.Fatalf("Expected a single aggregated value for guarantee %s but found %d values instead", guarantee.Name, len(vals))
		}

	}

}

func TestNotifier(t *testing.T) {
	bp, err := blueprint.ReadBlueprint("resources/concrete_blueprint_doctor.json")
	if err != nil {
		t.Fatalf("Error reading blueprint: %s", err.Error())
	}

	testNotifier := NewNotifier("VDC_2", DS4MUrl, TestingConfiguration{
		Enabled: false,
	}, false)

	httpmock.ActivateNonDefault(testNotifier.Client.GetClient())
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("POST", DS4MUrl+DS4MNotifyPath, httpmock.NewStringResponder(http.StatusOK, ""))

	slas, _ := CreateAgreements(bp)
	slas[0].State = model.STARTED

	var m1 = assessment_model.ExpressionData{
		"availability": model.MetricValue{Key: "availability", Value: 90, DateTime: t_(0)},
		"responseTime": model.MetricValue{Key: "responseTime", Value: 1.5, DateTime: t_(0)},
		"timeliness":   model.MetricValue{Key: "timeliness", Value: 102.0, DateTime: t_(0)},
		"volume":       model.MetricValue{Key: "volume", Value: 1100.0, DateTime: t_(0)},
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
		expectedMetrics["volume"] = false
		expectedMetrics["timeliness"] = false

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
