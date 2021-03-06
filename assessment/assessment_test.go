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
	assessment_model "SLALite/assessment/model"
	"SLALite/assessment/monitor/simpleadapter"
	"SLALite/model"
	"SLALite/utils"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/Knetic/govaluate"
)

type ValidationNotifier struct {
	Expected map[string]map[string]int
	T        *testing.T
}

var a1 = createAgreement("a01", p1, c2, "Agreement 01", "m >= 0")
var p1 = model.Provider{Id: "p01", Name: "Provider01"}
var c2 = model.Client{Id: "c02", Name: "A client"}
var t0 = time.Now()

var repo = utils.CreateTestRepository()

func (n ValidationNotifier) NotifyViolations(agreement *model.Agreement, result *assessment_model.Result) {
	violations, ok := n.Expected[agreement.Id]
	if ok {
		checkAssessmentResult(n.T, agreement, *result, model.STARTED, violations, nil)
		updated, _ := repo.GetAgreement(agreement.Id)
		if updated != nil {
			checkTimes(n.T, agreement, updated.Assessment.FirstExecution, updated.Assessment.LastExecution)
		} else {
			n.T.Errorf("Can't get agreement %s from repository", agreement.Id)
		}
	} else {
		n.T.Errorf("Can't find test information for agreement %s", agreement.Id)
	}
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func TestAssessActiveAgreements(t *testing.T) {

	var aa1 = createAgreement("aa01", p1, c2, "Agreement aa01", "m >= 10")
	aa1.State = model.STARTED

	guarantees := map[string]string{
		"g1": "m >= 20",
		"g2": "n < 50",
	}
	var aa2 = createAgreementFull("aa02", p1, c2, "Agreement aa02", guarantees, nil)
	aa2.State = model.STARTED

	guarantees = map[string]string{
		"g1": "m >= 20 || n < 50",
	}
	var aa3 = createAgreementFull("aa03", p1, c2, "Agreement aa03", guarantees, nil)
	aa3.State = model.STARTED

	repo.CreateAgreement(&aa1)
	repo.CreateAgreement(&aa2)
	repo.CreateAgreement(&aa3)

	var m1 = assessment_model.GuaranteeData{
		{
			"m": model.MetricValue{Key: "m", Value: 5, DateTime: t_(0)},
			"n": model.MetricValue{Key: "n", Value: 25, DateTime: t_(0)},
		},
		{
			"m": model.MetricValue{Key: "m", Value: 15, DateTime: t_(1)},
			"n": model.MetricValue{Key: "n", Value: 40, DateTime: t_(1)},
		},
		{
			"m": model.MetricValue{Key: "m", Value: 7, DateTime: t_(2)},
			"n": model.MetricValue{Key: "n", Value: 75, DateTime: t_(2)},
		},
	}

	AssessActiveAgreements(repo, simpleadapter.New(m1), ValidationNotifier{Expected: map[string]map[string]int{
		"aa01": map[string]int{
			"TestGuarantee": 2,
		},
		"aa02": map[string]int{
			"g1": 3,
			"g2": 1,
		},
		"aa03": map[string]int{
			"g1": 1,
		},
	}, T: t})
}

func TestAssessAgreement(t *testing.T) {
	a2 := createAgreement("a02", p1, c2, "Agreement 02", "m >= 0")
	values := assessment_model.GuaranteeData{
		{"m": model.MetricValue{Key: "m", Value: 1, DateTime: t_(0)}},
		{"m": model.MetricValue{Key: "m", Value: -1, DateTime: t_(1)}},
	}
	ma := simpleadapter.New(values)

	expected := map[string]int{}
	expectedLast := map[string]model.LastValues{}

	a2.State = model.STOPPED
	result := AssessAgreement(&a2, ma, t0)
	checkAssessmentResult(t, &a2, result, model.STOPPED, expected, expectedLast)

	a2.State = model.TERMINATED
	result = AssessAgreement(&a2, ma, t0)
	checkAssessmentResult(t, &a2, result, model.TERMINATED, expected, expectedLast)

	a2.State = model.STARTED
	expected["TestGuarantee"] = 1
	expectedLast = map[string]model.LastValues{
		"TestGuarantee": model.LastValues{
			"m": values[1]["m"],
		},
	}
	result = AssessAgreement(&a2, ma, t0)
	checkAssessmentResult(t, &a2, result, model.STARTED, expected, expectedLast)
	checkTimes(t, &a2, t0, t0)

	t1 := t_(1)
	result = AssessAgreement(&a2, ma, t1)
	checkTimes(t, &a2, t0, t1)

	// check assessment without values
	values = assessment_model.GuaranteeData{}
	ma = simpleadapter.New(values)
	result = AssessAgreement(&a2, ma, t0)
}

func checkAssessmentResult(t *testing.T, a *model.Agreement,
	result assessment_model.Result, expectedState model.State,
	expectedViolatedGts map[string]int,
	expectedLast map[string]model.LastValues) {

	if a.State != expectedState {
		t.Errorf("Agreement in unexpected state. Expected: %v. Actual: %v", expectedState, a.State)
	}
	if len(result.Violated) != len(expectedViolatedGts) {
		t.Errorf("Unexpected violated GTs for agreement %s. Expected: %v. Actual:%v",
			a.Id, len(expectedViolatedGts), len(result.Violated))
	}
	for gt, numViolations := range expectedViolatedGts {
		gtr, ok := result.Violated[gt]
		if !ok {
			t.Errorf("Expected violation or guarantee %s but not found", gt)
		} else {
			if len(gtr.Violations) != numViolations {
				t.Errorf("Violation number differ for guarantee %s. Expected: %v. Actual %v", gt, numViolations, len(gtr.Violations))
			}
		}
	}
	if expectedLast != nil {

		for gtname := range expectedLast {
			for _, actual := range a.Assessment.GetGuarantee(gtname).LastValues {
				expected := expectedLast[gtname][actual.Key]
				if expected != actual {
					t.Errorf("Unexpected Assessment.LastValues[%s]. Expected: %v; Actual: %v. Assessment=%v",
						gtname, expected, actual, a.Assessment)
				}
			}
		}
	}
}

func checkTimes(t *testing.T, a *model.Agreement, expectedFirst time.Time, expectedLast time.Time) {

	if a.Assessment.FirstExecution.Unix() != expectedFirst.Unix() {
		t.Errorf("Unexpected firstExecution. Expected: %v. Actual: %v", expectedFirst, a.Assessment.FirstExecution)
	}
	if a.Assessment.LastExecution.Unix() != expectedLast.Unix() {
		t.Errorf("Unexpected lastExecution. Expected: %v. Actual: %v", expectedLast, a.Assessment.LastExecution)
	}
}

func TestAssessExpiredAgreement(t *testing.T) {
	a2 := createAgreement("a02", p1, c2, "Agreement 02", "m >= 0")
	ma := simpleadapter.New(nil)

	a2.State = model.STARTED
	expiration := t_(-1)
	a2.Details.Expiration = &expiration
	result := AssessAgreement(&a2, ma, t0)
	if a2.State != model.TERMINATED {
		t.Errorf("Agreement in unexpected state. Expected: terminated. Actual: %v", a2.State)
	}
	if len(result.Violated) != 0 {
		t.Errorf("Unexpected violated GTs. Expected: 0. Actual:%v", len(result.Violated))
	}
}

func TestEvaluateAgreement(t *testing.T) {
	values := assessment_model.GuaranteeData{
		{"m": model.MetricValue{Key: "m", Value: 1, DateTime: t_(0)}},
		{"m": model.MetricValue{Key: "m", Value: -1, DateTime: t_(1)}},
	}
	ma := simpleadapter.New(values)
	invalid, err := EvaluateAgreement(&a1, ma, time.Now())
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	gt := &a1.Details.Guarantees[0]
	gtev := invalid.Violated[gt.Name]
	fmt.Printf("%v\n", gtev)
	if len(gtev.Metrics) != 1 {
		t.Errorf("Error in number of violated metrics. Expected: 1. Actual: %v. %v", gtev.Metrics, invalid)
	}
	if len(gtev.Violations) != 1 {
		t.Errorf("Error in number of violations. Expected: 1. Actual: %v. %v", gtev.Violations, invalid)
	}
	validater := model.NewDefaultValidator(false, true)
	for _, v := range gtev.Violations {
		if errs := v.Validate(validater, model.CREATE); len(errs) != 1 {
			t.Errorf("Validation error in violation: %v", errs)
		}
		if len(v.Values) != 1 {
			t.Errorf("Unexpected Values number: %v", len(v.Values))
		} else {
			metric := v.Values[0]
			if metric.Key != "m" && metric.Value != -1 {
				t.Errorf("Unexpected Values information: [%s,%v]", metric.Key, metric.Value)
			}
		}

	}
}

func TestEvaluateAgreementWithWrongValues(t *testing.T) {
	values := assessment_model.GuaranteeData{
		{"n": model.MetricValue{Key: "n", Value: 1, DateTime: t_(0)}},
	}
	ma := simpleadapter.New(values)
	_, err := EvaluateAgreement(&a1, ma, time.Now())
	if err == nil {
		t.Errorf("Expected error evaluating agreement")
	}
}

func TestEvaluateGuarantee(t *testing.T) {
	values := assessment_model.GuaranteeData{
		{"m": model.MetricValue{Key: "m", Value: 1, DateTime: t_(0)}},
		{"m": model.MetricValue{Key: "m", Value: -1, DateTime: t_(1)}},
	}
	ma := simpleadapter.New(values)
	invalid, last, err := EvaluateGuarantee(&a1, a1.Details.Guarantees[0], ma, time.Now())
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
	if last["m"] != values[1]["m"] {
		t.Errorf("Unexpected lastvalues. Expected: %v; Actual: %v", values[1], last)
	}
}

func TestEvaluateGuaranteeWithWrongExpression(t *testing.T) {
	ma := simpleadapter.New(nil)
	a := createAgreement("a01", p1, c2, "Agreement 01", "wrong expression >= 0")
	_, _, err := EvaluateGuarantee(&a, a.Details.Guarantees[0], ma, time.Now())
	if err == nil {
		t.Errorf("Expected error evaluating guarantee")
	}
}

func TestEvaluateGuaranteeWithWrongValues(t *testing.T) {
	values := assessment_model.GuaranteeData{
		{"n": model.MetricValue{Key: "n", Value: 1, DateTime: t_(0)}},
	}
	ma := simpleadapter.New(values)
	_, _, err := EvaluateGuarantee(&a1, a1.Details.Guarantees[0], ma, time.Now())
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

func createAgreementFull(aid string, provider model.Provider, client model.Client, name string, constraints map[string]string, expiration *time.Time) model.Agreement {
	agreement := model.Agreement{
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
			Guarantees: make([]model.Guarantee, len(constraints)),
		},
	}

	var i = 0
	for k, v := range constraints {
		agreement.Details.Guarantees[i] = model.Guarantee{Name: k, Constraint: v}
		i++
	}

	return agreement
}

func createAgreement(aid string, provider model.Provider, client model.Client, name string, constraint string) model.Agreement {
	return createAgreementFull(aid, provider, client, name, map[string]string{"TestGuarantee": constraint}, nil)
}

func createSimpleEvaluationData(key string, value interface{}) assessment_model.ExpressionData {
	result := make(assessment_model.ExpressionData)
	result[key] = createMonitoringMetric(key, value)
	return result
}

func createMonitoringMetric(key string, value interface{}) model.MetricValue {
	return model.MetricValue{
		Key:      key,
		Value:    value,
		DateTime: time.Now(),
	}
}

func t_(second time.Duration) time.Time {
	return t0.Add(time.Second * second)
}
