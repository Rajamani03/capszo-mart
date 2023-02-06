package main

import (
	"capszo-mart/api"
	"capszo-mart/database"
	"capszo-mart/util"
	"context"
	"fmt"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {

	var server *api.Server
	var config util.Config
	var err error

	environment := os.Getenv("ENVIRONMENT")
	if environment != "" {
		gin.SetMode(environment)
	}

	config, err = util.LoadConfig(".")
	if err != nil {
		fmt.Println("CANNOT LOAD CONFIG FILE: %w", err)
		panic(err.Error())
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	mongoDB, err := database.ConnectMongoDB(config, ctx)
	if err != nil {
		panic(err.Error())
	}
	defer mongoDB.Disconnect(ctx)
	defer cancel()

	server, err = api.NewServer(mongoDB, config)
	if err != nil {
		panic(err.Error())
	}

	server.SetupRouter()
	server.Start(config.ServerAddress)
}
