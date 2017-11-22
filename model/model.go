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

import (
	"errors"
	"fmt"
	"time"
)

var ErrNotFound error = errors.New("Entity not found")
var ErrAlreadyExist error = errors.New("Entity already exists")

type Identity interface {
	GetId() string
}

type Validable interface {
	Validate() []error
}

type State string
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

// Provider is the entity that represents a service provider
type Provider struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

func (p *Provider) GetId() string {
	return p.Id
}

func (p *Provider) Validate() []error {
	result := make([]error, 0, 10)

	result = checkEmpty(p.Id, "Provider.Id", result)
	result = checkEmpty(p.Name, "Provider.Name", result)

	return result
}

// Agreement is the entity that represents an agreement between a provider and a client.
// The Text is ReadOnly in normal conditions, with the exception of a renegotiation.
// The Assessment cannot be modified externally.
// The Signature is the Text digitally signed by the Client (not used yet)
type Agreement struct {
	Id         string        `json:"id"`
	Name       string        `json:"name"`
	State      State         `json:"state"`
	Assessment Assessment    `json:"assessment"`
	Text       AgreementText `json:"text"`
	/* Signature string `json:"signature"` */
}

// Assessment is the struct that provides assessment information
type Assessment struct {
	FirstExecution time.Time `json:"first_execution"`
	LastExecution  time.Time `json:"last_execution"`
}

// AgreementText is the struct that represents the "contract" signed by the client
type AgreementText struct {
	Id         string      `json:"id"`
	Type       TextType    `json:"type"`
	Name       string      `json:"name"`
	Provider   Provider    `json:"provider"`
	Client     Provider    `json:"client"`
	Creation   time.Time   `json:"creation"`
	Expiration time.Time   `json:"expiration"`
	Guarantees []Guarantee `json:"guarantees"`
}

// Guarantee is the struct that represents an SLO
type Guarantee struct {
	Name       string    `json:"name"`
	Constraint string    `json:"constraints"`
	Warning    string    `json:"warning"`
	Penalties  []Penalty `json:"penalties"`
}

// Penalty is the struct that represents a penalty in case of an SLO violation
type Penalty struct {
	Type  string `json:"type"`
	Value string `json:"value"`
	Unit  string `json:"unit"`
}

func (a *Agreement) GetId() string {
	return a.Id
}

func (a *Agreement) IsStarted() bool {
	return a.State == STARTED
}

func (a *Agreement) IsTerminated() bool {
	return a.State == TERMINATED
}

func (a *Agreement) IsStopped() bool {
	return a.State == STOPPED
}

func (a *Agreement) Validate() []error {
	result := make([]error, 0)

	a.State = normalizeState(a.State)
	result = checkEmpty(a.Id, "Agreement.Id", result)
	result = checkEmpty(a.Name, "Agreement.Name", result)
	for _, e := range a.Assessment.Validate() {
		result = append(result, e)
	}
	for _, e := range a.Text.Validate() {
		result = append(result, e)
	}

	result = checkEquals(a.Id, "Agreement.Id", a.Text.Id, "Agreement.Text.Id", result)
	result = checkEquals(a.Name, "Agreement.Name", a.Text.Name, "Agreement.Text.Name", result)

	return result
}

func (as *Assessment) Validate() []error {
	return []error{}
}

func (t *AgreementText) Validate() []error {
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

func (g *Guarantee) Validate() []error {
	result := make([]error, 0)
	result = checkEmpty(g.Name, "Guarantee.Name", result)
	result = checkEmpty(g.Constraint, fmt.Sprintf("Guarantee['%s']", g.Name), result)

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

type Providers []Provider
type Agreements []Agreement
