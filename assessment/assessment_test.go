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
package assessment

import (
	"SLALite/model"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/Knetic/govaluate"
)

var a1 = createAgreement("a01", p1, c2, "Agreement 01", "m >= 0")
var p1 = model.Provider{Id: "p01", Name: "Provider01"}
var c2 = model.Client{Id: "c02", Name: "A client"}
var t0 = time.Now()

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func TestAssessAgreement(t *testing.T) {
	a2 := createAgreement("a02", p1, c2, "Agreement 02", "m >= 0")
	values := []map[string]MetricValue{
		{"m": MetricValue{Key: "m", Value: 1, DateTime: t_(0)}},
		{"m": MetricValue{Key: "m", Value: -1, DateTime: t_(1)}},
	}
	ma := NewSimpleMonitoring(values)

	a2.State = model.STOPPED
	result := AssessAgreement(&a2, ma, t0)
	checkAssessmentResult(t, &a2, result, model.STOPPED, 0)

	a2.State = model.TERMINATED
	result = AssessAgreement(&a2, ma, t0)
	checkAssessmentResult(t, &a2, result, model.TERMINATED, 0)

	a2.State = model.STARTED
	result = AssessAgreement(&a2, ma, t0)
	checkAssessmentResult(t, &a2, result, model.STARTED, 1)	
	checkTimes(t, &a2, t0, t0)

	t1 := t_(1)
	result = AssessAgreement(&a2, ma, t1)
	checkTimes(t, &a2, t0, t1)
	
}

func checkAssessmentResult(t *testing.T, a *model.Agreement, result Result, expectedState model.State, expectedViolatedGts int) {
	if a.State != expectedState {
		t.Errorf("Agreement in unexpected state. Expected: %v. Actual: %v", expectedState, a.State)
	}
	if len(result) != expectedViolatedGts {
		t.Errorf("Unexpected violated GTs. Expected: %v. Actual:%v", expectedViolatedGts, len(result))
	}
}

func checkTimes(t *testing.T, a *model.Agreement, expectedFirst time.Time, expectedLast time.Time) {

	if a.Assessment.FirstExecution != expectedFirst {
		t.Errorf("Unexpected firstExecution. Expected: %v. Actual: %v", expectedFirst, a.Assessment.FirstExecution)
	}
	if a.Assessment.LastExecution != expectedLast {
		t.Errorf("Unexpected lastExecution. Expected: %v. Actual: %v", expectedLast, a.Assessment.LastExecution)
	}
}


func TestAssessExpiredAgreement(t *testing.T) {
	a2 := createAgreement("a02", p1, c2, "Agreement 02", "m >= 0")
	ma := NewSimpleMonitoring(nil)

	a2.State = model.STARTED
	a2.Details.Expiration = t_(-1)
	result := AssessAgreement(&a2, ma, t0)
	if a2.State != model.TERMINATED {
		t.Errorf("Agreement in unexpected state. Expected: terminated. Actual: %v", a2.State)
	}
	if len(result) != 0 {
		t.Errorf("Unexpected violated GTs. Expected: 0. Actual:%v", len(result))
	}
}

func TestEvaluateAgreement(t *testing.T) {
	values := []map[string]MetricValue{
		{"m": MetricValue{Key: "m", Value: 1, DateTime: t_(0)}},
		{"m": MetricValue{Key: "m", Value: -1, DateTime: t_(1)}},
	}
	ma := NewSimpleMonitoring(values)
	invalid, err := EvaluateAgreement(&a1, ma)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	gt := &a1.Details.Guarantees[0]
	gtev := invalid[gt.Name]
	fmt.Printf("%v\n", gtev)
	if len(gtev.Metrics) != 1 {
		t.Errorf("Error in number of violated metrics. Expected: 1. Actual: %v. %v", gtev.Metrics, invalid)
	}
	if len(gtev.Violations) != 1 {
		t.Errorf("Error in number of violations. Expected: 1. Actual: %v. %v", gtev.Violations, invalid)
	}
}

func TestEvaluateAgreementWithWrongValues(t *testing.T) {
	values := []map[string]MetricValue{
		{"n": MetricValue{Key: "n", Value: 1, DateTime: t_(0)}},
	}
	ma := NewSimpleMonitoring(values)
	_, err := EvaluateAgreement(&a1, ma)
	if err == nil {
		t.Errorf("Expected error evaluating agreement")
	}
}

func TestEvaluateGuarantee(t *testing.T) {
	values := []map[string]MetricValue{
		{"m": MetricValue{Key: "m", Value: 1, DateTime: t_(0)}},
		{"m": MetricValue{Key: "m", Value: -1, DateTime: t_(1)}},
	}
	ma := NewSimpleMonitoring(values)
	invalid, err := EvaluateGuarantee(&a1, a1.Details.Guarantees[0], ma)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(invalid) != 1 {
		t.Errorf("Number of invalid metrics. Expected %d. Actual: %d", 1, len(invalid))
		return
	}
	invalidvalue := invalid[0]["m"].Value
	if invalidvalue != -1 {
		t.Errorf("Wrong invalid metric. Expected: %d. Actual: %v", -1, invalidvalue)
	}
}

func TestEvaluateGuaranteeWithWrongExpression(t *testing.T) {
	ma := NewSimpleMonitoring(nil)
	a := createAgreement("a01", p1, c2, "Agreement 01", "wrong expression >= 0")
	_, err := EvaluateGuarantee(&a, a.Details.Guarantees[0], ma)
	if err == nil {
		t.Errorf("Expected error evaluating guarantee")
	}
}

func TestEvaluateGuaranteeWithWrongValues(t *testing.T) {
	values := []map[string]MetricValue{
		{"n": MetricValue{Key: "n", Value: 1, DateTime: t_(0)}},
	}
	ma := NewSimpleMonitoring(values)
	_, err := EvaluateGuarantee(&a1, a1.Details.Guarantees[0], ma)
	if err == nil {
		t.Errorf("Expected error evaluating guarantee")
	}
}

func TestEvaluateExpression(t *testing.T) {
	c := "m >= 0"
	expression, err := govaluate.NewEvaluableExpression(c)
	if err != nil {
		t.Errorf("Error parsing expression '%s': %s", c, err.Error())
	}

	v := createSimpleEvaluationData("m", 1)
	invalid, err := evaluateExpression(expression, v)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(invalid) > 0 {
		t.Errorf("expression: '%s', values:%v", c, v)
	}

	v = createSimpleEvaluationData("m", -1)
	invalid, err = evaluateExpression(expression, v)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(invalid) == 0 {
		t.Errorf("expression: '%s', values:%v", c, v)
	}
	fmt.Printf("%v", invalid)
}

// func TestEvaluationSuccess(t *testing.T) {

// 	monitoring := DummyMonitoring{
// 		Result: EvaluationData{"test_value": 11},
// 	}

// 	failed, err := EvaluateAgreement(a1, monitoring)
// 	if err != nil {
// 		t.Errorf("Error evaluating agreement: %s", err.Error())
// 	}

// 	if len(failed) > 0 {
// 		t.Errorf("Found penalties but none were expected")
// 	}
// }

// func TestEvaluationFailure(t *testing.T) {

// 	monitoring := DummyMonitoring{
// 		Result: EvaluationData{"test_value": 9},
// 	}

// 	failed, err := EvaluateAgreement(a1, monitoring)
// 	if err != nil {
// 		t.Errorf("Error evaluating agreement: %s", err.Error())
// 	}

// 	if len(failed) != 1 {
// 		t.Errorf("Penalty expected but none found")
// 	}
// }

// func TestNullVariables(t *testing.T) {
// 	a := createAgreement("a02", p1, c2, "Agreement", "a > 0 && b > 0")

// 	monitoring := DummyMonitoring{
// 		Result: EvaluationData{"a": 1},
// 	}

// 	failed, err := EvaluateAgreement(a, monitoring)
// 	if err != nil {
// 		t.Errorf("Error evaluating agreement: %s", err.Error())
// 	}
// 	if len(failed) != 0 {
// 		t.Errorf("failed != 0")
// 	}
// }

// func TestGetVars(t *testing.T) {
// 	c := "a > 0 && b > 0"

// 	expression, err := govaluate.NewEvaluableExpression(c)
// 	if err != nil {
// 		t.Errorf("%s", err.Error())
// 	}
// 	if len(expression.Vars()) != 2 {
// 		t.Errorf("number of vars != 2")
// 	} else {
// 		fmt.Printf("%v\n", expression.Vars())
// 	}

// }

// func (m DummyMonitoring) GetValues(vars []string) EvaluationData {
// 	return m.Result
// }

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

func createSimpleEvaluationData(key string, value interface{}) map[string]MetricValue {
	result := make(map[string]MetricValue)
	result[key] = createMonitoringMetric(key, value)
	return result
}

func createMonitoringMetric(key string, value interface{}) MetricValue {
	return MetricValue{
		Key:      key,
		Value:    value,
		DateTime: time.Now(),
	}
}

func t_(second time.Duration) time.Time {
	return t0.Add(time.Second * second)
}
