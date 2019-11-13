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

/*
Package memrepository is a simple implementation of a model.IRepository intended for
developing purposes.
*/
package memrepository

import (
	"SLALite/model"

	"github.com/spf13/viper"
)

// MemRepository is a repository in memory
type MemRepository struct {
	providers  map[string]model.Provider
	agreements map[string]model.Agreement
	violations map[string]model.Violation
	penalties  map[string]model.Penalty
	templates  map[string]model.Template
}

// NewMemRepository creates a MemRepository with an initial state set by the parameters
func NewMemRepository(providers map[string]model.Provider, agreements map[string]model.Agreement,
	violations map[string]model.Violation, penalties map[string]model.Penalty,
	templates map[string]model.Template) MemRepository {
	var r MemRepository

	if providers == nil {
		providers = make(map[string]model.Provider)
	}
	if agreements == nil {
		agreements = make(map[string]model.Agreement)
	}
	if violations == nil {
		violations = make(map[string]model.Violation)
	}
	if penalties == nil {
		penalties = make(map[string]model.Penalty)
	}
	if templates == nil {
		templates = make(map[string]model.Template)
	}
	r = MemRepository{
		providers:  providers,
		agreements: agreements,
		violations: violations,
		penalties:  penalties,
		templates:  templates,
	}
	return r
}

//New creates a new instance of MemRepository
func New(config *viper.Viper) (MemRepository, error) {
	return NewMemRepository(nil, nil, nil, nil, nil), nil
}

/*
GetAllProviders returns the list of providers.

The list is empty when there are no providers;
error != nil on error
*/
func (r MemRepository) GetAllProviders() (model.Providers, error) {
	result := make(model.Providers, 0, len(r.providers))

	for _, value := range r.providers {
		result = append(result, value)
	}
	return result, nil
}

/*
GetProvider returns the Provider identified by id.

error != nil on error;
error is sql.ErrNoRows if the provider is not found
*/
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

/*
CreateProvider stores a new provider.

error != nil on error;
error is sql.ErrNoRows if the provider already exists
*/
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

/*
DeleteProvider deletes from the repository the provider whose id is provider.Id.

error != nil on error;
error is sql.ErrNoRows if the provider does not exist.
*/
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

/*
GetAllAgreements returns the list of agreements.

The list is empty when there are no agreements;
error != nil on error
*/
func (r MemRepository) GetAllAgreements() (model.Agreements, error) {
	result := make(model.Agreements, 0, len(r.agreements))

	for _, value := range r.agreements {
		result = append(result, value)
	}
	return result, nil
}

/*
GetAgreementsByState returns the agreements that match any of the items in states.

error != nil on error
*/
func (r MemRepository) GetAgreementsByState(states ...model.State) (model.Agreements, error) {
	result := make(model.Agreements, 0)

	for _, a := range r.agreements {
		for _, state := range states {
			if a.State == state {
				result = append(result, a)
			}
		}
	}
	return result, nil
}

/*
GetAgreement returns the Agreement identified by id.

error != nil on error;
error is sql.ErrNoRows if the Agreement is not found
*/
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

/*
CreateAgreement stores a new Agreement.

error != nil on error;
error is sql.ErrNoRows if the Agreement already exists
*/
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

/*
UpdateAgreement updates the information of an already saved instance of an agreement
*/
func (r MemRepository) UpdateAgreement(agreement *model.Agreement) (*model.Agreement, error) {
	var err error

	id := agreement.Id
	_, ok := r.agreements[id]

	if !ok {
		err = model.ErrNotFound
	} else {
		r.agreements[id] = *agreement
	}
	return agreement, err
}

/*
DeleteAgreement deletes from the repository the Agreement whose id is provider.Id.

error != nil on error;
error is sql.ErrNoRows if the Agreement does not exist.
*/
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

/*
CreateViolation stores a new Violation.

error != nil on error;
error is sql.ErrNoRows if the Violation already exists
*/
func (r MemRepository) CreateViolation(v *model.Violation) (*model.Violation, error) {
	var err error

	id := v.Id

	if _, ok := r.violations[id]; ok {
		err = model.ErrAlreadyExist
	} else {
		r.violations[id] = *v
	}
	return v, err
}

/*
GetViolation returns the Violation identified by id.

error != nil on error;
error is sql.ErrNoRows if the Violation is not found
*/
func (r MemRepository) GetViolation(id string) (*model.Violation, error) {
	var err error

	item, ok := r.violations[id]

	if ok {
		err = nil
	} else {
		err = model.ErrNotFound
	}
	return &item, err
}

/*
UpdateAgreementState transits the state of the agreement
*/
func (r MemRepository) UpdateAgreementState(id string, newState model.State) (*model.Agreement, error) {

	var ok bool
	var err error
	var current model.Agreement
	var result *model.Agreement

	current, ok = r.agreements[id]

	if !ok {
		err = model.ErrNotFound
	} else {
		current.State = newState
		r.agreements[id] = current
		result = &current
	}
	return result, err
}

/*
GetAllTemplates returns the list of templates.

The list is empty when there are no templates;
error != nil on error
*/
func (r MemRepository) GetAllTemplates() (model.Templates, error) {

	result := make(model.Templates, 0, len(r.templates))

	for _, value := range r.templates {
		result = append(result, value)
	}
	return result, nil
}

/*
GetTemplate returns the Template identified by id.

error != nil on error;
error is sql.ErrNoRows if the Template is not found
*/
func (r MemRepository) GetTemplate(id string) (*model.Template, error) {
	var err error

	item, ok := r.templates[id]

	if ok {
		err = nil
	} else {
		err = model.ErrNotFound
	}
	return &item, err
}

/*
CreateTemplate stores a new Template.

error != nil on error;
error is sql.ErrNoRows if the Template already exists
*/
func (r MemRepository) CreateTemplate(template *model.Template) (*model.Template, error) {
	var err error

	id := template.Id
	_, ok := r.templates[id]

	if ok {
		err = model.ErrAlreadyExist
	} else {
		r.templates[id] = *template
	}
	return template, err
}
