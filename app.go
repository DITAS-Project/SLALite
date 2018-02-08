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
	"github.com/spf13/viper"
)

const (
	defaultPort        string = "8090"
	defaultEnableSsl   bool   = false
	defaultSslCertPath string = "cert.pem"
	defaultSslKeyPath  string = "key.pem"

	portPropertyName        = "port"
	enableSslPropertyName   = "enableSsl"
	sslCertPathPropertyName = "sslCertPath"
	sslKeyPathPropertyName  = "sslKeyPath"
)

// App is a main application "object", to be built by main and testmain
type App struct {
	Router      *mux.Router
	Repository  model.IRepository
	Port        string
	SslEnabled  bool
	SslCertPath string
	SslKeyPath  string
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
	"providers":  endpoint{"GET", "/providers", "Providers"},
	"agreements": endpoint{"GET", "/agreements", "Agreements"},
}

func NewApp(config *viper.Viper, repository model.IRepository) (App, error) {

	setDefaults(config)
	logConfig(config)

	a := App{
		Port:        config.GetString(portPropertyName),
		SslEnabled:  config.GetBool(enableSslPropertyName),
		SslCertPath: config.GetString(sslCertPathPropertyName),
		SslKeyPath:  config.GetString(sslKeyPathPropertyName),
	}

	a.initialize(repository)
	/*
	 * TODO Return error if files not found, for ex.
	 */
	return a, nil
}

func setDefaults(config *viper.Viper) {
	config.SetDefault(portPropertyName, defaultPort)
	config.SetDefault(sslCertPathPropertyName, defaultSslCertPath)
	config.SetDefault(sslKeyPathPropertyName, defaultSslKeyPath)
}

func logConfig(config *viper.Viper) {
	ssl := "no"
	if config.GetBool(enableSslPropertyName) == true {
		ssl = fmt.Sprintf(
			"(cert='%s', key='%s')",
			config.GetString(sslCertPathPropertyName),
			config.GetString(sslKeyPathPropertyName))
	}

	log.Printf("HTTP/S configuration\n"+
		"\tport: %v\n"+
		"\tssl: %v\n",
		config.GetString(portPropertyName),
		ssl)
}

// Initialize initializes the REST API passing the db connection
func (a *App) initialize(repository model.IRepository) {

	a.Repository = repository

	a.Router = mux.NewRouter().StrictSlash(true)

	a.Router.HandleFunc("/", a.Index).Methods("GET")

	a.Router.Methods("GET").Path("/providers").
		Handler(LoggerDecorator(http.HandlerFunc(a.GetAllProviders), "All Providers"))
	a.Router.HandleFunc("/providers/{id}", a.GetProvider).Methods("GET")
	a.Router.HandleFunc("/providers", a.CreateProvider).Methods("POST")
	a.Router.HandleFunc("/providers/{id}", a.DeleteProvider).Methods("DELETE")

	a.Router.Methods("GET").Path("/agreements").
		Handler(LoggerDecorator(http.HandlerFunc(a.GetAgreements), "Agreements"))
	a.Router.HandleFunc("/agreements/{id}", a.GetAgreement).Methods("GET")
	a.Router.HandleFunc("/agreements", a.CreateAgreement).Methods("POST")
	a.Router.HandleFunc("/agreements/{id}/start", a.StartAgreement).Methods("PUT")
	a.Router.HandleFunc("/agreements/{id}/stop", a.StopAgreement).Methods("PUT")
	a.Router.HandleFunc("/agreements/{id}", a.DeleteAgreement).Methods("DELETE")
}

// Run starts the REST API
func (a *App) Run() {
	addr := ":" + a.Port

	if a.SslEnabled {
		log.Fatal(http.ListenAndServeTLS(addr, a.SslCertPath, a.SslKeyPath, a.Router))
	} else {
		log.Fatal(http.ListenAndServe(addr, a.Router))
	}
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

func (a *App) getAll(w http.ResponseWriter, r *http.Request, f func() (interface{}, error)) {
	list, err := f()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
	} else {
		respondSuccessJSON(w, list)
	}
}

func (a *App) get(w http.ResponseWriter, r *http.Request, f func(string) (interface{}, error)) {
	vars := mux.Vars(r)
	id := vars["id"]

	provider, err := f(id)
	if err != nil {
		manageError(err, w)
	} else {
		respondSuccessJSON(w, provider)
	}
}

func (a *App) create(w http.ResponseWriter, r *http.Request, decode func() error, create func() (model.Identity, error)) {

	errDec := decode()
	if errDec != nil {
		respondWithError(w, http.StatusBadRequest, errDec.Error())
		return
	}
	/* check errors */
	created, err := create()
	if err != nil {
		manageError(err, w)
	} else {
		respondWithJSON(w, http.StatusCreated, created)
	}
}

func (a *App) update(w http.ResponseWriter, r *http.Request, upd func(string) error) {
	vars := mux.Vars(r)
	id := vars["id"]

	err := upd(id)

	if err != nil {
		manageError(err, w)
	} else {
		respondNoContent(w)
	}
}

// GetAllProviders return all providers in db
func (a *App) GetAllProviders(w http.ResponseWriter, r *http.Request) {
	a.getAll(w, r, func() (interface{}, error) {
		return a.Repository.GetAllProviders()
	})
}

// GetProvider gets a provider by REST ID
func (a *App) GetProvider(w http.ResponseWriter, r *http.Request) {
	a.get(w, r, func(id string) (interface{}, error) {
		return a.Repository.GetProvider(id)
	})
}

// CreateProvider creates a provider passed by REST params
func (a *App) CreateProvider(w http.ResponseWriter, r *http.Request) {

	var provider model.Provider

	a.create(w, r,
		func() error {
			return json.NewDecoder(r.Body).Decode(&provider)
		},
		func() (model.Identity, error) {
			return a.Repository.CreateProvider(&provider)
		})
}

// DeleteProvider deletes /provider/id
func (a *App) DeleteProvider(w http.ResponseWriter, r *http.Request) {
	a.update(w, r, func(id string) error {
		return a.Repository.DeleteProvider(&model.Provider{Id: id})
	})
}

// GetAllAgreements return all agreements in db
func (a *App) GetAgreements(w http.ResponseWriter, r *http.Request) {
	v := r.URL.Query()
	active := v.Get("active")

	a.getAll(w, r, func() (interface{}, error) {
		if active != "" {
			return a.Repository.GetActiveAgreements()
		}
		return a.Repository.GetAllAgreements()
	})
}

// GetAgreement gets an agreement by REST ID
func (a *App) GetAgreement(w http.ResponseWriter, r *http.Request) {
	a.get(w, r, func(id string) (interface{}, error) {
		return a.Repository.GetAgreement(id)
	})
}

// CreateAgreement creates a agreement passed by REST params
func (a *App) CreateAgreement(w http.ResponseWriter, r *http.Request) {

	var agreement model.Agreement

	a.create(w, r,
		func() error {
			return json.NewDecoder(r.Body).Decode(&agreement)
		},
		func() (model.Identity, error) {
			return a.Repository.CreateAgreement(&agreement)
		})
}

// DeleteAgreement deletes an agreement by id
func (a *App) DeleteAgreement(w http.ResponseWriter, r *http.Request) {
	a.update(w, r, func(id string) error {
		return a.Repository.DeleteAgreement(&model.Agreement{Id: id})
	})
}

// StartAgreement starts monitoring an agreement
func (a *App) StartAgreement(w http.ResponseWriter, r *http.Request) {
	a.update(w, r, func(id string) error {
		return a.Repository.StartAgreement(id)
	})
}

// StopAgreement stop monitoring an agreement
func (a *App) StopAgreement(w http.ResponseWriter, r *http.Request) {
	a.update(w, r, func(id string) error {
		return a.Repository.StopAgreement(id)
	})
}

func manageError(err error, w http.ResponseWriter) {
	switch err {
	case model.ErrAlreadyExist:
		respondWithError(w, http.StatusConflict, "Object already exist")
	case model.ErrNotFound:
		respondWithError(w, http.StatusNotFound, "Can't find object")
	default:
		respondWithError(w, http.StatusInternalServerError, err.Error())
	}
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, ApiError{strconv.Itoa(code), message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	enc.Encode(payload)
	
}

func respondSuccessJSON(w http.ResponseWriter, payload interface{}) {
	respondWithJSON(w, http.StatusOK, payload)
}

func respondNoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}
