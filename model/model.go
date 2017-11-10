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

type Provider struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

func (p Provider) GetId() string {
	return p.Id
}

type Assessment struct {
	Id             string    `json:"id"`
	Enabled        bool      `json:"enabled"`
	FirstExecution time.Time `json:"first_execution"`
	LastExecution  time.Time `json:"last_execution"`
}

type Penalty struct {
	Type  string `json:"type"`
	Value string `json:"value"`
	Unit  string `json:"unit"`
}

type Guarantee struct {
	Name       string    `json:"name"`
	Constraint string    `json:"constraints"`
	Warning    string    `json:"warning"`
	Penalties  []Penalty `json:"penalties"`
}

type Agreement struct {
	Type       string      `json:"type"`
	Id         string      `json:"id"`
	Name       string      `json:"name"`
	Active     bool        `json:"active"`
	Provider   Provider    `json:"provider"`
	Client     Provider    `json:"client"`
	Creation   time.Time   `json:"creation"`
	Expiration time.Time   `json:"expiration"`
	Guarantees []Guarantee `json:"guarantees"`
}

func (a Agreement) GetId() string {
	return a.Id
}

type Providers []Provider
type Agreements []Agreement
