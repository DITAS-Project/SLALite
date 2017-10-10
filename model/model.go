package model

import (
    "time"
)

type Provider struct {
    Id string           `json:"id"`
    Name string         `json:"name"`
}

type Providers []Provider

type Agreement struct {
    Id string
    Text string
}

type Assessment struct {
    Id string                   `json:"id"`
    Enabled bool                `json:"enabled"`
    FirstExecution time.Time    `json:"first_execution"`
    LastExecution  time.Time    `json:"last_execution"`
}