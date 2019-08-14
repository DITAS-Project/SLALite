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
	assessment_model "SLALite/assessment/model"
	"SLALite/model"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/Knetic/govaluate"
	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
)

type DitasViolation struct {
	VDCId   string              `json:"vdcId"`
	Method  string              `json:"methodId"`
	Metrics []model.MetricValue `json:"metrics"`
}

type DitasNotifier struct {
	VDCId      string
	NotifyUrl  string
	Client     *resty.Client
	Violations []DitasViolation
}

func NewNotifier(vdcId, url string) *DitasNotifier {
	return &DitasNotifier{
		VDCId:     vdcId,
		NotifyUrl: url + "/NotifyViolation",
		Client:    resty.New(),
	}
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
	case "==":
		return value == threshold, nil
	case ">":
		return value > threshold, nil
	case ">=":
		return value >= threshold, nil
	}
	return false, errors.New("Comparator not supported: " + comparator)
}

func (n *DitasNotifier) filterValues(methodId string, result *assessment_model.Result) []DitasViolation {
	violations := make([]DitasViolation, 0)
	violationMap := make(map[string][]model.MetricValue)
	for _, grResults := range result.Violated {
		for _, violation := range grResults.Violations {
			valueMap := make(map[string]model.MetricValue)

			for _, metricValue := range violation.Values {
				valueMap[metricValue.Key] = metricValue
			}

			expression, err := govaluate.NewEvaluableExpression(violation.Constraint)
			if err == nil {
				violationInformation, ok := violationMap[violation.AgreementId]
				if !ok {
					violationInformation = make([]model.MetricValue, 0)
				}
				tokens := expression.Tokens()
				for i, token := range tokens {
					if token.Kind == govaluate.VARIABLE && (i < len(tokens)-2) && (tokens[i+1].Kind == govaluate.COMPARATOR && tokens[i+2].Kind == govaluate.NUMERIC) {
						variable := token.Value.(string)
						comparator := tokens[i+1].Value.(string)
						threshold := tokens[i+2].Value
						value, found := valueMap[variable]
						if found {
							assessed, err := evaluate(comparator, threshold, value.Value)
							if err == nil && !assessed {
								violationInformation = append(violationInformation, value)
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
				violationMap[violation.AgreementId] = violationInformation
			} else {
				log.Errorf("Error tokenizing expression %s: %s", violation.Constraint, err.Error())
			}
		}
	}
	for k, v := range violationMap {
		violations = append(violations, DitasViolation{
			Method:  k,
			VDCId:   n.VDCId,
			Metrics: v,
		})
	}
	return violations
}

func (n *DitasNotifier) NotifyViolations(agreement *model.Agreement, result *assessment_model.Result) {
	logger := log.WithField("agreement", agreement.Id)
	logger.Debugf("Notifying %d violations", len(result.GetViolations()))
	if n.NotifyUrl != "" {
		n.Violations = n.filterValues(agreement.Id, result)
		rawJSON, err := json.Marshal(n.Violations)
		if err != nil {
			logger.WithError(err).Errorf("Error marshaling violations of agreement %s", agreement.Id)
			return
		}
		data := map[string]string{
			"violations": string(rawJSON),
		}
		logger.Debugf("Got %d violations after filtering", len(n.Violations))
		_, err = n.Client.R().SetFormData(data).Post(n.NotifyUrl)
		if err != nil {
			log.WithError(err).Errorf("Error notifying violations of SLA %s", agreement.Id)
		}
	}
}
