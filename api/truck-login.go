package api

import (
	"capszo-mart/database"
	"capszo-mart/token"
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (server *Server) getTruckLoginOTP(ctx *gin.Context) {
	var request loginOTPRequest
	var err error
	db := server.mongoDB.Database("capszo")
	truckColl := db.Collection(string(database.TruckColl))

	// get request data
	if err = ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// check if mobile number exists
	var user map[string]interface{}
	filter := bson.M{"mobile_number": request.MobileNumber}
	err = truckColl.FindOne(context.TODO(), filter).Decode(&user)
	if err != nil {
		err = errors.New("USER NOT FOUND")
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	request.UserID = getID(user["_id"])

	// store login OTP
	validateKey, err := server.storeLoginOTP(request)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// response
	ctx.JSON(http.StatusOK, gin.H{"validate_key": validateKey})
}

func (server *Server) truckLogin(ctx *gin.Context) {
	var request loginRequest
	var loginInfo loginOTPRequest
	var truck database.Truck
	var err error
	db := server.mongoDB.Database("capszo")
	truckColl := db.Collection(string(database.TruckColl))

	// get request data
	if err = ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// validate login otp
	loginInfo, err = server.validateLoginOTP(request)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// convert validate key string to objectID
	objectID, err := primitive.ObjectIDFromHex(loginInfo.UserID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// get truck info
	filter := bson.M{"_id": objectID}
	if err = truckColl.FindOne(context.TODO(), filter).Decode(&truck); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// create session and get access and refresh token
	sessionID, accessToken, refreshToken, err := server.createSession(loginInfo.UserID, token.TruckAccess)
	if err != nil {
		err = errors.New("SESSION ERROR")
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// response
	ctx.JSON(http.StatusOK, gin.H{"session_id": sessionID, "access_token": accessToken, "refresh_token": refreshToken, "user_info": truck})
}
