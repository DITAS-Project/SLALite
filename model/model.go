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
Package model contain the entities used in the SLALite: agreements, violations, penalties...
It also defines the interface IRepository, which defines the operations to be implemented
by any repository.
*/
package model

import (
	"errors"
	"fmt"
	"time"
)

//
// ErrNotFound is the sentinel error for an entity not found
//
var ErrNotFound = errors.New("Entity not found")

//
// ErrAlreadyExist is the sentinel error for creating an entity whose id already exists
//
var ErrAlreadyExist = errors.New("Entity already exists")

/*
 * ValidationErrors following behavioral errors
 * (https://dave.cheney.net/2016/04/27/dont-just-check-errors-handle-them-gracefully)
 */

// validationError is an interface that must be implemented by custom error implementations
// swagger:ignore
type validationError interface {
	IsErrValidation() bool
}

//
// IsErrValidation return true is an error is a validation error
//
func IsErrValidation(err error) bool {
	v, ok := err.(validationError)
	return ok && v.IsErrValidation()
}

// func IsErrNotFound(err error) bool

//
// Identity identifies entities with an Id field
//
type Identity interface {
	GetId() string
}

//
// Validable identifies entities that can be validated
//
type Validable interface {
	Validate(val Validator, mode ValidationMode) []error
}

// State is the type of possible states of an agreement
type State string

// TextType is the type of possible types a Details type
type TextType string

const (
	// STARTED is the state of an agreement that can be evaluated
	STARTED State = "started"

	// STOPPED is the state of an agreement temporaryly not evaluated
	STOPPED State = "stopped"

	// TERMINATED is the final state of an agreement
	TERMINATED State = "terminated"
)

const (
	// AGREEMENT is the text type of an Agreement text
	AGREEMENT TextType = "agreement"

	// TEMPLATE is the text type of a Template text
	TEMPLATE TextType = "template"
)

// States is the list of possible states of an agreement/template
var States = [...]State{STOPPED, STARTED, TERMINATED}

// Party is the entity that represents a service provider or a client
// swagger:model
type Party struct {
	Id   string `json:"id" bson:"_id"`
	Name string `json:"name"`
}

// Provider is the entity that represents a Provider
// swagger:model
type Provider Party

// GetId returns the Id of a provider
func (p *Provider) GetId() string {
	return p.Id
}

// Validate validates the consistency of a Provider entity
func (p *Provider) Validate(val Validator, mode ValidationMode) []error {
	return val.ValidateProvider(p, mode)
}

// Client is the entity that represents a client.
// swagger:model
type Client Party

// GetId returns the Id of a client
func (c *Client) GetId() string {
	return c.Id
}

// Validate validates the consistency of a Client entity
func (c *Client) Validate(val Validator, mode ValidationMode) []error {
	return val.ValidateClient(c, mode)
}

// Template is the entity that serves as base to create new agreements
//
// The Details field of the template contains placeholders that are substituted
// when generating an agreement from a template (see generator package).
// The Constraints fields contains constraints that a variable used in a guarantee
// must satisfy. F.e., if the guarantee expression is "cpu_usage < {{M}}", one could
// specify in Constraints that "M" : "M >= 0 && M <= 100".Template
//
// The Id and Name are relative to the template itself, and should not match
// the fields in Details.
type Template struct {
	Id   string `json:"id" bson:"_id"`
	Name string `json:"name"`
	//	State       State             `json:"state"`
	Details     Details           `json:"details"`
	Constraints map[string]string `json:"constraints"`
}

// Agreement is the entity that represents an agreement between a provider and a client.
// The Text is ReadOnly in normal conditions, with the exception of a renegotiation.
// The Assessment cannot be modified externally.
// The Signature is the Text digitally signed by the Client (not used yet)
// swagger:model
type Agreement struct {
	Id         string     `json:"id" bson:"_id"`
	Name       string     `json:"name"`
	State      State      `json:"state"`
	Assessment Assessment `json:"assessment"`
	Details    Details    `json:"details"`

	/* Signature string `json:"signature"` */
}

// Assessment is the struct that provides assessment information
// swagger:model
type Assessment struct {
	FirstExecution time.Time `json:"first_execution"`
	LastExecution  time.Time `json:"last_execution"`
}

// Details is the struct that represents the "contract" signed by the client
// swagger:model
type Details struct {
	Id         string      `json:"id"`
	Type       TextType    `json:"type"`
	Name       string      `json:"name"`
	Provider   Provider    `json:"provider"`
	Client     Client      `json:"client"`
	Creation   time.Time   `json:"creation"`
	Expiration *time.Time  `json:"expiration,omitempty"`
	Guarantees []Guarantee `json:"guarantees"`
}

// Guarantee is the struct that represents an SLO
// swagger:model
type Guarantee struct {
	Name       string       `json:"name"`
	Constraint string       `json:"constraint"`
	Warning    string       `json:"warning"`
	Penalties  []PenaltyDef `json:"penalties"`
}

// PenaltyDef is the struct that represents a penalty in case of an SLO violation
// swagger:model
type PenaltyDef struct {
	Type  string `json:"type"`
	Value string `json:"value"`
	Unit  string `json:"unit"`
}

// MetricValue is the SLALite representation of a metric value.
// swagger:model
type MetricValue struct {
	Key      string      `json:"key"`
	Value    interface{} `json:"value"`
	DateTime time.Time   `json:"datetime"`
}

func (v MetricValue) String() string {
	return fmt.Sprintf("{Key: %s, Value: %v, DateTime: %v}", v.Key, v.Value, v.DateTime)
}

// Violation is generated when a guarantee term is not fulfilled
// swagger:model
type Violation struct {
	Id          string        `json:"id"`
	AgreementId string        `json:"agreement_id"`
	Guarantee   string        `json:"guarantee"`
	Datetime    time.Time     `json:"datetime"`
	Constraint  string        `json:"constraint"`
	Values      []MetricValue `json:"values"`
}

// Penalty is generated when a guarantee term is violated is the term has
// PenaltyDefs associated.
// swagger:model
type Penalty struct {
	Id          string     `json:"id"`
	AgreementId string     `json:"agreement_id"`
	Guarantee   string     `json:"guarantee"`
	Datetime    time.Time  `json:"datetime"`
	Definition  PenaltyDef `json:"definition"`
}

// GetId returns the id of an template
func (t *Template) GetId() string {
	return t.Id
}

// Validate validates the consistency of a Template.
func (t *Template) Validate(val Validator, mode ValidationMode) []error {
	return val.ValidateTemplate(t, mode)
}

// GetId returns the id of an agreement
func (a *Agreement) GetId() string {
	return a.Id
}

// IsStarted is true if the agreement state is STARTED
func (a *Agreement) IsStarted() bool {
	return a.State == STARTED
}

// IsTerminated is true if the agreement state is TERMINATED
func (a *Agreement) IsTerminated() bool {
	return a.State == TERMINATED
}

// IsStopped is true if the agreement state is STOPPED
func (a *Agreement) IsStopped() bool {
	return a.State == STOPPED
}

// IsValidTransition returns if the transition to newState is valid
func (a *Agreement) IsValidTransition(newState State) bool {
	return a.State != TERMINATED
}

// Validate validates the consistency of an Agreement.
func (a *Agreement) Validate(val Validator, mode ValidationMode) []error {
	return val.ValidateAgreement(a, mode)
}

// Validate validates the consistency of an Assessment entity
func (as *Assessment) Validate(val Validator, mode ValidationMode) []error {
	return val.ValidateAssessment(as, mode)
}

// Validate validates the consistency of a Details entity
func (t *Details) Validate(val Validator, mode ValidationMode) []error {
	return val.ValidateDetails(t, mode)
}

// Validate validates the consistency of a Guarantee entity
func (g *Guarantee) Validate(val Validator, mode ValidationMode) []error {
	return val.ValidateGuarantee(g, mode)
}

// GetId returns the Id of a violation
func (v *Violation) GetId() string {
	return v.Id
}

// Validate validates the consistency of a Violation entity
func (v *Violation) Validate(val Validator, mode ValidationMode) []error {
	return val.ValidateViolation(v, mode)
}

// Normalize returns an always valid state: any different value from contained in States is STOPPED.
func (s State) Normalize() State {
	return normalizeState(s)
}

// Providers is the type of an slice of Provider
// swagger:model
type Providers []Provider

// Agreements is the type of an slice of Agreement
// swagger:model
type Agreements []Agreement

// Templates is the type of an slice of Template
// swagger:model
type Templates []Template
