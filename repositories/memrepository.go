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
		err = model.ErrNotFound
	}
	return &item, err
}

// CreateProvider ...
func (r MemRepository) CreateProvider(provider *model.Provider) (*model.Provider, error) {
	var err error

	id := provider.Id
	_, ok := providers[id]

	if ok {
		err = model.ErrAlreadyExist
	} else {
		providers[id] = *provider
		err = nil
	}
	return provider, err
}

// DeleteProvider ...
func (r MemRepository) DeleteProvider(provider *model.Provider) error {
	var err error

	id := provider.Id

	_, ok := providers[id]
	if ok {
		delete(providers, id)
		err = nil
	} else {
		err = model.ErrNotFound
	}
	return err
}
