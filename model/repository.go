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
	 * error != nil on error
	 * error is sql.ErrNoRows if the provider is not found
	 */
	GetProvider(id string) (*Provider, error)

	/*
	 * CreateProvider stores a new provider
	 * error != nil on error
	 * error is sql.ErrNoRows if the provider already exists
	 */
	CreateProvider(provider *Provider) (*Provider, error)

	/*
	 * DeleteProvider deletes from the repository the provider whose id is provider.Id.
	 * error != nil on error
	 * error is sql.ErrNoRows if the provider does not exist.
	 */
	DeleteProvider(provider *Provider) error

	/*
	 * GetAllAgreements returns the list of agreements.
	 * The list is empty when there are no agreements
	 * error != nil on error
	 */
	GetAllAgreements() (Agreements, error)

	/*
	 * GetAgreement returns the Agreement identified by id
	 * error != nil on error
	 * error is sql.ErrNoRows if the Agreement is not found
	 */
	GetAgreement(id string) (*Agreement, error)

	/*
	 * CreateAgreement stores a new Agreement
	 * error != nil on error
	 * error is sql.ErrNoRows if the Agreement already exists
	 */
	CreateAgreement(agreement *Agreement) (*Agreement, error)

	/*
	 * DeleteAgreement deletes from the repository the Agreement whose id is provider.Id.
	 * error != nil on error
	 * error is sql.ErrNoRows if the Agreement does not exist.
	 */
	DeleteAgreement(agreement *Agreement) error
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
