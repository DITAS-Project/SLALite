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
	"SLALite/model"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
)

type TestingConfiguration struct {
	Enabled       bool
	MethodID      string
	NumViolations int
	Metrics       map[string]float64
}
type DataAnalyticsMeter struct {
	OperationID string  `json:"operationID"`
	Name        string  `json:"name"`
	Value       float64 `json:"value"`
	Unit        string  `json:"unit"`
	Timestamp   string  `json:"timestamp"`
	Appendix    string  `json:"appendix"`
}
type DataAnalyticsMetrics struct {
	DataAnalyticsMeter DataAnalyticsMeter `json:"meter"`
}
type DataAnalyticsAdapter struct {
	Client               *resty.Client
	AnalyticsBaseUrl     string
	VdcID                string
	InfraID              string
	TestingConfiguration TestingConfiguration
}

func NewDataAnalyticsAdapter(analyticsBaseUrl, vdcID, infraID string, testingConfig TestingConfiguration, debugHTTP bool) *DataAnalyticsAdapter {
	client := resty.New().SetDebug(debugHTTP)
	return &DataAnalyticsAdapter{
		Client:               client,
		AnalyticsBaseUrl:     analyticsBaseUrl,
		VdcID:                vdcID,
		InfraID:              infraID,
		TestingConfiguration: testingConfig,
	}
}

// Initialize the monitoring retrieval for one evaluation of the agreement
//
// A new MonitoringAdapter, copy of current adapter, must be returned
func (d DataAnalyticsAdapter) Retrieve(agreement model.Agreement,
	items []monitor.RetrievalItem) map[model.Variable][]model.MetricValue {
	result := make(map[model.Variable][]model.MetricValue)

	useTesting := d.TestingConfiguration.Enabled && agreement.Id == d.TestingConfiguration.MethodID
	for _, item := range items {
		// Due to bug https://github.com/spf13/viper/issues/411 which transforms keys to lowercase, we need to check against the lowercase representation of the metric
		lowerMetric := strings.ToLower(item.Var.Metric)
		if metricValue, ok := d.TestingConfiguration.Metrics[lowerMetric]; ok && useTesting {
			result[item.Var] = []model.MetricValue{
				model.MetricValue{
					Key:      item.Var.Metric,
					Value:    metricValue,
					DateTime: time.Now(),
				},
			}
		} else {
			metrics := make([]DataAnalyticsMetrics, 0)
			res, err := d.Client.R().SetQueryParams(map[string]string{
				"operationID": agreement.Id,
				"name":        item.Var.Metric,
				"startTime":   item.From.Format(time.RFC3339),
				"endTime":     item.To.Format(time.RFC3339),
			}).SetPathParams(map[string]string{
				"infraId": d.VdcID,
			}).SetResult(&metrics).Get(d.AnalyticsBaseUrl)
			if err != nil {
				log.WithError(err).Errorf("Error getting values for metric %s", item.Var.Metric)
			} else {
				if !res.IsError() {
					currentMetrics, ok := result[item.Var]
					if !ok {
						currentMetrics = make([]model.MetricValue, 0, len(metrics))
						log.Printf("len(metrics): %d", len(metrics))
					}
					for _, metric := range metrics {
						metricTime, err := time.Parse(time.RFC3339, metric.DataAnalyticsMeter.Timestamp)
						if err != nil {
							log.WithError(err).Errorf("Error parsing timestamp %s for metric %s", metric.DataAnalyticsMeter.Timestamp, item.Var.Metric)
						} else {
							currentMetrics = append(currentMetrics, model.MetricValue{
								Key:      item.Var.Metric,
								Value:    metric.DataAnalyticsMeter.Value,
								DateTime: metricTime,
							})
						}
					}
					result[item.Var] = currentMetrics
				}
			}
		}
	}

	return result
}

func (d *DataAnalyticsAdapter) Process(v model.Variable, values []model.MetricValue) []model.MetricValue {
	sum := 0.0
	for _, value := range values {
		sum += value.Value.(float64)
	}
	//Availability is converted to a percentage
	result := sum / float64(len(values))
	if v.Name == "availability" {
		result = result * 100

	}

	processTime := time.Now()
	if len(values) > 0 {
		processTime = values[0].DateTime
	}

	return []model.MetricValue{
		model.MetricValue{
			Key:      v.Metric,
			Value:    result,
			DateTime: processTime,
		},
	}
}
