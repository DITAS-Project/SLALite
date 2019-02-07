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

package blueprint

import (
	"github.com/go-openapi/spec"
)

type LeafType struct {
	Id          *string  `json:"id"`
	Description string   `json:"description"`
	Weight      int      `json:"weight"`
	Attributes  []string `json:"attributes"`
}

type TreeStructureType struct {
	Type     *string             `json:"type"`
	Children []TreeStructureType `json:"children"`
	Leaves   []LeafType          `json:"leaves"`
}

type GoalTreeType struct {
	DataUtility TreeStructureType `json:"dataUtility`
	Security    TreeStructureType `json:"security`
	Privacy     TreeStructureType `json:"privacy`
}

type AbstractPropertiesMethodType struct {
	MethodId  *string      `json:"method_id"`
	GoalTrees GoalTreeType `json:"goalTrees"`
}

type MetricPropertyType struct {
	Unit    string       `json:"unit"`
	Minimum *float64     `json:"minimum"`
	Maximum *float64     `json:"maximum"`
	Value   *interface{} `json:"value"`
}

type ConstraintType struct {
	ID          *string                       `json:"id"`
	Description string                        `json:"description"`
	Type        string                        `json:"type"`
	Properties  map[string]MetricPropertyType `json:"properties"`
}

type DataManagementAttributesType struct {
	DataUtility []ConstraintType `json:"dataUtility"`
	Security    []ConstraintType `json:"security"`
	Privacy     []ConstraintType `json:"privacy"`
}

type DataManagementMethodType struct {
	MethodId   *string                      `json:"method_id"`
	Attributes DataManagementAttributesType `json:"attributes"`
}
type MethodTagType struct {
	ID   string   `json:"method_id"`
	Tags []string `json:"tags"`
}
type OverviewType struct {
	Name *string         `json:"Name"`
	Tags []MethodTagType `json:"tags"`
}

type DataSourceType struct {
	ID         *string                `json:"id"`
	Type       *string                `json:"type"`
	Parameters map[string]interface{} `json:"parameters"`
}

type InternalStructureType struct {
	Overview    OverviewType     `json:"Overview"`
	DataSources []DataSourceType `json:"Data_Sources"`
}

type BlueprintType struct {
	InternalStructure  InternalStructureType          `json:"INTERNAL_STRUCTURE"`
	DataManagement     []DataManagementMethodType     `json:"DATA_MANAGEMENT"`
	AbstractProperties []AbstractPropertiesMethodType `json:"ABSTRACT_PROPERTIES"`
	API                spec.Swagger                   `json:"EXPOSED_API"`
}
