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
package enforcement

import (
	"SLALite/model"

	"github.com/Knetic/govaluate"
	"github.com/labstack/gommon/log"
)

type EvaluationData map[string]interface{}

func Evaluate(condition string, data EvaluationData) (bool, error) {

	expression, err := govaluate.NewEvaluableExpression(condition)
	if err == nil {
		result, error := expression.Evaluate(data)
		return result.(bool), error
	}

	return false, err
}

func EvaluateAgreementValues(agreement model.Agreement,
	data map[string]EvaluationData) ([]model.Guarantee, error) {

	guarantees := agreement.Text.Guarantees
	failed := []model.Guarantee{}

	for _, guarantee := range guarantees {

		guaranteeData := data[guarantee.Name]
		if guaranteeData != nil {
			res, err := Evaluate(guarantee.Constraint, guaranteeData)
			if err != nil {
				log.Warn("Error evaluating expression " + guarantee.Constraint + ": " + err.Error())
			} else {
				if !res {
					failed = append(failed, guarantee)
				}
			}
		}

	}

	return failed, nil
}

func EvaluateAgreement(agreement model.Agreement, adapter MonitoringAdapter) ([]model.Guarantee, error) {

	failed := []model.Guarantee{}
	guarantees := agreement.Text.Guarantees

	for _, guarantee := range guarantees {
		expression, err := govaluate.NewEvaluableExpression(guarantee.Constraint)
		if err == nil {
			vars := expression.Vars()
			values := adapter.GetValues(vars)

			result, err := expression.Evaluate(values)
			if err != nil {
				log.Warn("Error evaluating expression " + guarantee.Constraint + ": " + err.Error())
			} else {
				if !result.(bool) {
					failed = append(failed, guarantee)
				}
			}

		}
	}

	return failed, nil

}
