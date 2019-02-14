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
This tests mongorepository, making use of the repository_testbase file.

To run this test, set up a mongodb and set env var SLA_REPOSITORY=mongodb.
If mongodb is not accessible at localhost:27017, set SLA_CONNECTION=<host>

To test other repository, copy this file, create the repository in TestMain
and remove/add methods as necessary.
*/

package mongodb

import (
	"SLALite/model"
	"os"
	"testing"

	log "github.com/sirupsen/logrus"
)

var repo model.IRepository

func TestMain(m *testing.M) {
	var err error
	result := -1

	if v, ok := os.LookupEnv("SLA_REPOSITORY"); !ok || v != Name {
		log.Info("Skipping mongodb integration test")
		os.Exit(0)
	}

	repo, err = createRepository()

	if err == nil {
		result = m.Run()
	} else {
		log.Fatal("Error creating repository: ", err.Error())
	}

	os.Exit(result)
}

func createRepository() (model.IRepository, error) {

	config, _ := NewDefaultConfig()
	config.Set("database", "slaliteTest")
	config.Set("clear_on_boot", true)
	mongoRepo, errMongo := New(config)
	return mongoRepo, errMongo
}

func TestRepository(t *testing.T) {
	ctx := model.TestContext{Repo: repo}
	/* Providers */
	t.Run("CreateProvider", ctx.TestCreateProvider)
	t.Run("CreateProviderExists", ctx.TestCreateProviderExists)
	t.Run("GetAllProviders", ctx.TestGetAllProviders)
	t.Run("GetProvider", ctx.TestGetProvider)
	t.Run("GetProviderNotExists", ctx.TestGetProviderNotExists)
	t.Run("DeleteProvider", ctx.TestDeleteProvider)
	t.Run("DeleteProviderNotExists", ctx.TestDeleteProviderNotExists)

	/* Agreements */
	t.Run("CreateAgreement", ctx.TestCreateAgreement)
	t.Run("CreateAgreementExists", ctx.TestCreateAgreementExists)
	t.Run("GetAllAgreements", ctx.TestGetAllAgreements)
	t.Run("GetAgreement", ctx.TestGetAgreement)
	t.Run("GetAgreementNotExists", ctx.TestGetAgreementNotExists)
	t.Run("UpdateAgreementState", ctx.TestUpdateAgreementState)
	t.Run("UpdateAgreementStateNotExists", ctx.TestUpdateAgreementStateNotExists)
	t.Run("GetAgreementsByState", ctx.TestGetAgreementsByState)
	t.Run("UpdateAgreement", ctx.TestUpdateAgreement)
	t.Run("UpdateAgreementNotExists", ctx.TestUpdateAgreementNotExists)
	t.Run("DeleteAgreement", ctx.TestDeleteAgreement)
	t.Run("DeleteAgreementNotExists", ctx.TestDeleteAgreementNotExists)

	/* Violations */
	// t.Run("CreateViolation", ctx.TestCreateViolation)
	// t.Run("CreateViolationExists", ctx.TestCreateViolationExists)

	// t.Run("GetViolation", ctx.TestGetViolation)
	// t.Run("GetViolationNotExists", ctx.TestGetViolationNotExists)

	/* Templates */
	// t.Run("CreateTemplate", ctx.TestCreateTemplate)
	// t.Run("CreateTemplateExists", ctx.TestCreateTemplateExists)
	// t.Run("GetAllTemplates", ctx.TestGetAllTemplates)
	// t.Run("GetTemplate", ctx.TestGetTemplate)
	// t.Run("GetTemplateNotExists", ctx.TestGetTemplateNotExists)
}
