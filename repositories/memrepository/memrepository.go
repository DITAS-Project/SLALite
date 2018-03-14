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
package memrepository

import (
	"SLALite/model"
	"time"

	"github.com/spf13/viper"
)

// MemRepository is a repository in memory
type MemRepository struct {
	providers  map[string]model.Provider
	agreements map[string]model.Agreement
}

func NewMemRepository(providers map[string]model.Provider, agreements map[string]model.Agreement) MemRepository {
	var r MemRepository

	if providers == nil {
		providers = make(map[string]model.Provider)
	}
	if agreements == nil {
		agreements = make(map[string]model.Agreement)
	}
	r = MemRepository{
		providers:  providers,
		agreements: agreements,
	}
	return r
}

//New creates a new instance of MemRepository
func New(config *viper.Viper) (MemRepository, error) {
	return NewMemRepository(nil, nil), nil
}

// GetAllProviders ...
func (r MemRepository) GetAllProviders() (model.Providers, error) {
	result := make(model.Providers, 0, len(r.providers))

	for _, value := range r.providers {
		result = append(result, value)
	}
	return result, nil
}

// GetProvider ...
func (r MemRepository) GetProvider(id string) (*model.Provider, error) {
	var err error

	item, ok := r.providers[id]

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
	_, ok := r.providers[id]

	if ok {
		err = model.ErrAlreadyExist
	} else {
		r.providers[id] = *provider
		err = nil
	}
	return provider, err
}

// DeleteProvider ...
func (r MemRepository) DeleteProvider(provider *model.Provider) error {
	var err error

	id := provider.Id

	_, ok := r.providers[id]
	if ok {
		delete(r.providers, id)
		err = nil
	} else {
		err = model.ErrNotFound
	}
	return err
}

func (r MemRepository) GetAllAgreements() (model.Agreements, error) {
	result := make(model.Agreements, 0, len(r.agreements))

	for _, value := range r.agreements {
		result = append(result, value)
	}
	return result, nil
}

func (r MemRepository) GetActiveAgreements() (model.Agreements, error) {
	result := make(model.Agreements, 0, len(r.agreements))

	now := time.Now()
	for _, value := range r.agreements {
		if value.State == model.STARTED && now.Before(value.Details.Expiration) {
			result = append(result, value)
		}
	}
	return result, nil
}

func (r MemRepository) GetAgreement(id string) (*model.Agreement, error) {
	var err error

	item, ok := r.agreements[id]

	if ok {
		err = nil
	} else {
		err = model.ErrNotFound
	}
	return &item, err
}

func (r MemRepository) CreateAgreement(agreement *model.Agreement) (*model.Agreement, error) {
	var err error

	id := agreement.Id
	_, ok := r.agreements[id]

	if ok {
		err = model.ErrAlreadyExist
	} else {
		r.agreements[id] = *agreement
	}
	return agreement, err
}

func (r MemRepository) DeleteAgreement(agreement *model.Agreement) error {
	var err error

	id := agreement.Id

	_, ok := r.agreements[id]
	if ok {
		delete(r.agreements, id)
	} else {
		err = model.ErrNotFound
	}
	return err
}

func (r MemRepository) StartAgreement(id string) error {
	var err error

	a, ok := r.agreements[id]

	if ok {
		a.State = model.STARTED
		r.agreements[id] = a
	} else {
		err = model.ErrNotFound
	}
	return err
}

func (r MemRepository) StopAgreement(id string) error {
	var err error

	a, ok := r.agreements[id]

	if ok {
		a.State = model.STOPPED
		r.agreements[id] = a
	} else {
		err = model.ErrNotFound
	}
	return err
}
