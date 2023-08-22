package mongo

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"os"
)

type CollectionName string

const (
	Users    CollectionName = "users"
	Settings CollectionName = "settings"
)

var ctx = context.TODO()
var DB *mongo.Client = connect()

func connect() *mongo.Client {
	clientOptions := options.Client().ApplyURI(os.Getenv("MONGODB_URI"))
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	return client
}

func GetCollection(col CollectionName) *mongo.Collection {
	return DB.Database(os.Getenv("DB_NAME")).Collection(string(col))
}
