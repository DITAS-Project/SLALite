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
	ma.Initialize(&agreement)
	for _, gt := range gts {
		for values := ma.NextValues(gt); values != nil; values = ma.NextValues(gt) {
			...
		}
	}
*/
package dummyadapter

import (
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

func (ma *monitoringAdapter) Initialize(a *model.Agreement) {
	ma.agreement = a
}

func (ma *monitoringAdapter) GetValues(gt model.Guarantee, vars []string) []map[string]monitor.MetricValue {
	result := make([]map[string]monitor.MetricValue, ma.size)
	for i := 0; i < ma.size; i++ {
		val := make(map[string]monitor.MetricValue)

		for _, key := range vars {
			val[key] = monitor.MetricValue{
				DateTime: time.Now(),
				Key:      key,
				Value:    rand.Float64(),
			}
		}

		result[i] = val
	}
	return result
}
