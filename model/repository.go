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
