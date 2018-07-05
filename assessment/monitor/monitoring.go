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
package monitor

import (
	"SLALite/model"
	"fmt"
	"time"
)

// MetricValue is the SLALite representation of a metric value.
type MetricValue struct {
	Key      string
	Value    interface{}
	DateTime time.Time
}

func (v *MetricValue) String() string {
	return fmt.Sprintf("{Key: %s, Value: %v, DateTime: %v}", v.Key, v.Value, v.DateTime)
}

//MonitoringAdapter is an interface which should be implemented per monitoring solution
type MonitoringAdapter interface {
	Initialize(a *model.Agreement)
	GetValues(gt model.Guarantee, vars []string) []map[string]MetricValue
}
