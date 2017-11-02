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
package repositories

import (
	"SLALite/model"
	"log"

	"github.com/spf13/viper"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	defaultURL              string = "localhost"
	repositoryDbName        string = "slalite"
	providersCollectionName string = "Providers"
	agreementCollectionName string = "Agreements"

	mongoConfigName string = "mongodb.yml"

	connectionURL string = "connection"
)

//MongoDBRepository contains the repository persistence implementation based on MongoDB
type MongoDBRepository struct {
	session  *mgo.Session
	database *mgo.Database
}

//CreateMongoDBRepository creates a new instance of the MongoDBRepository with the database configurarion read from a configuration file
func CreateMongoDBRepository() (MongoDBRepository, error) {
	config := viper.New()

	config.SetConfigName(mongoConfigName)
	config.AddConfigPath(model.UnixConfigPath)
	config.SetDefault(connectionURL, defaultURL)

	confError := config.ReadInConfig()
	if confError != nil {
		log.Println("Can't find MongoDB configuration file: " + confError.Error())
		log.Println("Using defaults")
	}

	session, err := mgo.Dial(config.GetString(connectionURL))
	if err != nil {
		log.Fatal("Error getting connection to Mongo DB: " + err.Error())
	}

	return MongoDBRepository{session, session.DB(repositoryDbName)}, err
}

//SetDatabase sets the database URL value. Useful for testing.
func (r *MongoDBRepository) SetDatabase(database string, empty bool) {
	if empty {
		err := r.session.DB(database).DropDatabase()
		if err != nil {
			log.Println("Error dropping database " + database + ": " + err.Error())
		}
	}
	r.database = r.session.DB(database)
}

func (r MongoDBRepository) getAll(collection string, result interface{}) (interface{}, error) {
	err := r.database.C(collection).Find(bson.M{}).All(result)
	return result, err
}

func (r MongoDBRepository) get(collection string, id string, result model.Identity) (model.Identity, error) {
	err := r.database.C(collection).Find(bson.M{"id": id}).One(result)
	if err == mgo.ErrNotFound {
		return result, model.ErrNotFound
	}

	return result, err
}

func (r MongoDBRepository) create(collection string, object model.Identity) (model.Identity, error) {
	_, err := r.get(collection, object.GetId(), object)
	if err != model.ErrNotFound {
		return object, model.ErrAlreadyExist
	}
	errCreate := r.database.C(collection).Insert(object)
	return object, errCreate
}

func (r MongoDBRepository) delete(collection string, object model.Identity) error {
	error := r.database.C(collection).Remove(bson.M{"id": object.GetId()})
	if error == mgo.ErrNotFound {
		return model.ErrNotFound
	}
	return error

}

func (r MongoDBRepository) GetAllProviders() (model.Providers, error) {
	res, err := r.getAll(providersCollectionName, new(model.Providers))
	return *((res).(*model.Providers)), err
}

func (r MongoDBRepository) GetProvider(id string) (*model.Provider, error) {
	res, err := r.get(providersCollectionName, id, new(model.Provider))
	return res.(*model.Provider), err
}

func (r MongoDBRepository) CreateProvider(provider *model.Provider) (*model.Provider, error) {
	res, err := r.create(providersCollectionName, provider)
	return res.(*model.Provider), err
}

func (r MongoDBRepository) DeleteProvider(provider *model.Provider) error {
	return r.delete(providersCollectionName, provider)
}

func (r MongoDBRepository) GetAllAgreements() (model.Agreements, error) {
	res, err := r.getAll(agreementCollectionName, new(model.Agreements))
	return (res).(model.Agreements), err
}

func (r MongoDBRepository) GetAgreement(id string) (*model.Agreement, error) {
	res, err := r.get(agreementCollectionName, id, new(model.Agreement))
	return res.(*model.Agreement), err
}

func (r MongoDBRepository) CreateAgreement(agreement *model.Agreement) (*model.Agreement, error) {
	res, err := r.create(agreementCollectionName, agreement)
	return res.(*model.Agreement), err
}

func (r MongoDBRepository) DeleteAgreement(agreement *model.Agreement) error {
	return r.delete(agreementCollectionName, agreement)
}
