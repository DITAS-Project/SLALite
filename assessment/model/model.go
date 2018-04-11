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
package model

import (
	"SLALite/assessment/monitor"
	"SLALite/model"
)

// ExpressionData represents the set of values needed to evaluate an expression at a single time
type ExpressionData map[string]monitor.MetricValue

// GuaranteeData represents the list of values needed to evaluate an expression at several points
// in time
type GuaranteeData []ExpressionData

// EvaluationGtResult is the result of the evaluation of a guarantee term
type EvaluationGtResult struct {
	Metrics    GuaranteeData     // violated metrics
	Violations []model.Violation // violations occurred as of violated metrics
}

// Result is the result of the agreement assessment
type Result map[string]EvaluationGtResult
