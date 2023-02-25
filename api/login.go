package api

import (
	"capszo-mart/util"
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type loginOTPRequest struct {
	UserID       string    `json:"user_id" bson:"user_id"`
	MobileNumber string    `json:"mobile_number" bson:"mobile_number" binding:"required,numeric,len=10"`
	OTP          string    `json:"otp" bson:"otp"`
	CreatedAt    time.Time `json:"created_at" bson:"created_at"`
}

type loginRequest struct {
	ValidateKey string `json:"validate_key" binding:"required"`
	OTP         string `json:"otp" binding:"required,numeric,len=6"`
}

func (server *Server) storeLoginOTP(request loginOTPRequest) (validateKey string, err error) {
	db := server.mongoDB.Database("capszo")
	loginInfoColl := db.Collection("login_info")

	// create TTL index
	model := mongo.IndexModel{
		Keys:    bson.M{"created_at": 1},
		Options: options.Index().SetExpireAfterSeconds(60),
	}
	_, err = loginInfoColl.Indexes().CreateOne(context.TODO(), model)
	if err != nil {
		return
	}

	// generate otp
	// otp := util.GetOTP(6)
	otp := "654321"
	request.OTP = util.Hash(otp + server.config.Salt)

	// store the login data to login_info collection
	request.CreatedAt = time.Now()
	result, err := loginInfoColl.InsertOne(context.TODO(), request)
	if err != nil {
		return
	}

	return getID(result.InsertedID), err
}

func (server *Server) validateLoginOTP(request loginRequest) (loginInfo loginOTPRequest, err error) {
	db := server.mongoDB.Database("capszo")
	loginInfoColl := db.Collection("login_info")

	// convert validate key string to objectID
	objectID, err := primitive.ObjectIDFromHex(request.ValidateKey)
	if err != nil {
		return
	}

	// get login info
	filter := bson.M{"_id": objectID}
	err = loginInfoColl.FindOne(context.TODO(), filter).Decode(&loginInfo)
	if err != nil {
		return
	}

	// validate input and saved otp
	hotp := util.Hash(request.OTP + server.config.Salt)
	if hotp != loginInfo.OTP {
		err = errors.New("INCORRECT OTP")
		return
	}

	return
}

// for future redis migration

// func storeLoginOTP
// inputs - userID, mobileNumber
// generate OTP
// generate validate key
// store on redis with {validate_key} as key and {user_id & otp} as value
// outputs - validateKey, err

// func validateLoginOTP
// inputs - validateKey, otp
// get data from redis using validate key
// validate input otp with redis data otp
// outputs - userID, err
