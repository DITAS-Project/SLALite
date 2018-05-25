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
package bolt

import (
	"SLALite/model"
	"errors"

	log "github.com/sirupsen/logrus"

	bolt "github.com/coreos/bbolt"
	"github.com/spf13/viper"
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

func New() (BBoltRepository, error) {
	config := viper.New()

	config.SetConfigName(bboltConfigName)
	config.AddConfigPath(model.UnixConfigPath)
	config.SetDefault(databasePropertyName, bboltDatabase)

	confError := config.ReadInConfig()
	if confError != nil {
		log.Println("Can't find bbolt configuration file: " + confError.Error())
		log.Println("Using defaults")
	}

	dbName := config.GetString(databasePropertyName)
	repo := BBoltRepository{dbName}

	err := repo.SetDatabase(dbName)

	return repo, err
}

func (r *BBoltRepository) SetDatabase(database string) error {
	r.dbFile = database
	return r.ExecuteTx(nil, func(db *bolt.DB) error {
		return db.Update(func(tx *bolt.Tx) error {
			_, err2 := tx.CreateBucketIfNotExists([]byte(providerBucket))
			return err2
		})
	})
}

func (r BBoltRepository) ExecuteTx(options *bolt.Options, f func(db *bolt.DB) error) error {
	db, err := bolt.Open(r.dbFile, 0666, options)
	if err != nil {
		return err
	}
	defer db.Close()
	return f(db)
}

func (r BBoltRepository) ExecuteOperation(ft func(fn func(tx *bolt.Tx) error) error, fb func(b *bolt.Bucket) error) error {
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
	return r.ExecuteTx(&bolt.Options{ReadOnly: true}, func(db *bolt.DB) error {
		return r.ExecuteOperation(db.View, f)
	})
}

func (r BBoltRepository) ExecuteWriteOperation(f func(b *bolt.Bucket) error) error {
	return r.ExecuteTx(nil, func(db *bolt.DB) error {
		return r.ExecuteOperation(db.Update, f)
	})
}

func (r BBoltRepository) GetAllProviders() (model.Providers, error) {

	var providers model.Providers

	err := r.ExecuteReadOperation(func(b *bolt.Bucket) error {
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
			return model.ErrNotFound
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
			return model.ErrAlreadyExist
		} else {
			return b.Put([]byte(provider.Id), []byte(provider.Name))
		}
	})
}

func (r BBoltRepository) DeleteProvider(provider *model.Provider) error {
	return r.ExecuteWriteOperation(func(b *bolt.Bucket) error {
		existing := b.Get([]byte(provider.Id))
		if existing == nil {
			return model.ErrNotFound
		} else {
			return b.Delete([]byte(provider.Id))
		}
	})
}

func (r BBoltRepository) GetAllAgreements() (model.Agreements, error) {
	return nil, nil
}

func (r BBoltRepository) GetActiveAgreements() (model.Agreements, error) {
	return nil, nil
}

func (r BBoltRepository) GetAgreement(id string) (*model.Agreement, error) {
	return nil, nil
}

func (r BBoltRepository) CreateAgreement(agreement *model.Agreement) (*model.Agreement, error) {
	return nil, nil
}

func (r BBoltRepository) DeleteAgreement(agreement *model.Agreement) error {
	return nil
}

func (r BBoltRepository) StartAgreement(id string) error {
	return nil
}

func (r BBoltRepository) StopAgreement(id string) error {
	return nil
}
