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
package simpleadapter

import "SLALite/model"
import "SLALite/assessment/monitor"

type ArrayMonitoringAdapter struct {
	agreement *model.Agreement
	values    []map[string]monitor.MetricValue
}

func New(values []map[string]monitor.MetricValue) *ArrayMonitoringAdapter {
	return &ArrayMonitoringAdapter{
		agreement: nil,
		values:    values,
	}
}

func (ma *ArrayMonitoringAdapter) Initialize(a *model.Agreement) {
	ma.agreement = a
}

func (ma *ArrayMonitoringAdapter) GetValues(gt model.Guarantee, vars []string) []map[string]monitor.MetricValue {
	return ma.values
}
