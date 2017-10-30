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

	"github.com/simplereach/timeutils"
)

var ErrNotFound error = errors.New("Entity not found")
var ErrAlreadyExist error = errors.New("Entity already exists")

type Provider struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type Providers []Provider
type Agreements []Agreement

type Assessment struct {
	Id             string    `json:"id"`
	Enabled        bool      `json:"enabled"`
	FirstExecution time.Time `json:"first_execution"`
	LastExecution  time.Time `json:"last_execution"`
}

type Penalty struct {
	Type  string `json:type`
	Value string `json:value`
	Unit  string `json:unit`
}

type Guarantee struct {
	Name       string    `json:name`
	Constraint string    `json:constraints`
	Warning    string    `json:warning`
	Penalties  []Penalty `json:penalties`
}

type Agreement struct {
	Type       string         `json:type`
	Id         string         `json:id`
	Name       string         `json:name`
	Provider   Provider       `json:provider`
	Client     Provider       `json:client`
	Creation   timeutils.Time `json:creation`
	Expiration timeutils.Time `json:expiration`
	Guarantees []Guarantee    `json:guarantees`
}
