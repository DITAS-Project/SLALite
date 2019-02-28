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
	"net/http"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"

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
// swagger:ignore
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

// endpoint represents an available operation represented by its HTTP method, the expected path for invocations and an optional help message.
// swagger:model
type endpoint struct {
	// example: GET
	Method string

	// example: /providers
	Path string

	// example: Gets a list of registered providers
	Help string
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

	a.Router.Methods("GET").Path("/providers").Handler(logger(a.GetAllProviders))

	a.Router.Methods("GET").Path("/providers/{id}").Handler(logger(a.GetProvider))
	a.Router.Methods("POST").Path("/providers").Handler(logger(a.CreateProvider))
	a.Router.Methods("DELETE").Path("/providers/{id}").Handler(logger(a.DeleteProvider))

	a.Router.Methods("GET").Path("/agreements").Handler(logger(a.GetAgreements))
	a.Router.Methods("GET").Path("/agreements/{id}").Handler(logger(a.GetAgreement))
	a.Router.Methods("POST").Path("/agreements").Handler(logger(a.CreateAgreement))
	a.Router.Methods("PUT").Path("/agreements/{id}/start").Handler(logger(a.StartAgreement))
	a.Router.Methods("PUT").Path("/agreements/{id}/stop").Handler(logger(a.StopAgreement))
	a.Router.Methods("PUT").Path("/agreements/{id}/terminate").Handler(logger(a.TerminateAgreement))
	a.Router.Methods("PUT").Path("/agreements/{id}").Handler(logger(a.UpdateAgreement))
	a.Router.Methods("DELETE").Path("/agreements/{id}").Handler(logger(a.DeleteAgreement))
	a.Router.Methods("GET").Path("/agreements/{id}/details").Handler(logger(a.GetAgreementDetails))

	a.Router.Methods("GET").Path("/templates").Handler(logger(a.GetTemplates))
	a.Router.Methods("GET").Path("/templates/{id}").Handler(logger(a.GetTemplate))
	a.Router.Methods("POST").Path("/templates").Handler(logger(a.CreateTemplate))
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
// swagger:operation GET / index
//
// Returns the available operations
//
// ---
// produces:
// - application/json
// responses:
//   '200':
//     description: API description
//     schema:
//       type: object
//       additionalProperties:
//         "$ref": "#/definitions/endpoint"
func (a *App) Index(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(api)
}

func logger(f func(w http.ResponseWriter, r *http.Request)) http.Handler {
	return loggerDecorator(http.HandlerFunc(f))
}

func loggerDecorator(inner http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		inner.ServeHTTP(w, r)

		log.Printf(
			"%s\t%s\t\t%s",
			r.Method,
			r.RequestURI,
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

// Update operation where the resource is updated with the body passed in request
func (a *App) updateEntity(w http.ResponseWriter, r *http.Request, decode func() error, update func(id string) (model.Identity, error)) {
	vars := mux.Vars(r)
	id := vars["id"]

	errDec := decode()
	if errDec != nil {
		respondWithError(w, http.StatusBadRequest, errDec.Error())
		return
	}
	/* check errors */
	updated, err := update(id)
	if err != nil {
		manageError(err, w)
	} else {
		respondWithJSON(w, http.StatusOK, updated)
	}
}

// Any other update operation not covered by updateEntity (e.g., delete)
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
// swagger:operation GET /providers getAllProviders
//
// Returns all registered providers
//
// ---
// produces:
// - application/json
// responses:
//   '200':
//     description: The complete list of registered providers
//     schema:
//       type: object
//       additionalProperties:
//         "$ref": "#/definitions/Providers"
func (a *App) GetAllProviders(w http.ResponseWriter, r *http.Request) {
	a.getAll(w, r, func() (interface{}, error) {
		return a.Repository.GetAllProviders()
	})
}

// GetProvider gets a provider by REST ID
// swagger:operation GET /providers/{id} getProvider
//
// Returns a provider given its ID
//
// ---
// produces:
// - application/json
// parameters:
// - name: id
//   in: path
//   description: The identifier of the provider
//   required: true
//   type: string
// responses:
//   '200':
//     description: The provider with the ID
//     schema:
//       "$ref": "#/definitions/Provider"
//   '404' :
//     description: Provider not found
func (a *App) GetProvider(w http.ResponseWriter, r *http.Request) {
	a.get(w, r, func(id string) (interface{}, error) {
		return a.Repository.GetProvider(id)
	})
}

// CreateProvider creates a provider passed by REST params
// swagger:operation POST /providers createProvider
//
// Creates a provider with the information passed in the request body
//
// ---
// produces:
// - application/json
// consumes:
// - application/json
// parameters:
// - name: provider
//   in: body
//   description: The provider to create
//   required: true
//   schema:
//     "$ref": "#/definitions/Provider"
// responses:
//   '200':
//     description: The new provider that has been created
//     schema:
//       "$ref": "#/definitions/Provider"
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
// swagger:operation DELETE /providers/{id} deleteProvider
//
// Deletes a provider given its ID
//
// ---
// produces:
// - application/json
// parameters:
// - name: id
//   in: path
//   description: The identifier of the provider
//   required: true
//   type: string
// responses:
//   '200':
//     description: The provider has been successfully deleted
//   '404' :
//     description: Provider not found
func (a *App) DeleteProvider(w http.ResponseWriter, r *http.Request) {
	a.update(w, r, func(id string) error {
		return a.Repository.DeleteProvider(&model.Provider{Id: id})
	})
}

// GetAgreements return all agreements in db
// swagger:operation GET /agreements getAllAgreements
//
// Returns all registered agreements
//
// ---
// produces:
// - application/json
// responses:
//   '200':
//     description: The complete list of registered agreements
//     schema:
//       type: object
//       additionalProperties:
//         "$ref": "#/definitions/Agreements"
func (a *App) GetAgreements(w http.ResponseWriter, r *http.Request) {
	v := r.URL.Query()
	active := v.Get("active")

	a.getAll(w, r, func() (interface{}, error) {
		if active != "" {
			return a.Repository.GetAgreementsByState(model.STARTED)
		}
		return a.Repository.GetAllAgreements()
	})
}

// GetAgreement gets an agreement by REST ID
// swagger:operation GET /agreements/{id} getAgreement
//
// Returns a agreement given its ID
//
// ---
// produces:
// - application/json
// parameters:
// - name: id
//   in: path
//   description: The identifier of the agreement
//   required: true
//   type: string
// responses:
//   '200':
//     description: The agreement with the ID
//     schema:
//       "$ref": "#/definitions/Agreement"
//   '404' :
//     description: Agreement not found
func (a *App) GetAgreement(w http.ResponseWriter, r *http.Request) {
	a.get(w, r, func(id string) (interface{}, error) {
		return a.Repository.GetAgreement(id)
	})
}

// GetAgreementDetails gets an agreement by REST ID
// swagger:operation GET /agreements/{id}/details getAgreementDetails
//
// Returns the agreement details given its ID
//
// ---
// produces:
// - application/json
// parameters:
// - name: id
//   in: path
//   description: The identifier of the agreement
//   required: true
//   type: string
// responses:
//   '200':
//     description: The agreement details with the provided ID
//     schema:
//       "$ref": "#/definitions/Details"
//   '404' :
//     description: Agreement not found
func (a *App) GetAgreementDetails(w http.ResponseWriter, r *http.Request) {
	a.get(w, r, func(id string) (interface{}, error) {
		agreement, error := a.Repository.GetAgreement(id)
		return agreement.Details, error
	})
}

// CreateAgreement creates a agreement passed by REST params
// swagger:operation POST /agreements createAgreement
//
// Creates an agreement with the information passed in the request body
//
// ---
// produces:
// - application/json
// consumes:
// - application/json
// parameters:
// - name: agreement
//   in: body
//   description: The agreement to create
//   required: true
//   schema:
//     "$ref": "#/definitions/Agreement"
// responses:
//   '200':
//     description: The new agreement that has been created
//     schema:
//       "$ref": "#/definitions/Agreement"
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
// swagger:operation DELETE /agreements/{id} deleteAgreement
//
// Deletes an agreement given its ID
//
// ---
// produces:
// - application/json
// parameters:
// - name: id
//   in: path
//   description: The identifier of the agreement
//   required: true
//   type: string
// responses:
//   '200':
//     description: The agreement has been successfully deleted
//   '404' :
//     description: Agreement not found
func (a *App) DeleteAgreement(w http.ResponseWriter, r *http.Request) {
	a.update(w, r, func(id string) error {
		return a.Repository.DeleteAgreement(&model.Agreement{Id: id})
	})
}

// UpdateAgreement updates the only field updateable by REST in an agreement: the state.
// The Id in the body is ignored; only the id path is taken into account.
// swagger:operation PUT /agreements/{id} updateAgreement
//
// Updates information in the agreement whose ID is passed as parameter. Only state is updated.
//
// ---
// produces:
// - application/json
// parameters:
// - name: id
//   in: path
//   description: The identifier of the agreement
//   required: true
//   type: string
// - name: agreement
//   in: body
//   description: The information to update
//   required: true
//   schema:
//     "$ref": "#/definitions/Agreement"
// responses:
//   '200':
//     description: The updated agreement
//     schema:
//       "$ref": "#/definitions/Agreement"
//   '404' :
//     description: Agreement not found
func (a *App) UpdateAgreement(w http.ResponseWriter, r *http.Request) {
	var agreement model.Agreement

	a.updateEntity(w, r,
		func() error {
			return json.NewDecoder(r.Body).Decode(&agreement)
		},
		func(id string) (model.Identity, error) {
			newState := agreement.State
			return a.Repository.UpdateAgreementState(id, newState)
		})
}

// StartAgreement starts monitoring an agreement
func (a *App) StartAgreement(w http.ResponseWriter, r *http.Request) {
	a.update(w, r, func(id string) error {
		_, err := a.Repository.UpdateAgreementState(id, model.STARTED)
		return err
	})
}

// StopAgreement stop monitoring an agreement
func (a *App) StopAgreement(w http.ResponseWriter, r *http.Request) {
	a.update(w, r, func(id string) error {
		_, err := a.Repository.UpdateAgreementState(id, model.STOPPED)
		return err
	})
}

// TerminateAgreement terminates an agreement
func (a *App) TerminateAgreement(w http.ResponseWriter, r *http.Request) {
	a.update(w, r, func(id string) error {
		_, err := a.Repository.UpdateAgreementState(id, model.TERMINATED)
		return err
	})
}

// GetTemplates return all templates in db
// swagger:operation GET /templates getAllTemplates
//
// Returns all registered templates
//
// ---
// produces:
// - application/json
// responses:
//   '200':
//     description: The complete list of registered templates
//     schema:
//       type: object
//       additionalProperties:
//         "$ref": "#/definitions/Templates"
func (a *App) GetTemplates(w http.ResponseWriter, r *http.Request) {

	a.getAll(w, r, func() (interface{}, error) {
		return a.Repository.GetAllTemplates()
	})
}

// GetTemplate gets a template by REST ID
// swagger:operation GET /templates/{id} getTemplate
//
// Returns a template given its ID
//
// ---
// produces:
// - application/json
// parameters:
// - name: id
//   in: path
//   description: The identifier of the template
//   required: true
//   type: string
// responses:
//   '200':
//     description: The template with the ID
//     schema:
//       "$ref": "#/definitions/Template"
//   '404' :
//     description: Template not found
func (a *App) GetTemplate(w http.ResponseWriter, r *http.Request) {
	a.get(w, r, func(id string) (interface{}, error) {
		return a.Repository.GetTemplate(id)
	})
}

// CreateTemplate creates a template passed by REST params
// swagger:operation POST /templates createTemplate
//
// Creates a template with the information passed in the request body
//
// ---
// produces:
// - application/json
// consumes:
// - application/json
// parameters:
// - name: template
//   in: body
//   description: The template to create
//   required: true
//   schema:
//     "$ref": "#/definitions/Template"
// responses:
//   '200':
//     description: The new template that has been created
//     schema:
//       "$ref": "#/definitions/Template"
func (a *App) CreateTemplate(w http.ResponseWriter, r *http.Request) {

	var template model.Template

	a.create(w, r,
		func() error {
			return json.NewDecoder(r.Body).Decode(&template)
		},
		func() (model.Identity, error) {
			return a.Repository.CreateTemplate(&template)
		})
}

func manageError(err error, w http.ResponseWriter) {
	switch err {
	case model.ErrAlreadyExist:
		respondWithError(w, http.StatusConflict, "Object already exist")
	case model.ErrNotFound:
		respondWithError(w, http.StatusNotFound, "Can't find object")
	default:
		if model.IsErrValidation(err) {
			respondWithError(w, http.StatusBadRequest, err.Error())
		} else {
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
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
