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

import (
	"SLALite/model"
	"encoding/json"
	"fmt"
	"io/ioutil"

	log "github.com/sirupsen/logrus"
)

func ReadBlueprint(path string) BlueprintType {
	raw, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Println(err.Error())
		log.Errorf("Error reading blueprint from %s: %s", path, err.Error())
	}

	var blueprint BlueprintType
	err = json.Unmarshal(raw, &blueprint)
	if err != nil {
		log.Errorf("Error reading blueprint: %s", err.Error())
	}
	return blueprint
}

func readProperty(property MetricPropertyType) string {
	if property.Value != nil {
		return fmt.Sprintf("%s = %f", property.Name, *property.Value)
	}

	if property.Maximum != nil && property.Minimum != nil {
		return fmt.Sprintf("%s <= %f && %s >= %f", property.Name, *property.Maximum, property.Name, *property.Minimum)
	}

	if property.Maximum != nil && property.Minimum == nil {
		return fmt.Sprintf("%s <= %f", property.Name, *property.Maximum)
	}

	if property.Minimum != nil && property.Maximum == nil {
		return fmt.Sprintf("%s >= %f", property.Name, *property.Minimum)
	}
	return ""
}

func readProperties(properties []MetricPropertyType) string {
	var result string
	if properties != nil {
		//This is repeated but so far Go doesn't support generics.
		//If it does some day and I'm not maintaining this code, please change this to something generic.
		for i, property := range properties {
			result = result + readProperty(property)
			if i < len(properties)-1 {
				result = result + " && "
			}
		}
	}
	return result
}

func getExpression(goal GoalType) string {
	var result string
	if goal.Metrics != nil {
		for i, metric := range goal.Metrics {
			result = result + readProperties(metric.Properties)
			if i < len(goal.Metrics)-1 {
				result = result + " %% "
			}
		}
	}
	return result
}

func getExpressions(goals []GoalType) map[string]string {
	result := make(map[string]string)
	for _, goal := range goals {
		if goal.ID != nil {
			result[*goal.ID] = getExpression(goal)
		} else {
			log.Errorf("Can't parse goal since it doesn't have a valid ID")
		}
	}
	return result
}

func createGuarantee(name string, expression string) model.Guarantee {
	return model.Guarantee{Name: name, Constraint: expression}
}

func flatten(children []TreeStructureType, expressions map[string]string, operator string) string {
	constraint := ""
	for _, child := range children {
		guarantees := parseTree(&child, expressions)
		for j, guarantee := range guarantees {
			constraint = constraint + "(" + guarantee.Constraint + ")"
			if j < len(guarantees)-1 {
				constraint = constraint + " " + operator + " "
			}
		}
	}
	return constraint
}

func parseTree(tree *TreeStructureType, expressions map[string]string) []model.Guarantee {
	if tree.Leaves != nil && len(tree.Leaves) > 0 {
		switch *tree.Type {
		case "AND":
			if len(tree.Leaves) == 2 {
				return []model.Guarantee{
					createGuarantee(tree.Leaves[0], expressions[tree.Leaves[0]]),
					createGuarantee(tree.Leaves[1], expressions[tree.Leaves[1]]),
				}
			}
			init := make([]model.Guarantee, 1)
			init[0] = createGuarantee(tree.Leaves[0], expressions[tree.Leaves[0]])
			for _, child := range tree.Children {
				init = append(init, parseTree(&child, expressions)...)
			}
			return init
		case "OR":
			if len(tree.Leaves) == 2 {
				name := tree.Leaves[0] + " or " + tree.Leaves[1]
				constraint := expressions[tree.Leaves[0]] + " || " + expressions[tree.Leaves[1]]
				return []model.Guarantee{createGuarantee(name, constraint)}
			}
			constraint := expressions[tree.Leaves[0]] + " || (" + flatten(tree.Children, expressions, "AND") + ")"
			return []model.Guarantee{createGuarantee(tree.Leaves[0]+"_complex", constraint)}
		}
	} else {
		switch *tree.Type {
		case "AND":
			result := make([]model.Guarantee, 0)
			for _, child := range tree.Children {
				result = append(result, parseTree(&child, expressions)...)
			}
			return result
		case "OR":
			constraint := flatten(tree.Children, expressions, "||")
			return []model.Guarantee{createGuarantee("complex", constraint)}
		}
	}
	return make([]model.Guarantee, 0)
}

func getGuarantees(method MethodType, expressions map[string]string) []model.Guarantee {
	if method.Constraints != nil && method.Constraints.DataUtility != nil && method.Constraints.DataUtility.TreeStructure != nil {
		return parseTree(method.Constraints.DataUtility.TreeStructure, expressions)
	}
	log.Errorf("Can't find tree structure for method %s", method.Name)

	return make([]model.Guarantee, 0)
}

func CreateAgreements(blueprint BlueprintType) []model.Agreement {
	blueprintName := blueprint.InternalStructure.Overview.Name

	agreements := make(map[string]*model.Agreement)
	expressions := make(map[string]map[string]string)

	dataManagement := blueprint.DataManagement
	if dataManagement != nil {
		methods := dataManagement.Methods
		if methods != nil && len(methods) > 0 {
			for _, method := range methods {
				var agreement model.Agreement
				agreement.Id = *method.Name
				if method.Constraints != nil && method.Constraints.DataUtility != nil && method.Constraints.DataUtility.Goals != nil {
					expressions[*method.Name] = getExpressions(method.Constraints.DataUtility.Goals)
				}
				agreements[agreement.Id] = &agreement
			}
		} else {
			log.Errorf("INVALID BLUEPRINT %s: Can't find any method in data management section", blueprintName)
		}
	} else {
		log.Errorf("INVALID BLUEPRINT %s: Data Management section not found")
	}

	if blueprint.AbstractProperties != nil && blueprint.AbstractProperties.Methods != nil {
		for _, method := range blueprint.AbstractProperties.Methods {
			exp, foundExp := expressions[*method.Name]
			agreement, foundAg := agreements[*method.Name]
			if foundExp && foundAg {
				agreement.Details.Guarantees = getGuarantees(method, exp)
			} else {
				log.Error("INVALID BLUEPRINT %s: Method %s goals or tree not found", blueprintName)
			}
		}
	} else {
		log.Errorf("INVALID BLUEPRINT %s: Abstract properties section not found", blueprintName)
	}

	var results = make([]model.Agreement, len(agreements))
	i := 0
	for _, value := range agreements {
		results[i] = *value
		i++
	}
	return results
}
