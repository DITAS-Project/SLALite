package repositories

import (
	"SLALite/model"
	"log"

	"github.com/spf13/viper"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	defaultUrl              string = "localhost"
	repositoryDbName        string = "slalite"
	providersCollectionName string = "Providers"

	mongoConfigName string = "mongodb.yml"

	connectionUrl string = "connection"
)

type MongoDBRepository struct {
	session  *mgo.Session
	database *mgo.Database
}

func CreateMongoDBRepository() (MongoDBRepository, error) {
	config := viper.New()

	config.SetConfigName(mongoConfigName)
	config.AddConfigPath(model.UnixConfigPath)
	config.SetDefault(connectionUrl, defaultUrl)

	confError := config.ReadInConfig()
	if confError != nil {
		log.Println("Can't find MongoDB configuration file: " + confError.Error())
		log.Println("Using defaults")
	}

	session, err := mgo.Dial(config.GetString(connectionUrl))
	if err != nil {
		log.Fatal("Error getting connection to Mongo DB: " + err.Error())
	}

	return MongoDBRepository{session, session.DB(repositoryDbName)}, err
}

func (r *MongoDBRepository) SetDatabase(database string, empty bool) {
	if empty {
		err := r.session.DB(database).DropDatabase()
		if err != nil {
			log.Println("Error dropping database " + database + ": " + err.Error())
		}
	}
	r.database = r.session.DB(database)
}

func (r MongoDBRepository) GetAllProviders() (model.Providers, error) {
	var result *model.Providers = new(model.Providers)

	err := r.database.C(providersCollectionName).Find(bson.M{}).All(result)

	return *result, err
}

func (r MongoDBRepository) GetProvider(id string) (*model.Provider, error) {
	var result *model.Provider = new(model.Provider)

	err := r.database.C(providersCollectionName).Find(bson.M{"id": id}).One(result)
	if result.Id == "" {
		return result, model.ErrNotFound
	}

	return result, err
}

func (r MongoDBRepository) CreateProvider(provider *model.Provider) (*model.Provider, error) {
	existing, _ := r.GetProvider(provider.Id)
	if existing.Id != "" {
		return existing, model.ErrAlreadyExist
	}
	errCreate := r.database.C(providersCollectionName).Insert(provider)
	return provider, errCreate
}

func (r MongoDBRepository) DeleteProvider(provider *model.Provider) error {
	error := r.database.C(providersCollectionName).Remove(bson.M{"id": provider.Id})
	if error == mgo.ErrNotFound {
		return model.ErrNotFound
	} else {
		return error
	}
}
