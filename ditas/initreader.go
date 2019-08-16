/**
 * Copyright 2018 Atos
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not
 * use this file except in compliance with the License. You may obtain a copy of
 * the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
 * WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
 * License for the specific language governing permissions and limitations under
 * the License.
 *
 * This is being developed for the DITAS Project: https://www.ditas-project.eu/
 */

package ditas

import (
	"SLALite/assessment/monitor"
	"SLALite/assessment/monitor/genericadapter"
	"SLALite/assessment/notifier"
	"SLALite/model"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/DITAS-Project/blueprint-go"
	"github.com/go-resty/resty/v2"
	"github.com/spf13/viper"

	log "github.com/sirupsen/logrus"
)

const (
	// BlueprintLocation is the location where the DITAS blueprint must be found
	BlueprintLocation = "/etc/ditas"

	// BlueprintName is the name of the DITAS blueprint file to read to compose SLAs
	BlueprintName = "blueprint.json"

	ConfigFileName = "slalite"

	// BlueprintPath is the path to the DITAS blueprint
	BlueprintPath = BlueprintLocation + "/" + BlueprintName

	DS4MUrlDefault = "http://vdm:8080"

	VDCIdPropery             = "vdcId"
	DataAnalyticsUrlProperty = "data.analytics.url"
	DS4MUrlProperty          = "ds4m.url"
)

type MethodInfo struct {
	MethodID  string
	Agreement model.Agreement
	Path      string
	Operation string
}

func readProperty(property blueprint.MetricPropertyType, name string) string {
	if property.Value != nil {
		return fmt.Sprintf("%s == %f", name, *property.Value)
	}

	if property.Maximum != nil && property.Minimum != nil {
		return fmt.Sprintf("%s <= %f && %s >= %f", name, *property.Maximum, name, *property.Minimum)
	}

	if property.Maximum != nil && property.Minimum == nil {
		return fmt.Sprintf("%s <= %f", name, *property.Maximum)
	}

	if property.Minimum != nil && property.Maximum == nil {
		return fmt.Sprintf("%s >= %f", name, *property.Minimum)
	}
	return ""
}

func readProperties(properties map[string]blueprint.MetricPropertyType) string {
	var result strings.Builder
	i := 0
	if properties != nil {
		for name, property := range properties {
			result.WriteString(readProperty(property, name))
			if i < len(properties)-1 {
				result.WriteString(" && ")
			}
			i++
		}
	}
	return result.String()
}

/*func getExpression(goal blueprint.GoalType) string {
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
}*/

func getExpressions(goals []blueprint.ConstraintType) map[string]string {
	result := make(map[string]string)
	for _, goal := range goals {
		if goal.ID != nil {
			result[*goal.ID] = readProperties(goal.Properties)
		} else {
			log.Errorf("Can't parse goal since it doesn't have a valid ID")
		}
	}
	return result
}

func composeExpression(ids []string, expressions map[string]string) string {
	var result strings.Builder
	for i, id := range ids {
		expression, ok := expressions[id]
		if ok {
			result.WriteString(expression)
			if i < len(ids)-1 {
				result.WriteString(" && ")
			}
		} else {
			log.Errorf("Invalid blueprint. Found attribute id %s not found in constraint list", id)
		}
	}
	return result.String()
}

func createGuarantee(leaf blueprint.LeafType, expressions map[string]string) model.Guarantee {
	return model.Guarantee{Name: *leaf.Id, Constraint: composeExpression(leaf.Attributes, expressions)}
}

func flattenLeaves(leaves []blueprint.LeafType, expressions map[string]string, operator string) (string, string) {
	var result strings.Builder
	var name strings.Builder
	for i, leaf := range leaves {
		name.WriteString(*leaf.Id)
		result.WriteString("(")
		result.WriteString(composeExpression(leaf.Attributes, expressions))
		result.WriteString(")")
		if i < len(leaves)-1 {
			result.WriteString(" ")
			result.WriteString(operator)
			result.WriteString(" ")

			name.WriteString(" ")
			name.WriteString(operator)
			name.WriteString(" ")
		}
	}
	return name.String(), result.String()
}

func flatten(tree blueprint.TreeStructureType, expressions map[string]string) (string, string) {
	operator := "||"
	if *tree.Type == "AND" {
		operator = "&&"
	}
	name, constraint := flattenLeaves(tree.Leaves, expressions, operator)
	var nameBuilder strings.Builder
	var constraintBuilder strings.Builder

	nameBuilder.WriteString(name)
	constraintBuilder.WriteString(constraint)

	for _, child := range tree.Children {
		if constraint != "" {
			constraintBuilder.WriteString(" ")
			constraintBuilder.WriteString(operator)
			constraintBuilder.WriteString(" ")

			nameBuilder.WriteString(" ")
			nameBuilder.WriteString(operator)
			nameBuilder.WriteString(" ")
		}
		partialName, partialConstraint := flatten(child, expressions)
		nameBuilder.WriteString(partialName)
		constraintBuilder.WriteString("(")
		constraintBuilder.WriteString(partialConstraint)
		constraintBuilder.WriteString(")")
	}
	return nameBuilder.String(), constraintBuilder.String()
}

func parseTree(tree blueprint.TreeStructureType, expressions map[string]string) []model.Guarantee {
	switch *tree.Type {
	case "AND":
		init := make([]model.Guarantee, 0, len(tree.Leaves)+len(tree.Children))
		for _, leaf := range tree.Leaves {
			init = append(init, createGuarantee(leaf, expressions))
		}
		for _, child := range tree.Children {
			init = append(init, parseTree(child, expressions)...)
		}
		return init
	case "OR":
		name, constraint := flatten(tree, expressions)
		return []model.Guarantee{model.Guarantee{Name: name, Constraint: constraint}}
	}

	return make([]model.Guarantee, 0)
}

func getGuarantees(method blueprint.AbstractPropertiesMethodType, expressions map[string]string) []model.Guarantee {
	return parseTree(method.GoalTrees.DataUtility, expressions)
}

func CreateAgreements(bp *blueprint.BlueprintType) (model.Agreements, map[string]blueprint.ExtendedOps) {
	blueprintName := bp.InternalStructure.Overview.Name

	methodInfo := blueprint.AssembleOperationsMap(*bp)

	agreements := make(map[string]*model.Agreement)
	expressions := make(map[string]map[string]string)

	methods := bp.DataManagement
	if methods != nil && len(methods) > 0 {
		for _, method := range methods {
			if method.MethodId != nil {
				agreement := model.Agreement{
					Id:   *method.MethodId,
					Name: *method.MethodId,
					Details: model.Details{
						Name: *method.MethodId,
						Provider: model.Provider{
							Id:   *blueprintName,
							Name: *blueprintName,
						},
						Client: model.Client{
							Id:   *blueprintName,
							Name: *blueprintName,
						},
						Id: *method.MethodId,
					},
					State: model.STARTED,
				}
				agreement.Id = *method.MethodId

				if method.Attributes.DataUtility != nil {
					expressions[*method.MethodId] = getExpressions(method.Attributes.DataUtility)
				}
				agreements[agreement.Id] = &agreement
			} else {
				log.Errorf("INVALID BLUEPRINT %s: Found method without name", *blueprintName)
			}
		}

		absMethods := bp.AbstractProperties
		if absMethods != nil {
			for _, method := range absMethods {
				exp, foundExp := expressions[*method.MethodId]
				agreement, foundAg := agreements[*method.MethodId]
				if foundExp && foundAg {
					agreement.Details.Guarantees = getGuarantees(method, exp)
				} else {
					log.Errorf("INVALID BLUEPRINT %s: Method %s goals or tree not found", *blueprintName, *method.MethodId)
				}
			}
		} else {
			log.Errorf("INVALID BLUEPRINT %s: Abstract properties section not found", *blueprintName)
		}
	} else {
		log.Errorf("INVALID BLUEPRINT %s: Can't find any method in data management section", *blueprintName)
	}

	var results = make(model.Agreements, 0)
	for _, value := range agreements {
		if value.Details.Guarantees != nil && len(value.Details.Guarantees) > 0 {
			results = append(results, *value)
		}
	}
	return results, methodInfo
}

func sendBlueprintToVDM(logger *log.Entry, ds4mUrl string) error {
	rawJSON, err := ioutil.ReadFile(BlueprintPath)
	if err != nil {
		logger.WithError(err).Error("Error reading")
		return err
	}
	rawJSONStr := string(rawJSON)
	data := map[string]string{
		"ConcreteBlueprint": rawJSONStr,
	}
	_, err = resty.New().R().SetFormData(data).Post(ds4mUrl + "/AddVDC")
	if err != nil {
		logger.WithError(err).Error("Error received from DS4M service")
		return err
	}
	return nil
}

func Configure(repo model.IRepository) (monitor.MonitoringAdapter, notifier.ViolationNotifier, error) {
	config := viper.New()

	config.SetDefault(DS4MUrlProperty, DS4MUrlDefault)

	config.AddConfigPath(BlueprintLocation)
	config.SetConfigName(ConfigFileName)
	config.ReadInConfig()

	bp, err := blueprint.ReadBlueprint(BlueprintPath)
	if err != nil {
		log.WithError(err).Error("Error reading blueprint")
		return nil, nil, err
	}

	logger := log.WithField("blueprint", *bp.InternalStructure.Overview.Name)

	logger.Debug("Creating blueptint at VDM")

	err = sendBlueprintToVDM(logger, config.GetString(DS4MUrlProperty))

	if err != nil {
		logger.WithError(err).Error("Error registering blueprint in VDM. Violation notification will have problems")
	}

	agreements, _ := CreateAgreements(bp)

	if agreements != nil {
		for _, agreement := range agreements {
			_, err := repo.CreateAgreement(&agreement)
			if err != nil {
				log.Errorf("Error creating agreement %s: %s", agreement.Id, err.Error())
			}
		}
	}
	da := NewDataAnalyticsAdapter(viper.GetString(DataAnalyticsUrlProperty), viper.GetString(VDCIdPropery))
	adapter := genericadapter.New(da.Retrieve, da.Process)
	return adapter, NewNotifier(config.GetString(VDCIdPropery), config.GetString(DS4MUrlProperty)), nil
}
