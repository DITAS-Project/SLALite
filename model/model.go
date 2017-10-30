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
