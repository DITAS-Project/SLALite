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
Package simpleadapter provides an example of MonitoringAdapter
that returns the same data passed on construction

Usage:
	ma := simpleadapter.New()
	ma = ma.Initialize(&agreement)
	for _, gt := range gts {
		for values := range ma.GetValues(gt, ...) {
			...
		}
	}
*/
package simpleadapter

import (
	assessment_model "SLALite/assessment/model"
	"SLALite/assessment/monitor"
	"SLALite/model"
	"time"
)

// ArrayMonitoringAdapter implements MonitoringAdapter
type ArrayMonitoringAdapter struct {
	agreement *model.Agreement
	values    assessment_model.GuaranteeData
}

// New constructs an ArrayMonitoringAdapter that returns the parameter "values" on
// GetValues() method
func New(values assessment_model.GuaranteeData) *ArrayMonitoringAdapter {
	return &ArrayMonitoringAdapter{
		agreement: nil,
		values:    values,
	}
}

// Initialize implements monitor.MonitoringAdapter.Initialize
func (ma *ArrayMonitoringAdapter) Initialize(a *model.Agreement) monitor.MonitoringAdapter {
	result := *ma
	result.agreement = a
	return &result
}

// GetValues implements monitor.MonitoringAdapter.GetValues
func (ma *ArrayMonitoringAdapter) GetValues(gt model.Guarantee, vars []string, now time.Time) assessment_model.GuaranteeData {
	return ma.values
}
