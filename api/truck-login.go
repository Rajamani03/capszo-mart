package api

import (
	"capszo-mart/database"
	"capszo-mart/token"
	"capszo-mart/util"
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (server *Server) getTruckLoginOTP(ctx *gin.Context) {
	var request loginOTPRequest
	var err error
	db := server.mongoDB.Database("capszo")
	loginInfoColl := db.Collection("login_info")
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

	// generate otp
	// otp := util.GetOTP(6)
	otp := "654321"
	request.OTP = util.Hash(otp + server.config.Salt)

	// store the login data to login_info collection
	request.CreatedAt = time.Now()
	result, err := loginInfoColl.InsertOne(context.TODO(), request)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// response
	ctx.JSON(http.StatusOK, gin.H{"validate_key": getID(result.InsertedID)})
}

func (server *Server) truckLogin(ctx *gin.Context) {
	var request loginRequest
	var loginInfo loginOTPRequest
	var truck database.Truck
	var err error
	db := server.mongoDB.Database("capszo")
	loginInfoColl := db.Collection("login_info")
	truckColl := db.Collection(string(database.TruckColl))

	// get request data
	if err = ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// convert validate key string to objectID
	objectID, err := primitive.ObjectIDFromHex(request.ValidateKey)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// get login info
	filter := bson.M{"_id": objectID}
	if err = loginInfoColl.FindOne(context.TODO(), filter).Decode(&loginInfo); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// validate input and saved otp
	hotp := util.Hash(request.OTP + server.config.Salt)
	if hotp != loginInfo.OTP {
		err = errors.New("INCORRECT OTP")
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// convert validate key string to objectID
	objectID, err = primitive.ObjectIDFromHex(loginInfo.UserID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// get truck info
	filter = bson.M{"_id": objectID}
	if err = truckColl.FindOne(context.TODO(), filter).Decode(&truck); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// get access and refresh token
	accessToken, refreshToken, err := server.getAuthTokens(loginInfo.UserID, token.TruckAccess)
	if err != nil {
		err = errors.New("TOKEN QUERY ERROR")
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// response
	ctx.JSON(http.StatusOK, gin.H{"access_token": accessToken, "refresh_token": refreshToken, "user_info": truck})
}
