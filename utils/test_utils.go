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
package utils

import (
	"SLALite/model"
	"SLALite/repositories/memrepository"
	"SLALite/repositories/mongodb"
	"SLALite/repositories/validation"
	"log"
	"os"
	"strings"
)

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
	repo, _ = validation.New(repo)
	return repo
}
