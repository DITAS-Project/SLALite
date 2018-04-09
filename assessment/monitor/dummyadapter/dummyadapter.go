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
package dummyadapter

import (
	"SLALite/assessment/monitor"
	"SLALite/model"

	"math/rand"
	"time"

	"github.com/Knetic/govaluate"
)

type monitoringAdapter struct {
	agreement *model.Agreement
	i         int
}

func New() monitor.MonitoringAdapter {
	return &monitoringAdapter{
		agreement: nil,
		i:         0,
	}
}

func (ma *monitoringAdapter) Initialize(a *model.Agreement) {
	ma.agreement = a
}

func (ma *monitoringAdapter) NextValues(gt model.Guarantee) map[string]monitor.MetricValue {
	result := make(map[string]monitor.MetricValue)

	if ma.i == 1 {
		return nil
	}
	ma.i = ma.i + 1

	/*
	 * The following means that the adapter have knowledge about the evaluation internals,
	 * but there is a problem with cyclic dependencies if importing "/assessment"
	 */
	expression, err := govaluate.NewEvaluableExpression(gt.Constraint)
	if err != nil {
		/* TODO */
	}

	for _, key := range expression.Vars() {
		result[key] = monitor.MetricValue{
			DateTime: time.Now(),
			Key:      key,
			Value:    rand.Float64(),
		}
	}
	return result
}
