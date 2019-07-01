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

/*
Package genericadapter provides a configurable MonitoringAdapter
that works with advanced agreement schema.

Usage:
	ma := genericadapter.New(retriever, processor)
	ma = ma.Initialize(&agreement)
	for _, gt := range gts {
		for values := range ma.GetValues(gt, ...) {
			...
		}
	}
*/
package genericadapter

import (
	amodel "SLALite/assessment/model"
	"SLALite/assessment/monitor"
	"SLALite/model"
	"math/rand"
	"time"
)

/*
GenericAdapter is the type of a customizable adapter.

The Retrieve field is a function to query data to monitoring;
the Process field is a function to perform additional processing
on data.

Two Process functions are provided in the package:
Identity (returns the input) and Aggregation (aggregates values according
to the aggregation type)
*/
type GenericAdapter struct {
	Retrieve  Retrieve
	Process   Process
	agreement *model.Agreement
}

// Retrieve is the type of the function that makes the actual request to monitoring.
//
// It receives the list of variables to be able to retrieve all of them at once if possible.
type Retrieve func(agreement model.Agreement,
	v []model.Variable,
	from, to time.Time) map[model.Variable][]model.MetricValue

// Process is the type of the function that performs additional custom processing on
// retrieved data.
type Process func(v model.Variable, values []model.MetricValue) []model.MetricValue

/*
Initialize implements MonitoringAdapter.Initialize().

Usage:

	ga := GenericAdapter{
		Retrieve: randomRetrieve,
		Process: Aggregation,
	}
	ga := ga.Initialize(agreement)
	for _, gt := range gts {
		for values := range ga.GetValues(gt, ...) {
			...
		}
	}

*/
func (ga *GenericAdapter) Initialize(a *model.Agreement) monitor.MonitoringAdapter {
	result := *ga
	result.agreement = a
	return &result
}

// GetValues implements Monitoring.GetValues().
func (ga *GenericAdapter) GetValues(gt model.Guarantee,
	varnames []string) amodel.GuaranteeData {

	a := ga.agreement
	now := time.Now()

	var from time.Time
	if a.Assessment.LastExecution.IsZero() {
		from = a.Details.Creation
	} else {
		from = a.Assessment.LastExecution
	}

	vars := buildVarsFromVarnames(a, varnames)

	unprocessed := ga.Retrieve(*a, vars, from, now)

	/* process each of the series*/
	valuesmap := map[model.Variable][]model.MetricValue{}
	for v := range unprocessed {
		valuesmap[v] = ga.Process(v, unprocessed[v])
	}
	result := Mount(valuesmap, map[string]model.MetricValue{}, 0.1)
	return result
}

/*
GetFromForVariable returns the interval start for the query to monitoring.

If the variable is aggregated, it depends on the aggregation window.
If not, returns defaultFrom (which should be the last time the guarantee term
was evaluated)
*/
func GetFromForVariable(v model.Variable, defaultFrom, to time.Time) time.Time {
	if v.Aggregation != nil && v.Aggregation.Window != 0 {
		return to.Add(-time.Duration(v.Aggregation.Window) * time.Second)
	}
	return defaultFrom
}

func buildVarsFromVarnames(a *model.Agreement, names []string) []model.Variable {
	vars := make([]model.Variable, 0, len(names))
	for _, name := range names {
		v, _ := a.Details.GetVariable(name)

		vars = append(vars, v)
	}
	return vars
}

// Retriever is a simple struct that generates a RetrieveFunction that works similar
// to the DummyAdapter.
type Retriever struct {
	// Size is the number of values that the retrieval returns per metric
	Size int
}

// RetrieveFunction returns a Retrieve function.
func (r *Retriever) RetrieveFunction() Retrieve {

	return func(agreement model.Agreement,
		vars []model.Variable,
		from, to time.Time) map[model.Variable][]model.MetricValue {

		result := map[model.Variable][]model.MetricValue{}
		for _, v := range vars {
			result[v] = make([]model.MetricValue, 0, r.Size)
			actualFrom := GetFromForVariable(v, from, to)
			step := time.Duration(int(to.Sub(actualFrom)) / (r.Size + 1))

			for i := 0; i < r.Size; i++ {
				m := model.MetricValue{
					Key:      v.Name,
					Value:    rand.Float64(),
					DateTime: actualFrom.Add(step * time.Duration(i+1)),
				}
				result[v] = append(result[v], m)
			}
		}
		return result
	}
}

// Identity returns the input
func Identity(v model.Variable, values []model.MetricValue) []model.MetricValue {
	return values
}

// Aggregate performs an aggregation function on the input.
//
// This expects that all the values are in the appropriate window. For that,
// the Retrieve function needs to return only the values in the window. If not,
// this function will return an invalid result.
func Aggregate(v model.Variable, values []model.MetricValue) []model.MetricValue {
	if len(values) == 0 || v.Aggregation == nil || v.Aggregation.Type == "" {
		return values
	}
	if v.Aggregation.Type == model.AVERAGE {
		avg := average(values)
		return []model.MetricValue{
			model.MetricValue{
				Key:      v.Name,
				Value:    avg,
				DateTime: values[len(values)-1].DateTime,
			},
		}
	}
	/* fallback */
	return values
}

func average(values []model.MetricValue) float64 {
	sum := 0.0
	for _, value := range values {
		sum += value.Value.(float64)
	}
	result := sum / float64(len(values))

	return result
}
