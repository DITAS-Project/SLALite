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
	"errors"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"time"

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

	// ConfigFileName is the name of the configuration file to read
	ConfigFileName = "slalite"

	// BlueprintPath is the path to the DITAS blueprint
	BlueprintPath = BlueprintLocation + "/" + BlueprintName

	// VDCIdPropery is the name of the property holding the VDC Id value in the configuration file
	VDCIdPropery = "vdcId"

	// InfrastructureIDProperty is the name of the property holding the infrastructure identifier in which this instance of the SLA manager is running
	InfrastructureIDProperty = "infrastructureId"

	// DataAnalyticsURLProperty is the name of the property holding the URL to the data analytics REST service
	DataAnalyticsURLProperty = "data_analytics.url"

	// DS4MPortProperty is the name of the property holding the port of the DS4M in the VDM service
	DS4MPortProperty = "ds4m.port"

	TestingEnabledProperty       = "testing.enabled"
	TestingMethodIDProperty      = "testing.method"
	TestingNumViolationsProperty = "testing.num_violations"
	TestingMetricsProperty       = "testing.metrics"

	TestingMetricNameKey  = "name"
	TestingMetricValueKey = "value"

	DebugHTTPCallsProperty = "debug.trace_http"

	// DS4MDefaultPortValue is the default port in which the DS4M listens at the VDM
	DS4MDefaultPortValue = 30003

	DS4MVDCIDHeaderName = "VDCID"

	TestingEnabledDefaultValue       = false
	TestingNumViolationsDefaultValue = 1

	VDMRetryTimeoutProperty     = "ds4m.timeout"
	VDMRetryTimeoutDefaultValue = 60

	DebugHTTPCallsDefaultValue = false
)

type methodInfo struct {
	MethodID  string
	Agreement model.Agreement
	Path      string
	Operation string
}

type constraintExpression struct {
	Expression string
	Variables  []string
}

// readProperty composes a rule based on the property value
// If there's a maximum and minimum it will create an expression x € [min, max]
// If there's just a maximum or minimum it will create x >= min or x <= max
// If it's a fixed value it will create x == value
func readProperty(property blueprint.Property, name string) string {
	if property.Value != nil {
		return fmt.Sprintf("%s == %f", name, property.Value)
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

// readProperties forms a constraint in the form metric1 && metric2 &&...
// for every property in DATA_MANAGEMENT.method.dataUtility.properties,
// i.e if there are availability and responseTime properties it will form an expression of type
// availability >= 90 && responseTime <= 1
//
// It will also return a list of variables associated to this constraint
func readProperties(properties map[string]blueprint.Property) constraintExpression {
	var result strings.Builder
	vars := make([]string, len(properties))
	i := 0
	if properties != nil {
		for name, property := range properties {
			result.WriteString(readProperty(property, name))
			vars[i] = name
			if i < len(properties)-1 {
				result.WriteString(" && ")
			}
			i++
		}
	}
	return constraintExpression{
		Expression: result.String(),
		Variables:  vars,
	}
}

// getExpressions associates to every expression in DATA_MANAGEMENT.method.dataUtility an expression
// which ANDs of all its properties and index it by the rule id
func getExpressions(goals []blueprint.DataUtility) map[string]constraintExpression {
	result := make(map[string]constraintExpression)
	for _, goal := range goals {
		if goal.ID != nil {
			result[*goal.ID] = readProperties(goal.Properties)
		} else {
			log.Errorf("Can't parse goal since it doesn't have a valid ID")
		}
	}
	return result
}

// composeExpression will create an AND expression composing all attributes of a leaf node type
// it receives the leaves as ids and the partial expressions indexed by id
// i.e in a tree of
// leaf.attibutes = [id1, id2] and expresions = {id1: "responseTime <= 2"; id2: "availability <= 90"}
// it will compose the expression "responseTime <= 2 && availability <= 90"
func composeExpression(ids []string, expressions map[string]constraintExpression) string {
	var result strings.Builder
	for i, id := range ids {
		expression, ok := expressions[id]
		if ok {
			result.WriteString(expression.Expression)
			if i < len(ids)-1 {
				result.WriteString(" && ")
			}
		} else {
			log.Errorf("Invalid blueprint. Found attribute id %s not found in constraint list", id)
		}
	}
	return result.String()
}

// createGuarantee creates a guarantee from a tree leaf assigning it the leaf id as the guarantee id.
// It will do so by creating and AND expression of all its attributes.
func createGuarantee(leaf blueprint.LeafType, expressions map[string]constraintExpression) model.Guarantee {
	return model.Guarantee{Name: *leaf.Id, Constraint: composeExpression(leaf.Attributes, expressions)}
}

// flattenLeaves will create and expression applying the passed operator to all its leafs
func flattenLeaves(leaves []blueprint.LeafType, expressions map[string]constraintExpression, operator string) (string, string) {
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

// flatten traverses recursively the tree creating a flat expression of the node operator and its leafs
func flatten(tree blueprint.TreeStructureType, expressions map[string]constraintExpression) (string, string) {
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

// flatten will create guarantees by analyzing the tree structure recursively
// As long as it finds an AND node, it will create one guarantee for each of its leaves
// and recursively create guarantees for its children
// When it finds an OR node, it will create a flat || rule with all its children and leafs
func parseTree(tree blueprint.TreeStructureType, expressions map[string]constraintExpression) []model.Guarantee {
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

// getGuarantees parses the goal tree and the expression data to form guarantees for the SLA
func getGuarantees(method blueprint.AbstractPropertiesMethodType, expressions map[string]constraintExpression) []model.Guarantee {
	return parseTree(method.GoalTrees.DataUtility, expressions)
}

// CreateAgreements creates one SLA per method found in the blueprint by:
// 1. Getting the individual constraints defined in DATA_MANAGEMENT[method_id].attributes.dataUtility
// 2. Creating guarantees with the goal tree defined in ABSTRACT_PROPERTIES[method_id].goalTrees.dataUtility
func CreateAgreements(bp *blueprint.Blueprint) (model.Agreements, map[string]blueprint.ExtendedOps) {
	blueprintID := bp.ID
	blueprintName := bp.InternalStructure.Overview.Name

	methodInfo := blueprint.AssembleOperationsMap(*bp)

	agreements := make(map[string]*model.Agreement)
	expressions := make(map[string]map[string]constraintExpression)

	methods := bp.DataManagement
	if methods != nil && len(methods) > 0 {
		for _, method := range methods {
			if method.MethodID != "" {
				agreement := model.Agreement{
					Id:   method.MethodID,
					Name: method.MethodID,
					Details: model.Details{
						Name: method.MethodID,
						Provider: model.Provider{
							Id:   blueprintID,
							Name: blueprintName,
						},
						Client: model.Client{
							Id:   blueprintID,
							Name: blueprintName,
						},
						Id:        method.MethodID,
						Variables: make([]model.Variable, 0),
					},
					State: model.STARTED,
				}
				agreement.Id = method.MethodID

				if method.Attributes.DataUtility != nil {
					attExpressions := getExpressions(method.Attributes.DataUtility)

					for _, exp := range attExpressions {
						for _, variable := range exp.Variables {
							agreement.Details.Variables = append(agreement.Details.Variables, model.Variable{
								Name:   variable,
								Metric: variable,
							})
						}
					}
					expressions[method.MethodID] = attExpressions
				}

				agreements[agreement.Id] = &agreement
			} else {
				log.Errorf("INVALID BLUEPRINT %s: Found method without name", blueprintName)
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
					log.Errorf("INVALID BLUEPRINT %s: Method %s goals or tree not found", blueprintName, *method.MethodId)
				}
			}
		} else {
			log.Errorf("INVALID BLUEPRINT %s: Abstract properties section not found", blueprintName)
		}
	} else {
		log.Errorf("INVALID BLUEPRINT %s: Can't find any method in data management section", blueprintName)
	}

	var results = make(model.Agreements, 0)
	for _, value := range agreements {
		if value.Details.Guarantees != nil && len(value.Details.Guarantees) > 0 {
			results = append(results, *value)
		}
	}
	return results, methodInfo
}

func sendBlueprintToVDM(logger *log.Entry, ds4mURL, vdcID string, timeout int64, debugHTTP bool) error {
	rawJSON, err := ioutil.ReadFile(BlueprintPath)
	if err != nil {
		logger.WithError(err).Error("Error reading")
		return err
	}

	start := time.Now()
	limit := start.Add(time.Second * time.Duration(timeout))
	success := false
	client := resty.New().SetDebug(debugHTTP).SetLogger(logger)
	for limit.After(start) && !success {
		client.R().Get("http://www.google.com")
		_, err = client.R().SetHeader("VDCID", vdcID).SetBody(rawJSON).Post(ds4mURL + "/v2/AddVDC")
		if err != nil {
			logger.WithError(err).Error("Error received from DS4M service. Will retry again in 10 seconds")
			time.Sleep(time.Second * 10)
			start = time.Now()
		} else {
			success = true
		}
	}
	if !success {
		return errors.New("Timeout waiting for VDM to be ready")
	}
	return nil
}

// Configure creates SLAs from methods found in a blueprint, returning the Ditas monitoring adapter
// and a violation notifier that will inform the DS4M
func Configure(repo model.IRepository) (monitor.MonitoringAdapter, notifier.ViolationNotifier, error) {
	config := viper.New()

	config.SetDefault(DS4MPortProperty, DS4MDefaultPortValue)
	config.SetDefault(TestingEnabledProperty, TestingEnabledDefaultValue)
	config.SetDefault(VDMRetryTimeoutProperty, VDMRetryTimeoutDefaultValue)
	config.SetDefault(TestingNumViolationsProperty, TestingNumViolationsDefaultValue)
	config.SetDefault(DebugHTTPCallsProperty, DebugHTTPCallsDefaultValue)

	config.AddConfigPath(BlueprintLocation)
	config.SetConfigName(ConfigFileName)
	config.ReadInConfig()

	log.Infof("Read DITAS configuration file %s", config.ConfigFileUsed())

	bp, err := blueprint.ReadBlueprint(BlueprintPath)
	if err != nil {
		log.WithError(err).Error("Error reading blueprint")
		return nil, nil, err
	}

	logger := log.WithField("blueprint", bp.InternalStructure.Overview.Name)

	logger.Debug("Creating blueptint at VDM")

	vdcID := config.GetString(VDCIdPropery)
	vdmURL := fmt.Sprintf("http://vdm:%d", config.GetInt(DS4MPortProperty))
	debugHTTP := config.GetBool(DebugHTTPCallsProperty)
	err = sendBlueprintToVDM(logger, vdmURL, vdcID, config.GetInt64(VDMRetryTimeoutProperty), debugHTTP)

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

	testingConfig := TestingConfiguration{
		Enabled:       config.GetBool(TestingEnabledProperty),
		MethodID:      config.GetString(TestingMethodIDProperty),
		NumViolations: config.GetInt(TestingNumViolationsProperty),
		Metrics:       make(map[string]float64),
	}

	metrics := config.GetStringMapString(TestingMetricsProperty)
	if metrics != nil {
		for metricName, metricValue := range metrics {
			floatValue, err := strconv.ParseFloat(metricValue, 64)
			if err != nil {
				logger.WithError(err).Errorf("Error converting testing metric value of %s", metricName)
			}
			testingConfig.Metrics[metricName] = floatValue
		}
	}
	da := NewDataAnalyticsAdapter(config.GetString(DataAnalyticsURLProperty), config.GetString(VDCIdPropery), config.GetString(InfrastructureIDProperty), testingConfig, debugHTTP)
	adapter := genericadapter.New(da.Retrieve, da.Process)
	return adapter, NewNotifier(vdcID, vdmURL, testingConfig, debugHTTP), nil
}
