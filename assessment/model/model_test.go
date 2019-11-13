/*
Copyright 2018 Atos

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

package model

import (
	"SLALite/model"
	"SLALite/utils"
	"os"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func TestGetViolations(test *testing.T) {

	var t = utils.Timeline{T0: time.Now()}

	values := GuaranteeData{
		ExpressionData{
			"m": model.MetricValue{Key: "m", Value: 1, DateTime: t.T(1)},
		},
	}

	r := Result{
		Violated: map[string]EvaluationGtResult{
			"gt": EvaluationGtResult{
				Metrics: values,
				Violations: []model.Violation{
					model.Violation{
						/* We are not interested in actual violation contents */
					},
				},
			},
		},
	}

	if violations := r.GetViolations(); len(violations) != 1 {
		test.Errorf("Unexpected number of violations. Expected: %d, Actual: %d", 1, len(violations))
	}
}
