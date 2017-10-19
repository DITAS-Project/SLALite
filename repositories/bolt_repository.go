package repositories

import (
	bolt "github.com/coreos/bbolt"
	"errors"
	"database/sql"
	"github.com/spf13/viper"
	"log"
	"SLALite/model"
)

const (
	providerBucket string = "Providers"
	bboltDatabase  string = "slalite.db"


	bboltConfigName = "bbolt.yml"

	databasePropertyName = "database"
)

type BBoltRepository struct {
	dbFile string
}

func CreateBBoltRepository() (BBoltRepository, error) {
	config := viper.New()

	config.SetConfigName(bboltConfigName)
	config.AddConfigPath(model.UnixConfigPath)
	config.SetDefault(databasePropertyName, bboltDatabase)

	confError := config.ReadInConfig()
	if confError != nil {
		log.Println("Can't find bbolt configuration file: " + confError.Error())
		log.Println("Using defaults")
	}

	repo := BBoltRepository{config.GetString(databasePropertyName)}

	err := repo.ExecuteTx(nil, func(db *bolt.DB) error {
		return db.Update(func (tx *bolt.Tx) error {
			_, err2 := tx.CreateBucketIfNotExists([]byte(providerBucket))
			return err2
		})
	})

	return repo, err
}

func (r BBoltRepository) ExecuteTx(options *bolt.Options, f func(db *bolt.DB) error) error {
	db, err := bolt.Open(r.dbFile, 0666, options)
	if err != nil {
		return err
	}
	defer db.Close()
	return f(db)
}

func (r BBoltRepository) ExecuteOperation(ft func (fn func(tx *bolt.Tx) error) error, fb func(b *bolt.Bucket) error ) error {
	return ft(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(providerBucket))
		if b != nil {
			return fb(b)
		} else {
			return errors.New("Error getting providers bucket")
		}
	})
}

func (r BBoltRepository) ExecuteReadOperation(f func(b *bolt.Bucket) error) error {
	return r.ExecuteTx(&bolt.Options{ReadOnly: true}, func (db *bolt.DB) error {
		return r.ExecuteOperation(db.View, f)
	})
}

func (r BBoltRepository) ExecuteWriteOperation(f func(b *bolt.Bucket) error) error {
	return r.ExecuteTx(nil, func (db *bolt.DB) error {
		return r.ExecuteOperation(db.Update, f)
	})
}

func (r BBoltRepository) GetAllProviders() (model.Providers, error) {

	var providers model.Providers

	err := r.ExecuteReadOperation(func (b *bolt.Bucket) error {
		b.ForEach(func(k, v []byte) error {
			providers = append(providers, model.Provider{string(k), string(v)})
			return nil
		})

		return nil
	})
	return providers, err
}

func (r BBoltRepository) GetProvider(id string) (*model.Provider, error) {
	var provider *model.Provider = nil
	err := r.ExecuteReadOperation(func(b *bolt.Bucket) error {
		value := b.Get([]byte(id))
		if value != nil {
			provider = &model.Provider{id, string(value)}
		} else {
			return sql.ErrNoRows
		}
		return nil
	})
	return provider, err
}

func (r BBoltRepository) CreateProvider(provider *model.Provider) (*model.Provider, error) {
	//log.Info("Trying to create provider " + provider.Id)
	return provider, r.ExecuteWriteOperation(func(b *bolt.Bucket) error {
		existing := b.Get([]byte(provider.Id))
		if existing != nil {
			return sql.ErrNoRows
		} else {
			return b.Put([]byte(provider.Id), []byte(provider.Name))
		}
	})
}
