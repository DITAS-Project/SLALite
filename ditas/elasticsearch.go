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
	"time"

	"github.com/Knetic/govaluate"
	"github.com/olivere/elastic"
)

const (
	defaultURL              string = "http://elasticsearch:9200"
	repositoryDbName        string = "slalite"
	providersCollectionName string = "Providers"
	agreementCollectionName string = "Agreements"

	mongoConfigName string = "mongodb.yml"

	connectionURL string = "connection"
	mongoDatabase string = "database"
	clearOnBoot   string = "clear_on_boot"
)

type elasticSearchAdapter struct {
	agreement *model.Agreement
	client    *elastic.Client
}

func (ma *elasticSearchAdapter) Initialize(a *model.Agreement) {
	ma.agreement = a
}

func (ma *elasticSearchAdapter) NextValues(gt model.Guarantee) map[string]monitor.MetricValue {
	result := make(map[string]monitor.MetricValue)

	/*
	 * The following means that the adapter have knowledge about the evaluation internals,
	 * but there is a problem with cyclic dependencies if importing "/assessment"
	 */
	expression, err := govaluate.NewEvaluableExpression(gt.Constraint)
	if err != nil {
		/* TODO */
	}

	for _, key := range expression.Vars() {
		query := elastic.NewTermQuery("data", nil)
		address := "http://localhost:9200"
		client, _ := elastic.NewSimpleClient(
			elastic.SetURL(address),
		)
		data, err := client.Search().Index("tubvdc-*").Query(query).Do(context.Background())
		if err == nil {

			result[key] = monitor.MetricValue{
				DateTime: time.Now(),
				Key:      key,
				Value:    data,
			}
		}
	}
	return result
}
