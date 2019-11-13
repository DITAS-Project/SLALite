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
	amodel "SLALite/assessment/model"
	"SLALite/model"
	"math"
	"time"
)

// all values are read-only but index and last
type mountCtx struct {
	// metric series
	values map[model.Variable][]model.MetricValue
	// last known values for each variable
	last map[string]model.MetricValue
	// index contains the index of the process on each series
	index map[model.Variable]int
	// lens contains each series length
	lens map[model.Variable]int
	// maxlen contains the maximum length
	maxlen int
	// maxdelta is maximum time allowed for metrics to be considered in the same pointset
	maxdelta float64
	// sumlens is the sum of the series lengths in values
	sumlens int
}

// Be careful if you run the code here after 73069258126-09-25!
var _INF = time.Unix(1<<61, 0)

/*
Mount builds the GuaranteeData structure, directly used for agreement assessment,
considering constant interpolation.

Constant interpolation means that for a variable whose value is not known at a time t,
it is considered that it has the value of last known value.

It receives the metric series separated by metric name and returns the metric series grouped by
point sets that occur at the same time (actually, in a delta minor than the maxdelta parameter).

A point is considered a MetricValue.
An amodel.ExpressionData is called here a point set, i.e.
the set of MetricValues needed for evaluating an agreement at some instant t.
The current point of a series is the next value to consider, and it is determined
by the series index.

The algorithm is:

	1 take the first in time of all current points
	2 build a pointset with all the points that happen in a delta less than deltamax
  		- if for a variable there is no point in delta, take the last known value for it.
	 	- if not all values of variables can have a value
		(this can happen if there are not values in lastvalues), discard the point set.
	3 goto 1 if there are more values to consider

Example (considering no last values):

	3        o-----
	2    o---      o----
	1      x----   x----
	0           x--

The point sets are: (o=2,x=1), (o=3,x=1), (o=3,x=0), (o=2,x=1). The first pointset is discarded,
as there it not known value for x.
*/
func Mount(valuesmap map[model.Variable][]model.MetricValue,
	lastvalues map[string]model.MetricValue,
	maxdelta float64) amodel.GuaranteeData {

	ctx := initCtx(valuesmap, lastvalues, maxdelta)

	result := make(amodel.GuaranteeData, 0, ctx.maxlen)

	var step int
	for step := 0; step < ctx.sumlens && !areAllValuesExhausted(ctx.index, ctx.lens); step++ {
		nextp := ctx.findNextPoint()
		pointset, ok := ctx.buildNextPointSet(nextp)

		if ok {
			result = append(result, pointset)
		}
	}
	/*
	 * The maximum number of pointsets is the sum of all series lengths.
	 * Above that, we have entered an infinite loop (probably because of wrong input)
	 */
	if step == ctx.sumlens {
		return amodel.GuaranteeData{}
	}

	return result
}

func initCtx(valuesmap map[model.Variable][]model.MetricValue,
	lastvalues map[string]model.MetricValue,
	maxdelta float64) mountCtx {

	if lastvalues == nil {
		lastvalues = model.LastValues{}
	}
	index := make(map[model.Variable]int)
	lens := make(map[model.Variable]int)
	max := 0
	sum := 0

	for v := range valuesmap {
		// fill index
		index[v] = 0

		// lens and calculate maximum length
		l := len(valuesmap[v])
		lens[v] = l
		if l > max {
			max = l
		}
		sum += l
	}
	ctx := mountCtx{
		values:   valuesmap,
		last:     lastvalues,
		index:    index,
		lens:     lens,
		maxlen:   max,
		maxdelta: maxdelta,
		sumlens:  sum,
	}
	return ctx
}

func (ctx *mountCtx) findNextPoint() model.MetricValue {
	var result model.MetricValue

	valuesmap := ctx.values
	index := ctx.index

	mint := _INF
	for v := range valuesmap {
		i := index[v]
		if i == ctx.lens[v] {
			continue
		}
		point := valuesmap[v][i]
		t := point.DateTime

		if t.Before(mint) {
			mint = t
			result = point
		}
	}
	return result
}

func (ctx *mountCtx) buildNextPointSet(nextp model.MetricValue) (amodel.ExpressionData, bool) {
	var data amodel.ExpressionData = map[string]model.MetricValue{}
	var discard = false

	for v := range ctx.values {
		value := ctx.getCurrentValue(v)
		if deltaTimes(nextp, value) <= ctx.maxdelta {
			data[v.Name] = value
			ctx.index[v]++
			ctx.last[v.Name] = value
		} else if _, ok := ctx.last[v.Name]; !ok {
			discard = true
		} else {
			data[v.Name] = ctx.last[v.Name]
		}
	}
	return data, !discard
}

/*
getCurrentValue returns the next value of a variable according to the
index.

If the series has been exhausted or is empty,
it returns a empty metric in the infinite future.
*/
func (ctx *mountCtx) getCurrentValue(v model.Variable) model.MetricValue {
	i := ctx.index[v]
	if i == ctx.lens[v] {
		return model.MetricValue{
			Key:      v.Name,
			DateTime: _INF,
		}
	}
	return ctx.values[v][i]
}

func deltaTimes(p1 model.MetricValue, p2 model.MetricValue) float64 {
	sub := p1.DateTime.Sub(p2.DateTime).Seconds()
	delta := math.Abs(sub)
	return delta
}

func areAllValuesExhausted(index map[model.Variable]int, lens map[model.Variable]int) bool {
	for v := range index {
		if index[v] != lens[v] {
			return false
		}
	}
	return true
}
