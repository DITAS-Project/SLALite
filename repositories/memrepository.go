package repositories

import (
	"database/sql"
	"SLALite/model"
)

// MemRepository is a repository in memory
type MemRepository struct {
}

var providers = map[string]model.Provider{
// "01": Provider{Id: "01", Name: "provider01"},
// "02": Provider{Id: "02", Name: "provider02"},
}

// GetAllProviders ...
func (r MemRepository) GetAllProviders() (model.Providers, error) {

	result := make(model.Providers, 0, len(providers))

	for _, value := range providers {
		result = append(result, value)
	}
	return result, nil
}

// GetProvider ...
func (r MemRepository) GetProvider(id string) (*model.Provider, error) {
	var err error

	item, ok := providers[id]

	if ok {
		err = nil
	} else {
		err = sql.ErrNoRows
	}
	return &item, err
}

// CreateProvider ...
func (r MemRepository) CreateProvider(provider *model.Provider) (*model.Provider, error) {
	var err error

	id := provider.Id
	_, ok := providers[id]

	if ok {
		err = sql.ErrNoRows
	} else {
		providers[id] = *provider
		err = nil
	}
	return provider, err
}
