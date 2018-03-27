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
package assessment

import "SLALite/model"

//MonitoringAdapter is an interface which should be implemented per monitoring solution
type MonitoringAdapter interface {
	Initialize(a *model.Agreement)
	// GetValues(vars []string) EvaluationData
	NextValues(gt model.Guarantee) map[string]MetricValue
}

type ArrayMonitoringAdapter struct {
	agreement *model.Agreement
	values    []map[string]MetricValue
	i         int
}

func NewSimpleMonitoring(values []map[string]MetricValue) *ArrayMonitoringAdapter {
	return &ArrayMonitoringAdapter{
		agreement: nil,
		values:    values,
		i:         0,
	}
}

func (ma *ArrayMonitoringAdapter) Initialize(a *model.Agreement) {
	ma.agreement = a
}

func (ma *ArrayMonitoringAdapter) NextValues(gt model.Guarantee) map[string]MetricValue {
	if ma.i == len(ma.values) {
		return nil
	}
	result := ma.values[ma.i]
	ma.i++
	return result
}
