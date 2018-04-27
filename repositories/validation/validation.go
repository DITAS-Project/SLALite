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
)

type repository struct {
	backend model.IRepository
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
func New(backend model.IRepository) (model.IRepository, error) {
	return repository{
		backend: backend,
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
	if errs := provider.Validate(); len(errs) > 0 {
		err := newValError(errs)
		return provider, err
	}
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

// GetActiveAgreements gets all active agreements.
func (r repository) GetActiveAgreements() (model.Agreements, error) {
	return r.backend.GetActiveAgreements()
}

// CreateAgreement validates and persists an agreement.
func (r repository) CreateAgreement(agreement *model.Agreement) (*model.Agreement, error) {
	if errs := agreement.Validate(); len(errs) > 0 {
		err := newValError(errs)
		return agreement, err
	}
	return r.backend.CreateAgreement(agreement)
}

// UpdateAgreement validates and updates an agreement.
func (r repository) UpdateAgreement(agreement *model.Agreement) (*model.Agreement, error) {
	if errs := agreement.Validate(); len(errs) > 0 {
		err := newValError(errs)
		return agreement, err
	}
	return r.backend.UpdateAgreement(agreement)
}

// DeleteAgreement deletes an agreement from repository.
func (r repository) DeleteAgreement(agreement *model.Agreement) error {
	return r.backend.DeleteAgreement(agreement)
}

// StartAgreement changes an agreement state to started.
func (r repository) StartAgreement(id string) error {
	return r.backend.StartAgreement(id)
}

// StopAgreement changes an agreement state to stopped.
func (r repository) StopAgreement(id string) error {
	return r.backend.StopAgreement(id)
}

// CreateViolation validates and persists a new Violation.
func (r repository) CreateViolation(v *model.Violation) (*model.Violation, error) {
	if errs := v.Validate(); len(errs) > 0 {
		err := newValError(errs)
		return v, err
	}
	return r.backend.CreateViolation(v)
}

// GetViolation returns the Violation identified by id.
func (r repository) GetViolation(id string) (*model.Violation, error) {
	return r.backend.GetViolation(id)
}
