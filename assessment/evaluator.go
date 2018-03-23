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
package assessment

import (
	"SLALite/model"
	"fmt"
	"time"

	"github.com/Knetic/govaluate"
	"github.com/labstack/gommon/log"
)

// MetricValue is the SLALite representation of a metric value.
type MetricValue struct {
	Key      string
	Value    interface{}
	DateTime time.Time
}

func (v *MetricValue) String() string {
	return fmt.Sprintf("{Key: %s, Value: %v, DateTime: %v}", v.Key, v.Value, v.DateTime)
}

// ExpressionData represents the set of values needed to evaluate an expression at a single time
type ExpressionData map[string]MetricValue

// GuaranteeData represents the list of values needed to evaluate an expression at several points
// in time
type GuaranteeData []ExpressionData

// EvaluationGtResult is the result of the evaluation of a guarantee term
type EvaluationGtResult struct {
	Metrics    GuaranteeData     // violated metrics
	Violations []model.Violation // violations occurred as of violated metrics
}

// Result is the result of the agreement assessment
type Result map[string]EvaluationGtResult

// AssessAgreement is the process that assess an agreement. The process is:
// 1. Check expiration date
// 2. Evaluate metrics if agreement is started
// 3. Set LastExecution time.
//
// The output is:
// - parameter a is modified
// - evaluation results are the function return (violated metrics and raised violations)
//
// The function results are not persisted. The output must be persisted/handled accordingly.
// E.g.: agreement and violations must be persisted to DB. Violations must be notified to
// observers
func AssessAgreement(a *model.Agreement, ma MonitoringAdapter, now time.Time) Result {
	var result Result
	var err error

	if a.Details.Expiration.Before(now) {
		// agreement has expired
		a.State = model.TERMINATED
	}

	if a.State == model.STARTED {
		result, err = EvaluateAgreement(*a, ma)
		if err != nil {
			// TODO
		}
		a.Assessment.LastExecution = now
	}
	return result
}

// EvaluateAgreement evaluates the guarantee terms of an agreement. The metric values
// are retrieved from a MonitoringAdapter.
// The MonitoringAdapter must feed the process correctly
// (e.g. if the constraint of a guarantee term is of the type "A>B && C>D", the
// MonitoringAdapter must supply pairs of values).
func EvaluateAgreement(a model.Agreement, ma MonitoringAdapter) (Result, error) {
	ma.Initialize(a)

	result := make(Result)
	gts := a.Details.Guarantees

	for _, gt := range gts {
		failed, err := EvaluateGuarantee(a, gt, ma)
		if err != nil {
			log.Warn("Error evaluating expression " + gt.Constraint + ": " + err.Error())
			return nil, err
		}
		violations := EvaluateGtViolations(a, gt, failed)
		gtResult := EvaluationGtResult{
			Metrics:    failed,
			Violations: violations,
		}
		result[gt.Name] = gtResult
	}

	return result, nil
}

// EvaluateGuarantee evaluates a guarantee term of an Agreement
// (see EvaluateAgreement)
//
// Returns the metrics that failed the GT constraint.
func EvaluateGuarantee(a model.Agreement, gt model.Guarantee, ma MonitoringAdapter) (GuaranteeData, error) {
	failed := make(GuaranteeData, 0, 1)

	expression, err := govaluate.NewEvaluableExpression(gt.Constraint)
	if err != nil {
		return nil, err
	}

	for values := ma.NextValues(gt); values != nil; values = ma.NextValues(gt) {
		aux, err := evaluateExpression(expression, values)
		if err != nil {
			log.Warn("Error evaluating expression " + gt.Constraint + ": " + err.Error())
			return nil, err
		}
		if aux != nil {
			failed = append(failed, aux)
		}
	}
	return failed, nil
}

// EvaluateGtViolations creates violations for the detected violated metrics in EvaluateGuarantee
func EvaluateGtViolations(a model.Agreement, gt model.Guarantee, violated GuaranteeData) []model.Violation {
	gtv := make([]model.Violation, 0, len(violated))
	for _, tuple := range violated {
		// find newer metric
		var d *time.Time
		for _, m := range tuple {
			if d == nil || m.DateTime.After(*d) {
				d = &m.DateTime
			}
		}
		v := model.Violation{
			AgreementId: a.Id,
			Guarantee:   gt.Name,
			Datetime:    *d,
		}
		gtv = append(gtv, v)
	}
	return gtv
}

// evaluateExpression evaluate a GT expression at a single point in time with a tuple of metric values
// (one value per variable in GT expresssion)
//
// The result is: the values if the expression is false (i.e., the failing values) ,
// or nil if expression was true
func evaluateExpression(expression *govaluate.EvaluableExpression, values ExpressionData) (ExpressionData, error) {

	evalues := make(map[string]interface{})
	for key, value := range values {
		evalues[key] = value.Value
	}
	result, err := expression.Evaluate(evalues)

	if err == nil && !result.(bool) {
		return values, nil
	}
	return nil, err
}
