package repositories

import (
	"github.com/spf13/viper"
	"log"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"SLALite/model"
	"database/sql"
)

const (
	DEFAULT_HOST string = "localhost"
	REPOSITORY_DB_NAME string = "slalite"
	PROVIDERS_COLLECTION_NAME string = "Providers"

	hostPropertyName string = "host"
)

type MongoDBRepository struct {
	session *mgo.Session
	database *mgo.Database
}

func CreateMongoDBRepository() (MongoDBRepository, error) {
	config := viper.New()

	config.SetDefault(hostPropertyName, DEFAULT_HOST)

	confError := config.ReadInConfig()
	if confError != nil {
		log.Println("Can't find MongoDB configuration file: " + confError.Error())
		log.Println("Using defaults")
	}

	session, err := mgo.Dial(config.GetString(hostPropertyName))
	if err != nil {
		log.Fatal("Error getting connectin to Mongo DB: " + err.Error())
	}

	return MongoDBRepository{session,session.DB(REPOSITORY_DB_NAME)}, err
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
	var result model.Providers

	err := r.database.C(PROVIDERS_COLLECTION_NAME).Find(bson.M{}).All(&result);

	return result, err;
}

func (r MongoDBRepository) GetProvider(id string) (*model.Provider, error) {
	var result *model.Provider = new(model.Provider)

	err := r.database.C(PROVIDERS_COLLECTION_NAME).Find(bson.M{"id": id}).One(result)
	if result.Id == "" {
		return result, sql.ErrNoRows
	}

	return result,err
}

func (r MongoDBRepository) CreateProvider(provider *model.Provider) (*model.Provider, error) {
	existing, _ := r.GetProvider(provider.Id)
	if existing.Id != "" {
		return existing,sql.ErrNoRows
	}
	errCreate := r.database.C(PROVIDERS_COLLECTION_NAME).Insert(provider)
	return provider, errCreate
}


