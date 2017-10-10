package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sla/model"
	"strconv"

	"github.com/gorilla/mux"
)

// App is a main application "object", to be built by main and testmain
type App struct {
	Router     *mux.Router
	Repository model.IRepository
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

	a.Router.HandleFunc("/providers", a.GetAllProviders).Methods("GET")
	a.Router.HandleFunc("/providers/{id}", a.GetProvider).Methods("GET")
	a.Router.HandleFunc("/providers", a.CreateProvider).Methods("POST")
}

// Run starts the REST API
func (a *App) Run(addr string) {
	log.Fatal(http.ListenAndServe(addr, a.Router))
}

// Index is the API index
func (a *App) Index(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(api)
}

// GetAllProviders return all providers in db
func (a *App) GetAllProviders(w http.ResponseWriter, r *http.Request) {
	providers, err := a.Repository.GetAllProviders()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
	}
	json.NewEncoder(w).Encode(providers)
}

// GetProvider gets a provider by REST ID
func (a *App) GetProvider(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	provider, err := a.Repository.GetProvider(id)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			respondWithError(w, http.StatusNotFound, fmt.Sprintf("Provider{id:%s} not found", id))
		default:
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	json.NewEncoder(w).Encode(provider)
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
		case sql.ErrNoRows: /* XXX CHANGE Err type !!! */
			respondWithError(w, http.StatusConflict,
				fmt.Sprintf("Provider{id: %s} already exists", provider.Id))
		default:
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
	}
	respondWithJSON(w, http.StatusCreated, created)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"code": strconv.Itoa(code), "error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	// response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	// w.Write(response)
	json.NewEncoder(w).Encode(payload)
}
