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
	"fmt"
	"strconv"
	"time"

	"github.com/Knetic/govaluate"
	log "github.com/sirupsen/logrus"
)

type GuaranteeViolation struct {
	Constraint    string
	GuaranteeId   string
	ViolationTime time.Time
	Values        map[string]interface{}
}

type DitasViolation struct {
	Method             string
	GuaranteeViolation []GuaranteeViolation
}

type DitasNotifier struct {
	Violations DitasViolation
}

func toFloat(value interface{}) (float64, error) {
	return strconv.ParseFloat(fmt.Sprint(value), 64)
}

func evaluate(comparator string, thresholdIf interface{}, valueIf interface{}) (bool, error) {
	threshold, err := toFloat(thresholdIf)
	if err != nil {
		return false, errors.New("Can't parse threshold value: " + err.Error())
	}
	value, err := toFloat(valueIf)
	if err != nil {
		return false, errors.New("Can't parse value: " + err.Error())
	}
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

func (n *DitasNotifier) filterValues(methodId string, result *assessment_model.Result) {
	violations := DitasViolation{
		Method:             methodId,
		GuaranteeViolation: make([]GuaranteeViolation, 0),
	}
	for _, grResults := range *result {
		for _, violation := range grResults.Violations {
			gtViolation := GuaranteeViolation{
				GuaranteeId:   violation.Guarantee,
				Constraint:    violation.Constraint,
				ViolationTime: violation.Datetime,
				Values:        make(map[string]interface{}),
			}
			expression, err := govaluate.NewEvaluableExpression(violation.Constraint)
			if err == nil {
				tokens := expression.Tokens()
				for i, token := range tokens {
					if token.Kind == govaluate.VARIABLE && (i < len(tokens)-2) && (tokens[i+1].Kind == govaluate.COMPARATOR && tokens[i+2].Kind == govaluate.NUMERIC) {
						variable := token.Value.(string)
						comparator := tokens[i+1].Value.(string)
						threshold := tokens[i+2].Value
						value, found := violation.Values[variable]
						if found {
							assessed, err := evaluate(comparator, threshold, value)
							if err == nil && !assessed {
								gtViolation.Values[variable] = value
							} else {
								if err != nil {
									log.Errorf("Error assessing expression %s: %s", violation.Constraint, err.Error())
								}
							}
						} else {
							log.Errorf("Can't find value for variable %s", variable)
						}
					}
				}
				violations.GuaranteeViolation = append(violations.GuaranteeViolation, gtViolation)
			} else {
				log.Errorf("Error tokenizing expression %s: %s", violation.Constraint, err.Error())
			}
		}
	}
	n.Violations = violations
}

func (n *DitasNotifier) NotifyViolations(agreement *model.Agreement, result *assessment_model.Result) {
	n.filterValues(agreement.Id, result)
}
