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
	"SLALite/assessment/monitor"
	"SLALite/assessment/notifier"
	"SLALite/model"
	log2 "log"
	"time"

	"github.com/Knetic/govaluate"
	log "github.com/labstack/gommon/log"
)

//AssessActiveAgreements will get the active agreements from the provided repository and assess them, notifying about violations with the provided notifier.
func AssessActiveAgreements(repo model.IRepository, ma monitor.MonitoringAdapter, not notifier.ViolationNotifier) {
	agrements, err := repo.GetActiveAgreements()
	if err != nil {
		log.Errorf("Error getting active agreements: " + err.Error())
	} else {
		for _, agreement := range agrements {
			result := AssessAgreement(&agreement, ma, time.Now())
			repo.UpdateAgreement(&agreement)
			not.NotifyViolations(&agreement, &result)
		}
	}
}

// AssessAgreement is the process that assess an agreement. The process is:
// 1. Check expiration date
// 2. Evaluate metrics if agreement is started
// 3. Set LastExecution time.
//
// The output is:
// - parameter a is modified
// - evaluation results are the function return (violated metrics and raised violations).
//   a guarantee term is filled in the result only if there are violations.
//
// The function results are not persisted. The output must be persisted/handled accordingly.
// E.g.: agreement and violations must be persisted to DB. Violations must be notified to
// observers
func AssessAgreement(a *model.Agreement, ma monitor.MonitoringAdapter, now time.Time) notifier.Result {
	var result notifier.Result
	var err error

	if a.Details.Expiration.Before(now) {
		// agreement has expired
		a.State = model.TERMINATED
	}

	if a.State == model.STARTED {
		result, err = EvaluateAgreement(a, ma)
		if err != nil {
			// TODO
		}
		if a.Assessment.FirstExecution.IsZero() {
			a.Assessment.FirstExecution = now
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
func EvaluateAgreement(a *model.Agreement, ma monitor.MonitoringAdapter) (notifier.Result, error) {
	ma.Initialize(a)

	result := make(notifier.Result)
	gts := a.Details.Guarantees

	for _, gt := range gts {
		failed, err := EvaluateGuarantee(a, gt, ma)
		if err != nil {
			log.Warn("Error evaluating expression " + gt.Constraint + ": " + err.Error())
			return nil, err
		}
		if len(failed) > 0 {
			violations := EvaluateGtViolations(a, gt, failed)
			gtResult := notifier.EvaluationGtResult{
				Metrics:    failed,
				Violations: violations,
			}
			result[gt.Name] = gtResult
		}
	}

	return result, nil
}

// EvaluateGuarantee evaluates a guarantee term of an Agreement
// (see EvaluateAgreement)
//
// Returns the metrics that failed the GT constraint.
func EvaluateGuarantee(a *model.Agreement, gt model.Guarantee, ma monitor.MonitoringAdapter) (notifier.GuaranteeData, error) {
	failed := make(notifier.GuaranteeData, 0, 1)

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
func EvaluateGtViolations(a *model.Agreement, gt model.Guarantee, violated notifier.GuaranteeData) []model.Violation {
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
func evaluateExpression(expression *govaluate.EvaluableExpression, values notifier.ExpressionData) (notifier.ExpressionData, error) {

	log2.Printf("Evaluating %v", values)
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
