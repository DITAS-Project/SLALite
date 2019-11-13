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

/*
Package dummyadapter provides an example of MonitoringAdapter. It just
returns a random value each time it is called.

Usage:
	ma := dummyadapter.New()
	ma = ma.Initialize(&agreement)
	for _, gt := range gts {
		for values := range ma.GetValues(gt, ...) {
			...
		}
	}
*/
package dummyadapter

import (
	assessment_model "SLALite/assessment/model"
	"SLALite/assessment/monitor"
	"SLALite/model"

	"math/rand"
	"time"
)

type monitoringAdapter struct {
	agreement *model.Agreement
	size      int
}

// New returns a new Dummy Monitoring Adapter.
func New(size int) monitor.MonitoringAdapter {
	return &monitoringAdapter{
		agreement: nil,
		size:      size,
	}
}

func (ma *monitoringAdapter) Initialize(a *model.Agreement) monitor.MonitoringAdapter {
	result := *ma
	result.agreement = a

	return &result
}

func (ma *monitoringAdapter) GetValues(gt model.Guarantee, vars []string, now time.Time) assessment_model.GuaranteeData {
	result := make(assessment_model.GuaranteeData, ma.size)
	for i := 0; i < ma.size; i++ {
		val := make(assessment_model.ExpressionData)

		for _, key := range vars {
			val[key] = model.MetricValue{
				DateTime: time.Now(),
				Key:      key,
				Value:    rand.Float64(),
			}
		}

		result[i] = val
	}
	return result
}
