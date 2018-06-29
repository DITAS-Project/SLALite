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

package ditas

import (
	assessment_model "SLALite/assessment/model"
	"SLALite/model"
	"errors"

	"github.com/Knetic/govaluate"
	log "github.com/sirupsen/logrus"
)

type DitasNotifier struct {
	Result *assessment_model.Result
}

func evaluate(comparator string, threshold float64, value float64) (bool, error) {
	switch comparator {
	case "<":
		return value < threshold, nil
	case "<=":
		return value <= threshold, nil
	case "=":
		return value == threshold, nil
	case ">":
		return value > threshold, nil
	case ">=":
		return value >= threshold, nil
	}
	return false, errors.New("Comparator not supported: " + comparator)
}

func filterValues(result *assessment_model.Result) {
	for _, violation := range result.GetViolations() {
		expression, err := govaluate.NewEvaluableExpression(violation.Constraint)
		if err == nil {
			tokens := expression.Tokens()
			for i, token := range tokens {
				if token.Kind == govaluate.VARIABLE && (i < len(tokens)-2) && (tokens[i+1].Kind == govaluate.COMPARATOR && tokens[i+2].Kind == govaluate.NUMERIC) {
					variable := token.Value.(string)
					comparator := tokens[i+1].Value.(string)
					threshold := tokens[i+2].Value.(float64)
					value, found := violation.Values[variable]
					if found {
						assessed, err := evaluate(comparator, threshold, value.(float64))
						if err == nil && assessed {
							delete(violation.Values, variable)
						} else {
							log.Errorf("Error assessing expression %s: %s", violation.Constraint, err.Error())
						}
					} else {
						log.Errorf("Can't find value for variable %s", variable)
					}
				}
			}
		} else {
			log.Errorf("Error tokenizing expression %s: %s", violation.Constraint, err.Error())
		}
	}
}

func (n DitasNotifier) NotifyViolations(agreement *model.Agreement, result *assessment_model.Result) {
	filterValues(result)
	n.Result = result
}
