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
	"SLALite/assessment/monitor"
	"SLALite/model"
	"context"
	"reflect"
	"time"

	"github.com/olivere/elastic"
)

type DataValue struct {
	Timestamp   time.Time `json:"@timestamp"`
	MeterValue  string    `json:"meter.value"`
	MeterUnit   string    `json:"meter.unit"`
	MeterName   string    `json:"meter.name"`
	RequestId   string    `json:"request.id"`
	RequestTime string    `json:"request.requestTime"`
}

type elasticSearchAdapter struct {
	agreement   *model.Agreement
	client      *elastic.Client
	currentData map[string][]monitor.MetricValue
	maxLength   int
}

func (ma *elasticSearchAdapter) Initialize(a *model.Agreement) {
	ma.agreement = a
	ma.currentData = make(map[string][]monitor.MetricValue)
	ma.maxLength = 0
	query := elastic.NewTermQuery("request.path", "/"+ma.agreement.Id)
	address := "http://elasticsearch:9200"
	client, _ := elastic.NewSimpleClient(
		elastic.SetURL(address),
	)
	data, err := client.Search().Index("tubvdc-*").Query(query).Do(context.Background())
	if err == nil {
		var dataValue DataValue
		for _, item := range data.Each(reflect.TypeOf(dataValue)) {
			if dataValue, ok := item.(DataValue); ok {
				if dataValue.MeterName != "" {
					values, ok := ma.currentData[dataValue.MeterName]
					if !ok {
						values = make([]monitor.MetricValue, 0)
					}
					currentValue := monitor.MetricValue{
						DateTime: dataValue.Timestamp,
						Key:      dataValue.MeterName,
						Value:    dataValue.MeterValue,
					}
					values = append(values, currentValue)
					ma.currentData[dataValue.MeterName] = values
					if len(values) > ma.maxLength {
						ma.maxLength = len(values)
					}
				}
			}
		}
	}
}

func (ma *elasticSearchAdapter) GetValues(gt model.Guarantee, vars []string) []map[string]monitor.MetricValue {
	result := make([]map[string]monitor.MetricValue, 0)

	for i := 0; i < ma.maxLength; i++ {
		iteration := make(map[string]monitor.MetricValue)
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