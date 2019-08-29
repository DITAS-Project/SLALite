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

// Violation contains information about the violation of an SLA, including the metrics that made it fail
type Violation struct {
	VDCId   string              `json:"vdcId"`
	Method  string              `json:"methodId"`
	Metrics []model.MetricValue `json:"metrics"`
}

// Notifier is the default Ditas Notifier that will inform the DS4M of violations
type Notifier struct {
	VDCId      string
	NotifyURL  string
	Client     *resty.Client
	Violations []Violation
}

// NewNotifier creates a new Ditas notifier that will use the VDC identifier and DS4M URL provided as parameters
func NewNotifier(vdcID, url string) *Notifier {
	return &Notifier{
		VDCId:     vdcID,
		NotifyURL: url + "/NotifyViolation",
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

// filterValues will filter those metric values that don't meet its threshold in the guarantee
// and may be the responsibles for the failure of the evaluation and so, of the violation.
func (n *Notifier) filterValues(methodID string, result *assessment_model.Result) []Violation {
	violations := make([]Violation, 0)
	violationMap := make(map[string][]model.MetricValue)

	// Iterate over the guarantees that were violated
	for _, grResults := range result.Violated {

		// Iterate over the violations found of that guarantee
		for _, violation := range grResults.Violations {

			// Make a map of each metric. We only have one average value per metric so that's fine.
			valueMap := make(map[string]model.MetricValue)

			for _, metricValue := range violation.Values {
				valueMap[metricValue.Key] = metricValue
			}

			// Make a govaluate expression from the guarantee to re-evaluate it
			expression, err := govaluate.NewEvaluableExpression(violation.Constraint)
			if err == nil {
				violationInformation, ok := violationMap[violation.AgreementId]
				if !ok {
					violationInformation = make([]model.MetricValue, 0)
				}
				tokens := expression.Tokens()
				// Go over tokens to find expressions of type <variable> <operator> <value> i.e. availability >= 90
				for i, token := range tokens {
					if token.Kind == govaluate.VARIABLE && (i < len(tokens)-2) && (tokens[i+1].Kind == govaluate.COMPARATOR && tokens[i+2].Kind == govaluate.NUMERIC) {
						variable := token.Value.(string)
						comparator := tokens[i+1].Value.(string)
						threshold := tokens[i+2].Value
						value, found := valueMap[variable]
						if found {
							// Evaluate the comparison to see if it violates the threshold that was defined
							assessed, err := evaluate(comparator, threshold, value.Value)
							if err == nil && !assessed {
								// If so, add it to the list of values that will be sent
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

	// Transform the map to a list to send to DS4M
	for k, v := range violationMap {
		violations = append(violations, Violation{
			Method:  k,
			VDCId:   n.VDCId,
			Metrics: v,
		})
	}
	return violations
}

// NotifyViolations calls the DS4M if there is any violation in the results provided by the assessment process
func (n *Notifier) NotifyViolations(agreement *model.Agreement, result *assessment_model.Result) {
	logger := log.WithField("agreement", agreement.Id)
	logger.Debugf("Notifying %d violations", len(result.GetViolations()))
	if n.NotifyURL != "" {
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
		_, err = n.Client.R().SetFormData(data).Post(n.NotifyURL)
		if err != nil {
			log.WithError(err).Errorf("Error notifying violations of SLA %s", agreement.Id)
		}
	}
}
