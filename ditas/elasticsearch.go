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
	"reflect"
	"time"

	"github.com/DITAS-Project/blueprint-go"

	"github.com/olivere/elastic"
)

const (
	ResponseTimeKey = "ResponseTime"
)

type DataValue struct {
	Timestamp   time.Time `json:"@timestamp"`
	MeterValue  string    `json:"meter.value"`
	MeterUnit   string    `json:"meter.unit"`
	MeterName   string    `json:"meter.name"`
	RequestId   string    `json:"request.id"`
	RequestTime *int      `json:"request.requestTime"`
}

type ElasticSearchAdapter struct {
	agreement   *model.Agreement
	methodInfo  map[string]blueprint.ExtendedOps
	client      *elastic.Client
	currentData map[string][]model.MetricValue
	maxLength   int
}

func NewAdapter(methodInfo map[string]blueprint.ExtendedOps) *ElasticSearchAdapter {
	address := "http://elasticsearch:9200"
	client, _ := elastic.NewSimpleClient(
		elastic.SetURL(address),
	)
	return &ElasticSearchAdapter{
		client:     client,
		methodInfo: methodInfo,
	}
}

func (ma *ElasticSearchAdapter) addValue(value model.MetricValue) {
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

func (ma *ElasticSearchAdapter) Initialize(a *model.Agreement) {
	ma.agreement = a
	ma.currentData = make(map[string][]model.MetricValue)
	ma.maxLength = 0
	query := elastic.NewTermQuery("operationId", ma.agreement.Id)
	data, err := ma.client.Search().Index("tubvdc-*").Query(query).Do(context.Background())
	if err == nil {
		var dataValue DataValue
		for _, item := range data.Each(reflect.TypeOf(dataValue)) {
			if dataValue, ok := item.(DataValue); ok {
				if dataValue.RequestTime != nil {
					responseTime := model.MetricValue{
						DateTime: dataValue.Timestamp,
						Key:      ResponseTimeKey,
						Value:    *dataValue.RequestTime,
					}
					ma.addValue(responseTime)
				}
				if dataValue.MeterName != "" {
					currentValue := model.MetricValue{
						DateTime: dataValue.Timestamp,
						Key:      dataValue.MeterName,
						Value:    dataValue.MeterValue,
					}
					ma.addValue(currentValue)
				}
			}
		}
	}
}

func (ma *ElasticSearchAdapter) GetValues(gt model.Guarantee, vars []string) assessment_model.GuaranteeData {
	result := make(assessment_model.GuaranteeData, 0)

	for i := 0; i < ma.maxLength; i++ {
		iteration := make(assessment_model.ExpressionData)
		for _, v := range vars {
			vals, ok := ma.currentData[v]
			if ok && i < len(vals) {
				iteration[v] = vals[i]
			}
		}
		result = append(result, iteration)
	}

	return result
}
