package api

import (
	"capszo-mart/database"
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
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
	isUserExists, err := checkUserExists(martColl, request.MobileNumber)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	if isUserExists {
		err = errors.New("USER ALREADY EXISTS")
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// store mart data in db
	request.CreatedAt = time.Now()
	result, err := martColl.InsertOne(context.TODO(), request)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// response
	ctx.JSON(http.StatusCreated, gin.H{"message": "mart created successfully", "mart_id": result.InsertedID})
}
