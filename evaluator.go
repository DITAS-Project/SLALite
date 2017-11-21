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
package main

import (
	"SLALite/model"
	"strings"

	"github.com/labstack/gommon/log"

	"github.com/oleksandr/conditions"
)

func evaluate(condition string, data map[string]interface{}) (bool, error) {

	p := conditions.NewParser(strings.NewReader(condition))
	expr, err := p.Parse()
	if err != nil {
		return false, err
	}

	r, err := conditions.Evaluate(expr, data)
	if err != nil {
		return false, err
	}

	return r, nil

}

func evaluateAgreement(agreement model.Agreement,
	data map[string]map[string]interface{}) ([]model.Guarantee, error) {

	guarantees := agreement.Text.Guarantees
	failed := []model.Guarantee{}

	for _, guarantee := range guarantees {

		guaranteeData := data[guarantee.Name]
		if guaranteeData != nil {
			res, err := evaluate(guarantee.Constraint, guaranteeData)
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
