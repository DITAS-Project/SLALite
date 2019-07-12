/*
Copyright 2018 Atos

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

/*
Package utils contain util functions.
*/
package utils

import (
	"SLALite/model"
	"SLALite/repositories/memrepository"
	"SLALite/repositories/mongodb"
	"SLALite/repositories/validation"
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

// CreateTestRepository creates a repository according to the SLA_REPOSITORY env var.
func CreateTestRepository() model.IRepository {
	envvar := "SLA_" + strings.ToUpper(RepositoryTypePropertyName)
	repotype, ok := os.LookupEnv(envvar)
	if !ok {
		repotype = DefaultRepositoryType
	}
	return createRepository(repotype)
}

func createRepository(repoType string) model.IRepository {
	var repo model.IRepository

	switch repoType {
	case DefaultRepositoryType:
		memrepo, _ := memrepository.New(nil)
		repo = memrepo
	case "mongodb":
		config, _ := mongodb.NewDefaultConfig()
		config.Set("database", "slaliteTest")
		config.Set("clear_on_boot", true)
		mongoRepo, errMongo := mongodb.New(config)
		if errMongo != nil {
			log.Fatal("Error creating mongo repository: ", errMongo.Error())
		}
		repo = mongoRepo
	}
	repo, _ = validation.New(repo, model.NewDefaultValidator(false, true))
	return repo
}

// Timeline calculates delta times from a time origin
// Inialize the struct with t0 as your desired time origin
// Ex.:
//    t = Timeline { T0: time.Now() }
type Timeline struct {
	T0 time.Time
}

// T calculates the delta in seconds with respect to the origin
// Ex:
//     t.T(2)
//     t.T(-1)
func (t *Timeline) T(second float64) time.Time {
	ms := int(1000 * second)
	d := time.Duration(ms) * time.Millisecond
	return t.T0.Add(d)
}

// ReadAgreement returns the agreement read from the file pointed by path.
// The CWD is the location of the test.
//
// Ex:
//    a, err := readAgreement("testdata/a.json")
//    if err != nil {
//      t.Errorf("Error reading agreement: %v", err)
//    }
func ReadAgreement(path string) (model.Agreement, error) {
	/*
	 * Moved to model to avoid cyclic dependency when used in model package tests
	 */
	return model.ReadAgreement(path)
}

// ReadTemplate returns the template read from the file pointed by path.
// The CWD is the location of the test.
//
// Ex:
//    a, err := readAgreement("testdata/a.json")
//    if err != nil {
//      t.Errorf("Error reading agreement: %v", err)
//    }
func ReadTemplate(path string) (model.Template, error) {
	/*
	 * Moved to model to avoid cyclic dependency when used in model package tests
	 */
	return model.ReadTemplate(path)
}
