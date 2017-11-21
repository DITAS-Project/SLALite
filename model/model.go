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
	"time"
)

var ErrNotFound error = errors.New("Entity not found")
var ErrAlreadyExist error = errors.New("Entity already exists")

type Identity interface {
	GetId() string
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

	// AGREEMENT is the text type of an Agreement text
	AGREEMENT TextType = "agreement"

	// TEMPLATE is the text type of a Template text
	TEMPLATE TextType = "template"
)

// Provider is the entity that represents a service provider
type Provider struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

func (p Provider) GetId() string {
	return p.Id
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

func (a Agreement) GetId() string {
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

type Providers []Provider
type Agreements []Agreement
