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
	assessment_model "SLALite/assessment/model"
	"SLALite/model"
	"context"
	"encoding/json"
	"time"

	"github.com/DITAS-Project/blueprint-go"
	log "github.com/sirupsen/logrus"

	"github.com/olivere/elastic"
)

const (
	ResponseTimeKey = "responseTime"
)

type Logger struct {
}

type MeterType struct {
	Timestamp   time.Time   `json:"timestamp"`
	OperationID string      `json:"operationID"`
	Value       interface{} `json:"value"`
	Unit        string      `json:"unit"`
	Name        string      `json:"name"`
	Appendix    string      `json:"appendix"`
}
type DataValue struct {
	Timestamp   time.Time  `json:"@timestamp"`
	Meter       *MeterType `json:"meter"`
	RequestId   string     `json:"request.id"`
	RequestTime *int       `json:"request.requestTime"`
}

type ElasticSearchAdapter struct {
	agreement   *model.Agreement
	methodInfo  map[string]blueprint.ExtendedOps
	client      *elastic.Client
	currentData map[string][]model.MetricValue
	maxLength   int
}

func (l Logger) Printf(format string, v ...interface{}) {
	//log.Debugf(format, v)
}

func NewAdapter(url string, methodInfo map[string]blueprint.ExtendedOps) *ElasticSearchAdapter {
	address := url
	logger := Logger{}
	client, _ := elastic.NewSimpleClient(
		elastic.SetURL(address),
		elastic.SetTraceLog(logger),
		elastic.SetInfoLog(logger),
		elastic.SetErrorLog(logger),
	)
	return &ElasticSearchAdapter{
		client:     client,
		methodInfo: methodInfo,
	}
}

func (ma *ElasticSearchAdapter) addMetric(value model.MetricValue) {
	values, ok := ma.currentData[value.Key]
	if !ok {
		values = make([]model.MetricValue, 0)
	}
	values = append(values, value)
	ma.currentData[value.Key] = values
	if len(values) > ma.maxLength {
		ma.maxLength = len(values)
	}
}

func (ma *ElasticSearchAdapter) addValue(value DataValue) {
	currentValue := model.MetricValue{
		DateTime: value.Timestamp,
	}

	if value.Meter != nil {
		currentValue.Key = value.Meter.Name
		currentValue.Value = value.Meter.Value
	} else {
		if value.RequestTime != nil {
			currentValue.Key = ResponseTimeKey
			currentValue.Value = *value.RequestTime
		}
	}

	if currentValue.Key != "" {
		ma.addMetric(currentValue)
	} else {
		log.Errorf("Found invalid value without response time and meter name %s", value)
	}
}

func (ma *ElasticSearchAdapter) addValues(query *elastic.MatchQuery) {
	pageSize := 1000
	currentPage := 1

	currentQuery := ma.client.Search().Index("tubvdc-*").Query(query).
		Sort("@timestamp", true).From(0).Size(pageSize)
	last := int64(currentPage * pageSize)
	var err error
	for data, err := currentQuery.Do(context.Background()); data.TotalHits() > 0 && len(data.Hits.Hits) > 0 && err == nil; data, err = currentQuery.Do(context.Background()) {
		var dataValue DataValue
		for _, hit := range data.Hits.Hits {
			err := json.Unmarshal(*hit.Source, &dataValue)
			if err != nil {
				log.Errorf("Error unmarshalling hit: %s", err.Error())
			} else {
				ma.addValue(dataValue)
			}
		}

		currentPage++
		to := currentPage * pageSize
		currentQuery = currentQuery.From(int(last)).Size(to)
		last = int64(to)
	}

	if err != nil {
		log.Errorf("Error iterating over results: %s", err.Error())
	}
}

func (ma *ElasticSearchAdapter) Initialize(a *model.Agreement) {
	ma.agreement = a
	ma.currentData = make(map[string][]model.MetricValue)
	ma.maxLength = 0
	ma.addValues(elastic.NewMatchQuery("request.operationID", ma.agreement.Id))
	ma.addValues(elastic.NewMatchQuery("meter.operationID", ma.agreement.Id))
}

func (ma *ElasticSearchAdapter) GetValues(gt model.Guarantee, vars []string) assessment_model.GuaranteeData {
	result := make(assessment_model.GuaranteeData, 0)

	empty := false
	for i := 0; i < ma.maxLength && !empty; i++ {
		iteration := make(assessment_model.ExpressionData)
		for _, v := range vars {
			vals, ok := ma.currentData[v]
			if ok && i < len(vals) {
				iteration[v] = vals[i]
			}
		}
		if len(iteration) > 0 {
			result = append(result, iteration)
		} else {
			empty = true
		}
	}

	return result
}
