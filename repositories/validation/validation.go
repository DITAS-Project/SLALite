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

// Package validation provides a repository decorator that provides in-memory validation
// before calling the decorated repository.
//
// This decorator does not provide a ForeignKey-like functionality. This should be provided
// by business logic (or other decorator).
//
// Usage:
//   repo, err := mongodb.New(config)
//   repo, _ = validation.New(repo)
//
package validation

import (
	"SLALite/model"
	"bytes"
	"fmt"
)

const (
	fakeID = "_"
)

type repository struct {
	backend     model.IRepository
	externalIDs bool
}

type valError struct {
	msg string
}

func (e *valError) Error() string {
	return e.msg
}

func newValError(errs []error) *valError {
	var buffer bytes.Buffer
	for _, err := range errs {
		buffer.WriteString(err.Error())
		buffer.WriteString(". ")
	}
	return &valError{msg: buffer.String()}
}

func (e *valError) IsErrValidation() bool {
	return true
}

// New returns an IRepository that performs validation before calling the actual repository.
// The externalIDs parameter is true when the Id of the entity is set by the repository,
// and therefore, out of the control of the SLALite; in this case, we cannot enforce that
// the Id is set when creating an entity.
func New(backend model.IRepository, externalIDs bool) (model.IRepository, error) {
	return repository{
		backend:     backend,
		externalIDs: externalIDs,
	}, nil
}

// GetAllProviders gets all providers.
func (r repository) GetAllProviders() (model.Providers, error) {
	return r.backend.GetAllProviders()
}

// GetProviders get a provider.
func (r repository) GetProvider(id string) (*model.Provider, error) {
	return r.backend.GetProvider(id)
}

// CreateProvider validates and persists a provider.
func (r repository) CreateProvider(provider *model.Provider) (*model.Provider, error) {

	finalID, idErr := r.checkIDErrOnCreate(provider)
	if idErr != nil {
		return provider, idErr
	}
	if r.externalIDs {
		provider.Id = fakeID
	}
	if errs := provider.Validate(); len(errs) > 0 {
		err := newValError(errs)
		return provider, err
	}
	provider.Id = finalID
	return r.backend.CreateProvider(provider)
}

// DeleteProvider deletes a provider from repository.
func (r repository) DeleteProvider(provider *model.Provider) error {

	return r.backend.DeleteProvider(provider)
}

// GetAllAgreements gets all agreements.
func (r repository) GetAllAgreements() (model.Agreements, error) {
	return r.backend.GetAllAgreements()
}

// GetAgreement gets an agreement by id
func (r repository) GetAgreement(id string) (*model.Agreement, error) {
	return r.backend.GetAgreement(id)
}

// GetAgreementsByState returns the agreements that have one of the items in states.
func (r repository) GetAgreementsByState(states ...model.State) (model.Agreements, error) {
	return r.backend.GetAgreementsByState(states...)
}

// CreateAgreement validates and persists an agreement.
func (r repository) CreateAgreement(agreement *model.Agreement) (*model.Agreement, error) {
	finalID, idErr := r.checkIDErrOnCreate(agreement)
	if idErr != nil {
		return agreement, idErr
	}
	if r.externalIDs {
		agreement.Id = agreement.Details.Id
	}
	if errs := agreement.Validate(); len(errs) > 0 {
		err := newValError(errs)
		return agreement, err
	}
	agreement.Id = finalID
	return r.backend.CreateAgreement(agreement)
}

// UpdateAgreement validates and updates an agreement.
func (r repository) UpdateAgreement(agreement *model.Agreement) (*model.Agreement, error) {

	/*
		It does not validate change of State.
	*/

	finalID := agreement.Id

	idErr := r.checkIDErrOnUpdate(agreement)
	if idErr != nil {
		return agreement, idErr
	}
	if r.externalIDs {
		agreement.Id = agreement.Details.Id
	}

	if errs := agreement.Validate(); len(errs) > 0 {
		err := newValError(errs)
		return agreement, err
	}
	agreement.Id = finalID
	return r.backend.UpdateAgreement(agreement)
}

// DeleteAgreement deletes an agreement from repository.
func (r repository) DeleteAgreement(agreement *model.Agreement) error {
	return r.backend.DeleteAgreement(agreement)
}

// CreateViolation validates and persists a new Violation.
func (r repository) CreateViolation(v *model.Violation) (*model.Violation, error) {

	finalID, idErr := r.checkIDErrOnCreate(v)
	if idErr != nil {
		return v, idErr
	}
	if r.externalIDs {
		v.Id = fakeID
	}

	if errs := v.Validate(); len(errs) > 0 {
		err := newValError(errs)
		return v, err
	}
	v.Id = finalID
	return r.backend.CreateViolation(v)
}

// GetViolation returns the Violation identified by id.
func (r repository) GetViolation(id string) (*model.Violation, error) {
	return r.backend.GetViolation(id)
}

// UpdateAgreement changes the state of an Agreement.
func (r repository) UpdateAgreementState(id string, newState model.State) (*model.Agreement, error) {
	var err error
	newState = newState.Normalize()

	current, err := r.GetAgreement(id)
	if err != nil {
		return nil, err
	}
	if !current.IsValidTransition(newState) {
		msg := fmt.Sprintf("Not valid transition from %s to %s for agreement %s",
			current.State, newState, id)
		err := &valError{msg: msg}
		return nil, err
	}
	return r.backend.UpdateAgreementState(id, newState)
}

// GetAllTemplates gets all Templates.
func (r repository) GetAllTemplates() (model.Templates, error) {
	return r.backend.GetAllTemplates()
}

// GetTemplate gets an template by id
func (r repository) GetTemplate(id string) (*model.Template, error) {
	return r.backend.GetTemplate(id)
}

// CreateTemplate validates and persists an template.
func (r repository) CreateTemplate(template *model.Template) (*model.Template, error) {
	finalID, idErr := r.checkIDErrOnCreate(template)
	if idErr != nil {
		return template, idErr
	}
	if r.externalIDs {
		template.Id = template.Details.Id
	}
	if errs := template.Validate(); len(errs) > 0 {
		err := newValError(errs)
		return template, err
	}
	template.Id = finalID
	return r.backend.CreateTemplate(template)
}

func normalizeState(s model.State) model.State {
	for _, v := range model.States {
		if s == v {
			return s
		}
	}
	return model.STOPPED
}

func (r repository) checkIDErrOnCreate(e model.Identity) (string, error) {
	var finalID string
	if r.externalIDs {
		finalID = ""
		if e.GetId() != "" {
			return finalID, fmt.Errorf("Entity %T[id='%s'] must have empty Id on create", e, e.GetId())
		}
	} else {
		finalID = e.GetId()
	}
	return finalID, nil
}

func (r repository) checkIDErrOnUpdate(e model.Identity) error {
	if r.externalIDs && e.GetId() == "" {
		return fmt.Errorf("Entity %T[id='%s'] must have non empty Id on update", e, e.GetId())
	}
	return nil
}
