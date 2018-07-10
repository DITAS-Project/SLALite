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

//
// Validation errors must implement this interface
//
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
	Validate() []error
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
type Party struct {
	Id   string `json:"id" bson:"_id"`
	Name string `json:"name"`
}

// Provider is the entity that represents a Provider
type Provider Party

// GetId returns the Id of a provider
func (p *Provider) GetId() string {
	return p.Id
}

// Validate validates the consistency of a Provider entity
func (p *Provider) Validate() []error {
	result := make([]error, 0, 2)

	result = checkEmpty(p.Id, "Provider.Id", result)
	result = checkEmpty(p.Name, "Provider.Name", result)

	return result
}

// Client is the entity that represents a client.
type Client Party

// GetId returns the Id of a client
func (c *Client) GetId() string {
	return c.Id
}

// Validate validates the consistency of a Client entity
func (c *Client) Validate() []error {
	result := make([]error, 0, 2)

	result = checkEmpty(c.Id, "Client.Id", result)
	result = checkEmpty(c.Name, "Client.Name", result)

	return result
}

// Agreement is the entity that represents an agreement between a provider and a client.
// The Text is ReadOnly in normal conditions, with the exception of a renegotiation.
// The Assessment cannot be modified externally.
// The Signature is the Text digitally signed by the Client (not used yet)
type Agreement struct {
	Id         string     `json:"id" bson:"_id"`
	Name       string     `json:"name"`
	State      State      `json:"state"`
	Assessment Assessment `json:"assessment"`
	Details    Details    `json:"details"`

	/* Signature string `json:"signature"` */
}

// Assessment is the struct that provides assessment information
type Assessment struct {
	FirstExecution time.Time `json:"first_execution"`
	LastExecution  time.Time `json:"last_execution"`
}

// Details is the struct that represents the "contract" signed by the client
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
type Guarantee struct {
	Name       string       `json:"name"`
	Constraint string       `json:"constraint"`
	Warning    string       `json:"warning"`
	Penalties  []PenaltyDef `json:"penalties"`
}

// PenaltyDef is the struct that represents a penalty in case of an SLO violation
type PenaltyDef struct {
	Type  string `json:"type"`
	Value string `json:"value"`
	Unit  string `json:"unit"`
}

// MetricValue is the SLALite representation of a metric value.
type MetricValue struct {
	Key      string      `json:"key"`
	Value    interface{} `json:"value"`
	DateTime time.Time   `json:"datetime"`
}

func (v *MetricValue) String() string {
	return fmt.Sprintf("{Key: %s, Value: %v, DateTime: %v}", v.Key, v.Value, v.DateTime)
}

// Violation is generated when a guarantee term is not fulfilled
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
type Penalty struct {
	Id          string     `json:"id"`
	AgreementId string     `json:"agreement_id"`
	Guarantee   string     `json:"guarantee"`
	Datetime    time.Time  `json:"datetime"`
	Definition  PenaltyDef `json:"definition"`
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

// Validate validates the consistency of an Agreement.
func (a *Agreement) Validate() []error {
	result := make([]error, 0)

	a.State = normalizeState(a.State)
	result = checkEmpty(a.Id, "Agreement.Id", result)
	result = checkEmpty(a.Name, "Agreement.Name", result)
	for _, e := range a.Assessment.Validate() {
		result = append(result, e)
	}
	for _, e := range a.Details.Validate() {
		result = append(result, e)
	}

	result = checkEquals(a.Id, "Agreement.Id", a.Details.Id, "Agreement.Details.Id", result)
	result = checkEquals(a.Name, "Agreement.Name", a.Details.Name, "Agreement.Details.Name", result)

	return result
}

// Validate validates the consistency of an Assessment entity
func (as *Assessment) Validate() []error {
	return []error{}
}

// Validate validates the consistency of a Details entity
func (t *Details) Validate() []error {
	result := make([]error, 0)
	result = checkEmpty(t.Id, "Text.Id", result)
	result = checkEmpty(t.Name, "Text.Name", result)
	for _, e := range t.Provider.Validate() {
		result = append(result, e)
	}
	for _, e := range t.Client.Validate() {
		result = append(result, e)
	}
	for _, g := range t.Guarantees {
		for _, e := range g.Validate() {
			result = append(result, e)
		}
	}
	return result
}

// Validate validates the consistency of a Guarantee entity
func (g *Guarantee) Validate() []error {
	result := make([]error, 0)
	result = checkEmpty(g.Name, "Guarantee.Name", result)
	result = checkEmpty(g.Constraint, fmt.Sprintf("Guarantee['%s'].Constraint", g.Name), result)

	return result
}

// GetId returns the Id of a violation
func (v *Violation) GetId() string {
	return v.Id
}

// Validate validates the consistency of a Violation entity
func (v *Violation) Validate() []error {
	result := make([]error, 0)
	result = checkEmpty(v.Id, "Violation.Id", result)
	result = checkEmpty(v.AgreementId, "Violation.AgreementId", result)
	result = checkEmpty(v.Guarantee, "Violation.Guarantee", result)
	if v.Datetime.IsZero() {
		result = append(result, fmt.Errorf("%v is not a valid date", v.Datetime))
	}
	if v.Values == nil || len(v.Values) == 0 {
		result = append(result, fmt.Errorf("Violation.Values cannot be empty"))
	}
	result = checkEmpty(v.Constraint, "Violation.Constraint", result)

	return result
}

func checkEmpty(field string, description string, current []error) []error {
	if field == "" {
		current = append(current, fmt.Errorf("%s is empty", description))
	}
	return current
}

func checkEquals(f1 string, f1desc, f2 string, f2desc string, current []error) []error {
	if f1 != f2 {
		current = append(current, fmt.Errorf("%s and %s do not match", f1desc, f2desc))
	}
	return current
}

func normalizeState(s State) State {
	for _, v := range States {
		if s == v {
			return s
		}
	}
	return STOPPED
}

// Providers is the type of an slice of Provider
type Providers []Provider

// Agreements is the type of an slice of Agreement
type Agreements []Agreement
