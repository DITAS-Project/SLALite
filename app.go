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
package main

import (
	"SLALite/model"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

// App is a main application "object", to be built by main and testmain
type App struct {
	Router     *mux.Router
	Repository model.IRepository
}

// ApiError is the struct sent to client on errors
type ApiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *ApiError) Error() string {
	return e.Message
}

type endpoint struct {
	Method string
	Path   string
	Help   string
}

var api = map[string]endpoint{
	"providers": endpoint{"GET", "/providers", "Providers"},
}

// Initialize initializes the REST API passing the db connection
func (a *App) Initialize(repository model.IRepository) {

	a.Repository = repository

	a.Router = mux.NewRouter().StrictSlash(true)

	a.Router.HandleFunc("/", a.Index).Methods("GET")

	a.Router.Methods("GET").Path("/providers").Handler(
		LoggerDecorator(http.HandlerFunc(a.GetAllProviders), "All Providers"))
	a.Router.HandleFunc("/providers/{id}", a.GetProvider).Methods("GET")
	a.Router.HandleFunc("/providers", a.CreateProvider).Methods("POST")
	a.Router.HandleFunc("/providers/{id}", a.DeleteProvider).Methods("DELETE")
}

// Run starts the REST API
func (a *App) Run(addr string) {
	log.Fatal(http.ListenAndServe(addr, a.Router))
}

// Index is the API index
func (a *App) Index(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(api)
}

func LoggerDecorator(inner http.Handler, name string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		inner.ServeHTTP(w, r)

		log.Printf(
			"%s\t%s\t%s\t%s",
			r.Method,
			r.RequestURI,
			name,
			time.Since(start),
		)
	})
}

// GetAllProviders return all providers in db
func (a *App) GetAllProviders(w http.ResponseWriter, r *http.Request) {
	providers, err := a.Repository.GetAllProviders()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
	} else {
		respondSuccessJSON(w, providers)
	}
}

// GetProvider gets a provider by REST ID
func (a *App) GetProvider(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	provider, err := a.Repository.GetProvider(id)
	if err != nil {
		switch err {
		case model.ErrNotFound:
			respondWithError(w, http.StatusNotFound, fmt.Sprintf("Provider{id:%s} not found", id))
		default:
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	} else {
		respondSuccessJSON(w, provider)
	}
}

// CreateProvider creates a provider passed by REST params
func (a *App) CreateProvider(w http.ResponseWriter, r *http.Request) {

	var provider model.Provider
	errDec := json.NewDecoder(r.Body).Decode(&provider)
	if errDec != nil {
		respondWithError(w, http.StatusBadRequest, errDec.Error())
	}
	/* check errors */
	created, err := a.Repository.CreateProvider(&provider)
	if err != nil {
		switch err {
		case model.ErrAlreadyExist:
			respondWithError(w, http.StatusConflict,
				fmt.Sprintf("Provider{id: %s} already exists", provider.Id))
		default:
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
	} else {
		respondWithJSON(w, http.StatusCreated, created)
	}
}

// DeleteProvider deletes /provider/id
func (a *App) DeleteProvider(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	p := model.Provider{Id: id}
	err := a.Repository.DeleteProvider(&p)

	if err != nil {
		switch err {
		case model.ErrNotFound:
			respondWithError(w, http.StatusNotFound,
				fmt.Sprintf("Provider{id: %s} not found", p.Id))
		default:
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
	} else {
		respondNoContent(w)
	}
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, ApiError{strconv.Itoa(code), message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}

func respondSuccessJSON(w http.ResponseWriter, payload interface{}) {
	respondWithJSON(w, http.StatusOK, payload)
}

func respondNoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}
