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

/*
	5    o
	4  o      o
	3    -
	2 -      -
	1    *
	0*      *
	-0123456789

The first two points have to be discarded. The pointsets are:
- t = 2 -> (*=0, -=2, o=4)
- t = 4 -> (*=1, -=3, o=5)
- t = 7 -> (*=0, -=3, o=5)
- t = 8 -> (*=0, -=2, o=5)
- t = 9 -> (*=0, -=2, o=4)
*/

import (
	amodel "SLALite/assessment/model"
	"SLALite/model"
	"fmt"
	"testing"
	"time"
)

var v1 = newVar("*")
var v2 = newVar("-")
var v3 = newVar("o")
var t0 = time.Time{}

var v1V = newValues(v1.Metric, t0, []m{
	m{0.1, 0}, m{4, 1}, m{7, 0},
})
var v2V = newValues(v2.Metric, t0, []m{
	m{1, 2}, m{3.9, 3}, m{8, 2},
})
var v3V = newValues(v3.Metric, t0, []m{
	m{2, 4}, m{4.1, 5}, m{9, 4},
})

func TestFindNextPoint(t *testing.T) {

	valuesmap := map[model.Variable][]model.MetricValue{v1: v1V, v2: v2V, v3: v3V}
	lastvalues := map[string]model.MetricValue{}
	ctx := initCtx(valuesmap, lastvalues, 0.2)

	p := ctx.findNextPoint()
	if p != v1V[0] {
		t.FailNow()
	}

	for v := range ctx.index {
		ctx.index[v] = 1
	}
	p = ctx.findNextPoint()
	if p != v2V[1] {
		t.FailNow()
	}

	for v := range ctx.index {
		ctx.index[v] = 2
	}
	p = ctx.findNextPoint()
	if p != v1V[2] {
		t.FailNow()
	}
}

func TestBuildNextPointSet(t *testing.T) {

	valuesmap := map[model.Variable][]model.MetricValue{v1: v1V, v2: v2V, v3: v3V}
	lastvalues := map[string]model.MetricValue{}
	ctx := initCtx(valuesmap, lastvalues, 0.2)

	p := ctx.findNextPoint()
	data, ok := ctx.buildNextPointSet(p)
	if ok {
		t.Errorf("1st pointset should have been discarded")
		return
	}

	p = ctx.findNextPoint()
	data, ok = ctx.buildNextPointSet(p)
	if ok {
		t.Errorf("2nd pointset should have been discarded")
		return
	}

	p = ctx.findNextPoint()
	data, ok = ctx.buildNextPointSet(p)
	if !ok {
		t.Errorf("3rd pointset should not have been discarded")
		return
	}
	if !assertPointSet(t, data, v1V[0], v2V[0], v3V[0]) {
		t.Errorf("Does not match")
		return
	}

	p = ctx.findNextPoint()
	data, ok = ctx.buildNextPointSet(p)
	if !ok {
		t.Errorf("4rd pointset should not have been discarded")
		return
	}
	if !assertPointSet(t, data, v1V[1], v2V[1], v3V[1]) {
		t.Errorf("Does not match")
		return
	}
}

func TestMount(t *testing.T) {
	valuesmap := map[model.Variable][]model.MetricValue{v1: v1V, v2: v2V, v3: v3V}
	lastvalues := map[string]model.MetricValue{}

	pointsets := Mount(valuesmap, lastvalues, 0.2)

	var pointset amodel.ExpressionData
	if len(pointsets) != 5 {
		t.Errorf("Unexpected number of pointsets. Expected: %d; Actual: %d", 5, len(pointsets))
		return
	}

	pointset = pointsets[0]
	assertPointSet(t, pointset, v1V[0], v2V[0], v3V[0])

	pointset = pointsets[1]
	assertPointSet(t, pointset, v1V[1], v2V[1], v3V[1])

	pointset = pointsets[2]
	assertPointSet(t, pointset, v1V[2], v2V[1], v3V[1])

	pointset = pointsets[3]
	assertPointSet(t, pointset, v1V[2], v2V[2], v3V[1])

	pointset = pointsets[4]
	assertPointSet(t, pointset, v1V[2], v2V[2], v3V[2])

}

func TestEmptySeries(t *testing.T) {
	empty := newValues(v3.Metric, t0, []m{})
	valuesmap := map[model.Variable][]model.MetricValue{v1: v1V, v2: v2V, v3: empty}
	lastvalues := map[string]model.MetricValue{}

	pointsets := Mount(valuesmap, lastvalues, 0.2)

	if len(pointsets) != 0 {
		t.Errorf("Unexpected number of pointsets. Expected: %d; Actual: %d", 0, len(pointsets))
		return
	}

}

func TestUnsortedSeries(t *testing.T) {
	unsorted := newValues(v3.Metric, t0, []m{
		m{9, 4}, m{4.1, 5}, m{2, 4},
	})
	valuesmap := map[model.Variable][]model.MetricValue{v1: v1V, v2: v2V, v3: unsorted}
	lastvalues := map[string]model.MetricValue{}

	/* Just test that Mount finishes, althought results are umpredictable */
	pointsets := Mount(valuesmap, lastvalues, 0.2)
	fmt.Printf("%#v", pointsets)
}

func assertPointSet(t *testing.T, data amodel.ExpressionData, m1, m2, m3 model.MetricValue) bool {
	if data[m1.Key] != m1 || data[m2.Key] != m2 || data[m3.Key] != m3 {
		if data[m1.Key] != m1 {
			t.Errorf("Mismatch data[%s]=%v, m1=%v", m1.Key, data[m1.Key], m1)
		}
		if data[m2.Key] != m1 {
			t.Errorf("Mismatch data[%s]=%v, m2=%v", m2.Key, data[m2.Key], m2)
		}
		if data[m3.Key] != m3 {
			t.Errorf("Mismatch data[%s]=%v, m3=%v", m3.Key, data[m3.Key], m3)
		}
		return false
	}
	return true
}
