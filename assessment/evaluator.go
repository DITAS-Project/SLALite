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

// GuaranteeData represents the set of values needed to evaluate an expression at several points
// in time
type GuaranteeData []ExpressionData

func EvaluateAgreement(a model.Agreement, ma MonitoringAdapter) (map[*model.Guarantee]GuaranteeData, error) {
	ma.Initialize(a)

	result := make(map[*model.Guarantee]GuaranteeData)
	gts := a.Details.Guarantees

	for _, &gt := range gts {
		failed, err := EvaluateGuarantee(a, gt, ma)
		if err != nil {
			log.Warn("Error evaluating expression " + gt.Constraint + ": " + err.Error())			
			return nil, err
		}
		result[&gt] = failed
	}

	return result, nil
}

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

func evaluateExpression(expression *govaluate.EvaluableExpression, values map[string]MetricValue) (ExpressionData, error) {

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

// func Evaluate(condition string, data EvaluationData) (bool, error) {

// 	expression, err := govaluate.NewEvaluableExpression(condition)
// 	if err == nil {
// 		result, error := expression.Evaluate(data)
// 		return result.(bool), error
// 	}

// 	return false, err
// }

// func EvaluateAgreementValues(agreement model.Agreement,
// 	data map[string]EvaluationData) ([]model.Guarantee, error) {

// 	guarantees := agreement.Details.Guarantees
// 	failed := []model.Guarantee{}

// 	for _, guarantee := range guarantees {

// 		guaranteeData := data[guarantee.Name]
// 		if guaranteeData != nil {
// 			res, err := Evaluate(guarantee.Constraint, guaranteeData)
// 			if err != nil {
// 				log.Warn("Error evaluating expression " + guarantee.Constraint + ": " + err.Error())
// 			} else {
// 				if !res {
// 					failed = append(failed, guarantee)
// 				}
// 			}
// 		}

// 	}

// 	return failed, nil
// }

// func EvaluateAgreement(agreement model.Agreement, adapter MonitoringAdapter) ([]model.Guarantee, error) {

// 	failed := []model.Guarantee{}
// 	guarantees := agreement.Details.Guarantees

// 	for _, guarantee := range guarantees {
// 		expression, err := govaluate.NewEvaluableExpression(guarantee.Constraint)
// 		if err == nil {
// 			vars := expression.Vars()
// 			values := adapter.GetValues(vars)

// 			result, err := expression.Evaluate(values)
// 			if err != nil {
// 				log.Warn("Error evaluating expression " + guarantee.Constraint + ": " + err.Error())
// 			} else {
// 				if !result.(bool) {
// 					failed = append(failed, guarantee)
// 				}
// 			}

// 		}
// 	}

// 	return failed, nil

// }
