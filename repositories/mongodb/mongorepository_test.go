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
}

func testCreateProvider(t *testing.T) {
	model.TestCreateProvider(t, repo)
}

func testCreateProviderExists(t *testing.T) {
	model.TestCreateProviderExists(t, repo)
}

func testGetAllProviders(t *testing.T) {
	model.TestGetAllProviders(t, repo)
}

func testGetProvider(t *testing.T) {
	model.TestGetProvider(t, repo)
}

func testGetProviderNotExists(t *testing.T) {
	model.TestGetProviderNotExists(t, repo)
}

func testDeleteProvider(t *testing.T) {
	model.TestDeleteProvider(t, repo)
}

func testDeleteProviderNotExists(t *testing.T) {
	model.TestDeleteProviderNotExists(t, repo)
}

func testCreateAgreement(t *testing.T) {
	model.TestCreateAgreement(t, repo)
}

func testCreateAgreementExists(t *testing.T) {
	model.TestCreateAgreementExists(t, repo)
}

func testGetAllAgreements(t *testing.T) {
	model.TestGetAllAgreements(t, repo)
}

func testGetAgreement(t *testing.T) {
	model.TestGetAgreement(t, repo)
}

func testGetAgreementNotExists(t *testing.T) {
	model.TestGetAgreementNotExists(t, repo)
}

func testUpdateAgreementState(t *testing.T) {
	model.TestUpdateAgreementState(t, repo)
}

func testUpdateAgreementStateNotExists(t *testing.T) {
	model.TestUpdateAgreementStateNotExists(t, repo)
}

func testGetAgreementsByState(t *testing.T) {
	model.TestGetAgreementsByState(t, repo)
}

func testUpdateAgreement(t *testing.T) {
	model.TestUpdateAgreement(t, repo)
}

func testUpdateAgreementNotExists(t *testing.T) {
	model.TestUpdateAgreementNotExists(t, repo)
}

func testDeleteAgreement(t *testing.T) {
	model.TestDeleteAgreement(t, repo)
}

func testDeleteAgreementNotExists(t *testing.T) {
	model.TestDeleteAgreementNotExists(t, repo)
}

func testCreateViolation(t *testing.T) {
	model.TestCreateViolation(t, repo)
}

func testCreateViolationExists(t *testing.T) {
	model.TestCreateViolationExists(t, repo)
}

func testGetViolation(t *testing.T) {
	model.TestGetViolation(t, repo)
}

func testGetViolationNotExists(t *testing.T) {
	model.TestGetViolationNotExists(t, repo)
}
