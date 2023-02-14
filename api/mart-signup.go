package api

import (
	"capszo-mart/database"
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

func (server *Server) martSignup(ctx *gin.Context) {
	var request database.Mart
	var err error
	db := server.mongoDB.Database("capszo")
	martColl := db.Collection(string(database.MartColl))

	// get request data
	if err = ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// check if mobile number exists
	filter := bson.M{"mobile_number": request.MobileNumber}
	count, err := martColl.CountDocuments(context.TODO(), filter)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	if count > 0 {
		err = errors.New("MOBILE NUMBER ALREADY EXISTS")
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// store mart data in db
	result, err := martColl.InsertOne(context.TODO(), request)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// response
	ctx.JSON(http.StatusCreated, gin.H{"message": "mart created successfully", "mart_id": result.InsertedID})
}
