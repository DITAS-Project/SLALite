/*
Copyright 2019 Atos

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

package genericadapter

import (
	"SLALite/assessment"
	"SLALite/model"
	"SLALite/utils"
	"os"
	"testing"
	"time"
)

var Average = model.Aggregation{
	Type:   model.AVERAGE,
	Window: 0, /* don't care */
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func TestIdentity(t *testing.T) {
	name := "m1"
	t0 := time.Now()
	values := newValues(name, t0, []m{
		{0, 1}, {1, 2}, {2, 0.5}, {3, 1.5},
	})
	v := model.Variable{
		Metric: name,
	}
	Identity(v, values)
}

func TestAverage(t *testing.T) {
	name := "avg"
	t0 := time.Now()
	values := newValues(name, t0, []m{
		{0, 1}, {1, 2}, {2, 0.5}, {3, 1.5},
	})

	var avg float64
	if avg = average(values); avg != 1.25 {
		t.Errorf("Unexpected average. Expected: %f; Actual: %f", 1.25, avg)
	}
	v := model.Variable{
		Name:        name,
		Metric:      name,
		Aggregation: &Average,
	}
	output := Aggregate(v, values)
	if len(output) != 1 {
		t.Errorf("Unexpected values length. Expected: %d; Actual: %d", 1, len(output))
		return
	}
	expected := model.MetricValue{
		Key:      name,
		Value:    1.25,
		DateTime: values[len(values)-1].DateTime,
	}
	if output[0] != expected {
		t.Errorf("Unexpected average metric value. Expected: %#v; Actual: %#v", expected, output[0])
	}
}

func TestAverageWrongInput(t *testing.T) {
	name := "avg"
	t0 := time.Now()
	values := newValues(name, t0, []m{
		{0, 1}, {1, 2}, {2, 0.5}, {3, 1.5},
	})

	v := model.Variable{
		Metric:      name,
		Aggregation: nil,
	}
	testAverageWrongInput(t, v, values)

	v = model.Variable{
		Metric: name,
		Aggregation: &model.Aggregation{
			Type: "", Window: 0,
		},
	}
	testAverageWrongInput(t, v, values)

	v = model.Variable{
		Metric:      name,
		Aggregation: &Average,
	}
	testAverageWrongInput(t, v, []model.MetricValue{})

}
func testAverageWrongInput(t *testing.T, v model.Variable, values []model.MetricValue) {

	output := Aggregate(v, values)
	if len(output) != len(values) {
		t.Errorf("Unexpected values length. Expected: %d; Actual: %d", len(values), len(output))
		return
	}
}

func TestGenericAdapter(t *testing.T) {
	retriever := DummyRetriever{3}
	retrieve := retriever.Retrieve()

	ga := Adapter{
		Retrieve: retrieve,
		Process:  Aggregate,
	}
	a, _ := utils.ReadAgreement("testdata/a.json")

	ma := ga.Initialize(&a)
	assessment.EvaluateAgreement(&a, ma, time.Now())
	/*
	 * Just tests that nothing breaks
	 */
}

func newVar(name string) model.Variable {
	return model.Variable{
		Name:   name,
		Metric: name,
	}
}

type metricBuilder struct {
	name string
	T    utils.Timeline
}

func (mb metricBuilder) newValue(t float64, v float64) model.MetricValue {
	T := mb.T
	return model.MetricValue{
		Key: mb.name, Value: v, DateTime: T.T(t),
	}
}

type m struct {
	t float64
	v float64
}

func newValues(name string, t0 time.Time, ms []m) []model.MetricValue {
	result := make([]model.MetricValue, 0)
	_t := utils.Timeline{T0: t0}
	mb := metricBuilder{name, _t}
	for _, m := range ms {
		result = append(result, mb.newValue(m.t, m.v))
	}

	return result
}
