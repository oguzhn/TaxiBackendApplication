package database

import (
	"context"
	"log"
	"math"

	"github.com/globalsign/mgo/bson"
	models "github.com/oguzhn/TaxiBackendApplication"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

/*MongoClient is a mongo client which has a session, database name and collection name*/
type MongoClient struct {
	ms      *mongo.Client
	dbName  string
	colName string
	ctx     context.Context
}

// NewDatastore create connection to mongo db given the connection string
func NewDatastore(con string, db string, collection string, ctx context.Context) (*MongoClient, error) {
	client, err := mongo.NewClient(options.Client().ApplyURI(con))
	if err != nil {
		log.Fatal(err)
	}
	err = client.Connect(ctx)

	if err != nil {
		log.Fatal(err)
	}

	return &MongoClient{client, db, collection, ctx}, nil
}

func (m *MongoClient) TripsInASpecifiedRegion(query models.Query) ([]models.Trip, error) {
	collection := m.ms.Database(m.dbName).Collection(m.colName)
	// cursor, err := collection.Find(m.ctx, bson.M{"start": bson.M{"$geoWithin": bson.M{"$centerSphere": []interface{}{[]interface{}{query.Point.Long, query.Point.Lat}, query.Circle.RadiusInKm / models.EarthRadiusInKm}}}})
	filter := []bson.M{
		{"$geoNear": bson.M{
			"near":          bson.M{"type": "Point", "coordinates": []interface{}{query.Point.Long, query.Point.Lat}},
			"distanceField": "dist.calculated",
			"maxDistance":   query.RadiusInKm * models.EarthRadiusInKm / (math.Pi * 2),
			"key":           "start",
			"spherical":     true,
		}},
	}
	if !query.StartDate.IsZero() {
		filter = append(filter, bson.M{
			"$redact": bson.M{
				"$cond": bson.M{
					"if":   bson.M{"$gte": []interface{}{"$start_date", query.StartDate}},
					"then": "$$KEEP",
					"else": "$$PRUNE",
				},
			},
		})
	}
	if !query.EndDate.IsZero() {
		filter = append(filter, bson.M{
			"$redact": bson.M{
				"$cond": bson.M{
					"if":   bson.M{"$lte": []interface{}{"$complete_date", query.EndDate}},
					"then": "$$KEEP",
					"else": "$$PRUNE",
				},
			},
		})
	}
	cursor, err := collection.Aggregate(m.ctx, filter)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	var trips []models.Trip

	err = cursor.All(m.ctx, &trips)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	if trips == nil {
		trips = []models.Trip{}
	}
	return trips, nil
}

func (m *MongoClient) MinMaxDistanceTravelledInASpecifiedRegion(query models.Circle) ([2]int, error) {
	collection := m.ms.Database(m.dbName).Collection(m.colName)
	cursor, err := collection.Aggregate(m.ctx, []bson.M{
		{"$geoNear": bson.M{
			"near":          bson.M{"type": "Point", "coordinates": []interface{}{query.Point.Long, query.Point.Lat}},
			"distanceField": "dist.calculated",
			"maxDistance":   query.RadiusInKm * models.EarthRadiusInKm / (math.Pi * 2),
			"key":           "start",
			"spherical":     true,
		}},
		{"$sort": bson.M{"distance_travelled": 1}},
		{"$limit": 1},
	})

	if err != nil {
		return [2]int{}, err
	}
	var trips []models.Trip

	err = cursor.All(m.ctx, &trips)
	if err != nil {
		return [2]int{}, err
	}
	max, min := 0, math.MaxInt32
	if len(trips) >= 1 {
		min = trips[0].DistanceTravelled
	}
	cursor, err = collection.Aggregate(m.ctx, []bson.M{
		{"$geoNear": bson.M{
			"near":          bson.M{"type": "Point", "coordinates": []interface{}{query.Point.Long, query.Point.Lat}},
			"distanceField": "dist.calculated",
			"maxDistance":   query.RadiusInKm * models.EarthRadiusInKm / (math.Pi * 2),
			"key":           "start",
			"spherical":     true,
		}},
		{"$sort": bson.M{"distance_travelled": -1}},
		{"$limit": 1},
	})
	if err != nil {
		return [2]int{}, err
	}
	err = cursor.All(m.ctx, &trips)
	if err != nil {
		return [2]int{}, err
	}
	if len(trips) >= 1 {
		max = trips[0].DistanceTravelled
	}

	return [2]int{min, max}, nil
}

func (m *MongoClient) ReportModelYear(query models.Circle) ([]models.ReportModelYear, error) {
	collection := m.ms.Database(m.dbName).Collection(m.colName)
	// cursor, err := collection.Find(m.ctx, bson.M{"start": bson.M{"$geoWithin": bson.M{"$centerSphere": []interface{}{[]interface{}{query.Point.Long, query.Point.Lat}, query.RadiusInKm / models.EarthRadiusInKm}}}})
	cursor, err := collection.Aggregate(m.ctx, []bson.M{
		{"$geoNear": bson.M{
			"near":          bson.M{"type": "Point", "coordinates": []interface{}{query.Point.Long, query.Point.Lat}},
			"distanceField": "dist.calculated",
			"maxDistance":   query.RadiusInKm * models.EarthRadiusInKm / (math.Pi * 2),
			"key":           "start",
			"spherical":     true,
		}},
		{"$group": bson.M{
			"_id":   "$year",
			"trips": bson.M{"$sum": 1},
		}},
	})
	if err != nil {
		return nil, err
	}
	var result []models.ReportModelYear

	err = cursor.All(m.ctx, &result)
	if err != nil {
		return nil, err
	}
	if result == nil {
		result = []models.ReportModelYear{}
	}
	return result, nil
}
