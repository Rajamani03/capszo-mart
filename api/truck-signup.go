package api

import (
	"capszo-mart/database"
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func (server *Server) truckSignup(ctx *gin.Context) {
	var request database.Truck
	var err error
	db := server.mongoDB.Database("capszo")
	truckColl := db.Collection(string(database.TruckColl))

	// get request data
	if err = ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// check if mobile number exists
	isUserExists, err := checkUserExists(truckColl, request.MobileNumber)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	if isUserExists {
		err = errors.New("USER ALREADY EXISTS")
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// store truck data in db
	request.CreatedAt = time.Now()
	result, err := truckColl.InsertOne(context.TODO(), request)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// response
	ctx.JSON(http.StatusCreated, gin.H{"message": "truck created successfully", "truck_id": result.InsertedID})
}
