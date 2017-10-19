package model

import "database/sql"

const (
	UnixConfigPath = "/etc/slalite"
)

// IRepository expose the interface to be fulfilled by implementations of repositories.
type IRepository interface {
	/*
	 * GetAllProviders returns the list of providers.
	 * The list is empty when there are no providers
	 * error != nil on error
	 */
	GetAllProviders() (Providers, error)

	/*
	 * GetProvider returns the Provider identified by id
	 * error is sql.ErrNoRows if the provider is not found
	 */
	GetProvider(id string) (*Provider, error)

	/*
	 * CreateProvider stores a new provider
	 * error is sql.ErrNoRows if the provider already exists
	 */
	CreateProvider(provider *Provider) (*Provider, error)
}

// DbRepository is a repository backed up on a database
type DbRepository struct {
	db *sql.DB
}

// GetAllProviders ...
func (r DbRepository) GetAllProviders() Providers {
	/* code goes here
	 * r.db.blabla
	 */
	providers := Providers{}
	return providers
}

// GetProvider ...
func (r DbRepository) GetProvider(id string) *Provider {
	return &Provider{}
}

// CreateProvider ...
func (r DbRepository) CreateProvider(provider *Provider) {
}
