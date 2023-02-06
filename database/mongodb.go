package database

import (
	"capszo-mart/util"
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func ConnectMongoDB(config util.Config, ctx context.Context) (*mongo.Client, error) {
	var mongoDB *mongo.Client
	var err error

	mongoDB, err = mongo.NewClient(options.Client().ApplyURI(config.DBSource))
	if err != nil {
		return nil, err
	}

	err = mongoDB.Connect(ctx)
	if err != nil {
		return nil, err
	}

	err = mongoDB.Ping(ctx, readpref.Primary())
	if err != nil {
		return nil, err
	}

	return mongoDB, err
}
