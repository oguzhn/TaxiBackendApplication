package models

import (
	"math"
	"time"
)

type Trip struct {
	ID                string    `json:"id" bson:"_id"`
	DistanceTravelled int       `json:"distance_travelled" bson:"distance_travelled"`
	StartDate         time.Time `json:"start_date" bson:"start_date"`
	CompleteDate      time.Time `json:"complete_date" bson:"complete_date"`
	ModelYear         int       `json:"model_year" bson:"year"`
	Model             string    `json:"model" bson:"model"`
	Make              string    `json:"make" bson:"make"`
	Color             string    `json:"color" bson:"color"`
}

type ReportModelYear struct {
	NumberOfTrips int `json:"number_of_trips" bson:"trips"`
	ModelYear     int `json:"model_year" bson:"_id"`
}

type Point struct {
	Long float64 `json:"long"`
	Lat  float64 `json:"lat"`
}

type Query struct {
	Circle
	StartDate time.Time `json:"start_date,omitempty"`
	EndDate   time.Time `json:"end_date,omitempty"`
}

type Circle struct {
	Point      Point   `json:"point"`
	RadiusInKm float64 `json:"radius"`
}

func (c Circle) IsInsideRadius(p Point) bool {
	return c.RadiusInKm >= c.Point.distanceInMeters(p)
}

// haversin(θ) function
func hsin(theta float64) float64 {
	return math.Pow(math.Sin(theta/2), 2)
}

const (
	EarthRadiusInMeters = 6378137 // Earth radius in METERS
	EarthRadiusInKm     = 6378
)

// distance returned is METERS!!!!!!
// http://en.wikipedia.org/wiki/Haversine_formula
func (c Point) distanceInMeters(p Point) float64 {
	// convert to radians
	// must cast radius as float to multiply later
	var la1, lo1, la2, lo2 float64
	la1 = c.Lat * math.Pi / 180
	lo1 = c.Long * math.Pi / 180
	la2 = p.Lat * math.Pi / 180
	lo2 = p.Long * math.Pi / 180

	// calculate
	h := hsin(la2-la1) + math.Cos(la1)*math.Cos(la2)*hsin(lo2-lo1)

	return 2 * EarthRadiusInMeters * math.Asin(math.Sqrt(h))
}
