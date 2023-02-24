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

func (server *Server) getMartLoginOTP(ctx *gin.Context) {
	var request loginOTPRequest
	var err error
	db := server.mongoDB.Database("capszo")
	loginInfoColl := db.Collection("login_info")
	martColl := db.Collection(string(database.MartColl))

	// get request data
	if err = ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// check if mobile number exists
	var user map[string]interface{}
	filter := bson.M{"mobile_number": request.MobileNumber}
	err = martColl.FindOne(context.TODO(), filter).Decode(&user)
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

func (server *Server) martLogin(ctx *gin.Context) {
	var request loginRequest
	var loginInfo loginOTPRequest
	var mart database.Mart
	var err error
	db := server.mongoDB.Database("capszo")
	loginInfoColl := db.Collection("login_info")
	martColl := db.Collection(string(database.MartColl))

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

	// get mart info
	filter = bson.M{"_id": objectID}
	if err = martColl.FindOne(context.TODO(), filter).Decode(&mart); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// create session and get access and refresh token
	sessionID, accessToken, refreshToken, err := server.createSession(loginInfo.UserID, token.MartAccess)
	if err != nil {
		err = errors.New("SESSION ERROR")
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// response
	ctx.JSON(http.StatusOK, gin.H{"session_id": sessionID, "access_token": accessToken, "refresh_token": refreshToken, "user_info": mart})
}
