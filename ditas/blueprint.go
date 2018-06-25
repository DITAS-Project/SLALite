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

package ditas

type TreeStructureType struct {
	Type     *string             `json:"type"`
	Children []TreeStructureType `json:"children"`
	Leaves   []string            `json:"leaves"`
}

type MetricPropertyType struct {
	Name    string   `json:"name"`
	Unit    *string  `json:"unit"`
	Minimum *float64 `json:"minimum"`
	Maximum *float64 `json:"maximum"`
	Value   *float64 `json:"value"`
}

type MetricType struct {
	ID         *string              `json:"id"`
	Type       *string              `json:"type"`
	Properties []MetricPropertyType `json:"properties"`
}

type GoalType struct {
	ID      *string      `json:"id"`
	Metrics []MetricType `json:"metrics"`
}

type ConstraintType struct {
	Goals         []GoalType         `json:"goals"`
	TreeStructure *TreeStructureType `json:"treeStructure"`
}

type ConstraintsType struct {
	DataUtility *ConstraintType `json:"dataUtility"`
}

type MethodType struct {
	Name        *string          `json:"name"`
	Constraints *ConstraintsType `json:"constraints"`
}

type MethodListType struct {
	Methods []MethodType `json:"methods"`
}

type OverviewType struct {
	Name string `json:"Name"`
}

type InternalStructureType struct {
	Overview *OverviewType `json:"Overview"`
}

type BlueprintType struct {
	InternalStructure  *InternalStructureType `json:"INTERNAL_STRUCTURE"`
	DataManagement     *MethodListType        `json:"DATA_MANAGEMENT"`
	AbstractProperties *MethodListType        `json:"ABSTRACT_PROPERTIES"`
}
