package controller

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	models "github.com/oguzhn/TaxiBackendApplication"
	"github.com/oguzhn/TaxiBackendApplication/authentication"
)

type Controller struct {
	handler IHandler
	auth    authentication.Authenticator
	loginer authentication.Loginer
}

func NewController(handler IHandler, auth authentication.Authenticator, loginer authentication.Loginer) *Controller {
	return &Controller{handler: handler, auth: auth, loginer: loginer}
}
func (c *Controller) RegisterHandlers() http.Handler {
	router := mux.NewRouter()
	router.Path("/login").Methods(http.MethodGet).HandlerFunc(c.Login)
	router.Path("/tripsinaspecifiedregion").Methods(http.MethodGet).HandlerFunc(c.TripsInASpecifiedRegion)
	router.Path("/minmaxdistancetravelledinaspecifiedregion").Methods(http.MethodGet).HandlerFunc(c.MinMaxDistanceTravelledInASpecifiedRegion)
	router.Path("/reportmodelyear").Methods(http.MethodGet).HandlerFunc(c.ReportModelYear)
	return router
}

func (c *Controller) Login(w http.ResponseWriter, r *http.Request) {
	values := r.URL.Query()
	username := values.Get("username")
	password := values.Get("password")
	token, err := c.loginer.Login(username, password)
	if err != nil {
		log.Println(err)
		http.Error(w,
			http.StatusText(http.StatusUnauthorized),
			http.StatusUnauthorized)
		return
	}
	w.Write([]byte(token))
}

func (c *Controller) TripsInASpecifiedRegion(w http.ResponseWriter, r *http.Request) {
	client, err := c.auth.Authenticate(r)
	if err != nil {
		http.Error(w,
			http.StatusText(http.StatusUnauthorized),
			http.StatusUnauthorized)
		return
	}
	if client.Role() != authentication.RoleAdmin {
		http.Error(w,
			http.StatusText(http.StatusForbidden),
			http.StatusForbidden)
		return
	}
	var query models.Query
	err = json.NewDecoder(r.Body).Decode(&query)
	if err != nil {
		log.Printf("Failed to decode json form: %+v, err: %s\n", r.Body, err)
		http.Error(w,
			http.StatusText(http.StatusBadRequest),
			http.StatusBadRequest)
		return
	}
	trips, err := c.handler.TripsInASpecifiedRegion(query)
	if err != nil {
		log.Println(err)
		http.Error(w,
			http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(trips)
	if err != nil {
		log.Printf("Failed to load marshal data: %v, err: %s\n", r, err)
	}
}

func (c *Controller) MinMaxDistanceTravelledInASpecifiedRegion(w http.ResponseWriter, r *http.Request) {
	client, err := c.auth.Authenticate(r)
	if err != nil {
		http.Error(w,
			http.StatusText(http.StatusUnauthorized),
			http.StatusUnauthorized)
		return
	}
	if client.Role() != authentication.RoleAdmin {
		http.Error(w,
			http.StatusText(http.StatusForbidden),
			http.StatusForbidden)
		return
	}
	var query models.Circle
	err = json.NewDecoder(r.Body).Decode(&query)
	if err != nil {
		log.Printf("Failed to decode json form: %+v, err: %s\n", r.Body, err)
		http.Error(w,
			http.StatusText(http.StatusBadRequest),
			http.StatusBadRequest)
		return
	}
	minMaxDistanceTravelled, err := c.handler.MinMaxDistanceTravelledInASpecifiedRegion(query)
	if err != nil {
		log.Println(err)
		http.Error(w,
			http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(minMaxDistanceTravelled)
	if err != nil {
		log.Printf("Failed to load marshal data: %v, err: %s\n", r, err)
	}
}

func (c *Controller) ReportModelYear(w http.ResponseWriter, r *http.Request) {
	client, err := c.auth.Authenticate(r)
	if err != nil {
		http.Error(w,
			http.StatusText(http.StatusUnauthorized),
			http.StatusUnauthorized)
		return
	}
	if client.Role() != authentication.RoleAdmin {
		http.Error(w,
			http.StatusText(http.StatusForbidden),
			http.StatusForbidden)
		return
	}
	var query models.Circle
	err = json.NewDecoder(r.Body).Decode(&query)
	if err != nil {
		log.Printf("Failed to decode json form: %+v, err: %s\n", r.Body, err)
		http.Error(w,
			http.StatusText(http.StatusBadRequest),
			http.StatusBadRequest)
		return
	}
	report, err := c.handler.ReportModelYear(query)
	if err != nil {
		log.Println(err)
		http.Error(w,
			http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(report)
	if err != nil {
		log.Printf("Failed to load marshal data: %v, err: %s\n", r, err)
	}
}

type IHandler interface {
	TripsInASpecifiedRegion(models.Query) ([]models.Trip, error)
	MinMaxDistanceTravelledInASpecifiedRegion(models.Circle) ([2]int, error)
	ReportModelYear(models.Circle) ([]models.ReportModelYear, error)
}
