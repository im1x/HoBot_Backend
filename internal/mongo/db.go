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
	Users               CollectionName = "Users"
	Tokens                             = "Tokens"
	Vkpl                               = "Vkpl"
	Config                             = "Config"
	SettingsOptions                    = "SettingsOptions"
	UserSettings                       = "UserSettings"
	SongRequests                       = "SongRequests"
	SongRequestsHistory                = "SongRequestsHistory"
	Feedback                           = "Feedback"
	Statistics                         = "Statistics"
)

var ctx = context.TODO()
var DB *mongo.Client = nil

func Connect() {
	clientOptions := options.Client().ApplyURI(os.Getenv("MONGODB_URI"))
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	DB = client
}

func GetCollection(col CollectionName) *mongo.Collection {
	return DB.Database(os.Getenv("DB_NAME")).Collection(string(col))
}
