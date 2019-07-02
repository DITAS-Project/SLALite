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

// Package assessment contains the core code that evaluates the agreements.
package assessment

import (
	amodel "SLALite/assessment/model"
	"SLALite/assessment/monitor"
	"SLALite/assessment/notifier"
	"SLALite/model"
	"time"

	"github.com/Knetic/govaluate"
	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetLevel(log.DebugLevel)
}

//AssessActiveAgreements will get the active agreements from the provided repository and assess them, notifying about violations with the provided notifier.
func AssessActiveAgreements(repo model.IRepository, ma monitor.MonitoringAdapter, not notifier.ViolationNotifier) {
	agreements, err := repo.GetAgreementsByState(model.STARTED, model.STOPPED)
	if err != nil {
		log.Errorf("Error getting active agreements: %s", err.Error())
	} else {
		log.Printf("AssessActiveAgreements(). %d agreements to evaluate", len(agreements))
		for _, agreement := range agreements {
			result := AssessAgreement(&agreement, ma, time.Now())
			repo.UpdateAgreement(&agreement)
			if not != nil && len(result.Violated) > 0 {
				not.NotifyViolations(&agreement, &result)
			}
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
func AssessAgreement(a *model.Agreement, ma monitor.MonitoringAdapter, now time.Time) amodel.Result {
	var result amodel.Result
	var err error

	log.Debugf("AssessAgreement(%s)", a.Id)
	if a.Details.Expiration != nil && a.Details.Expiration.Before(now) {
		// agreement has expired
		a.State = model.TERMINATED
	}

	if a.State == model.STARTED {
		result, err = EvaluateAgreement(a, ma)
		if err != nil {
			log.Warn("Error evaluating agreement " + a.Id + ": " + err.Error())
			return result
		}
		if a.Assessment.FirstExecution.IsZero() {
			a.Assessment.FirstExecution = now
		}
		a.Assessment.LastExecution = now
		for _, d := range result.LastValues {
			for _, v := range d {
				a.Assessment.SetLastValue(v.Key, v)
			}
		}
	}
	return result
}

// EvaluateAgreement evaluates the guarantee terms of an agreement. The metric values
// are retrieved from a MonitoringAdapter.
// The MonitoringAdapter must feed the process correctly
// (e.g. if the constraint of a guarantee term is of the type "A>B && C>D", the
// MonitoringAdapter must supply pairs of values).
func EvaluateAgreement(a *model.Agreement, ma monitor.MonitoringAdapter) (amodel.Result, error) {
	ma = ma.Initialize(a)

	log.Debugf("EvaluateAgreement(%s)", a.Id)
	result := amodel.Result{
		Violated:   map[string]amodel.EvaluationGtResult{},
		LastValues: map[string]amodel.ExpressionData{},
	}
	gts := a.Details.Guarantees

	for _, gt := range gts {
		failed, lastvalues, err := EvaluateGuarantee(a, gt, ma)
		if err != nil {
			log.Warn("Error evaluating expression " + gt.Constraint + ": " + err.Error())
			return amodel.Result{}, err
		}
		if len(failed) > 0 {
			violations := EvaluateGtViolations(a, gt, failed)
			gtResult := amodel.EvaluationGtResult{
				Metrics:    failed,
				Violations: violations,
			}
			result.Violated[gt.Name] = gtResult
		}
		result.LastValues[gt.Name] = lastvalues
	}
	return result, nil
}

// EvaluateGuarantee evaluates a guarantee term of an Agreement
// (see EvaluateAgreement)
//
// Returns the metrics that failed the GT constraint.
func EvaluateGuarantee(a *model.Agreement,
	gt model.Guarantee,
	ma monitor.MonitoringAdapter) (
	failed []amodel.ExpressionData, last amodel.ExpressionData, err error) {

	log.Debugf("EvaluateGuarantee(%s, %s)", a.Id, gt.Name)
	failed = make(amodel.GuaranteeData, 0, 1)

	expression, err := govaluate.NewEvaluableExpression(gt.Constraint)
	if err != nil {
		log.Warnf("Error parsing expression '%s'", gt.Constraint)
		return nil, nil, err
	}
	values := ma.GetValues(gt, expression.Vars())
	for _, value := range values {
		aux, err := evaluateExpression(expression, value)
		if err != nil {
			log.Warn("Error evaluating expression " + gt.Constraint + ": " + err.Error())
			return nil, nil, err
		}
		if aux != nil {
			failed = append(failed, aux)
		}
	}
	if len(values) > 0 {
		last = values[len(values)-1]
	}
	return failed, last, nil
}

// EvaluateGtViolations creates violations for the detected violated metrics in EvaluateGuarantee
func EvaluateGtViolations(a *model.Agreement, gt model.Guarantee, violated amodel.GuaranteeData) []model.Violation {
	gtv := make([]model.Violation, 0, len(violated))
	for _, tuple := range violated {
		// build values map and find newer metric
		var d *time.Time
		var values = make([]model.MetricValue, 0, len(tuple))
		for _, m := range tuple {
			values = append(values, m)
			if d == nil || m.DateTime.After(*d) {
				d = &m.DateTime
			}
		}
		v := model.Violation{
			AgreementId: a.Id,
			Guarantee:   gt.Name,
			Datetime:    *d,
			Constraint:  gt.Constraint,
			Values:      values,
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
func evaluateExpression(expression *govaluate.EvaluableExpression, values amodel.ExpressionData) (amodel.ExpressionData, error) {

	evalues := make(map[string]interface{})
	for key, value := range values {
		evalues[key] = value.Value
	}
	result, err := expression.Evaluate(evalues)
	log.Debugf("Evaluating expression '%v'=%v with values %v", expression, result, values)

	if err == nil && !result.(bool) {
		return values, nil
	}
	return nil, err
}
