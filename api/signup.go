package api

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type signupOTPRequest struct {
	Name         string    `json:"name" bson:"name" binding:"required"`
	MobileNumber string    `json:"mobile_number" bson:"mobile_number" binding:"required,numeric,len=10"`
	OTP          string    `json:"otp" bson:"otp"`
	CreatedAt    time.Time `json:"created_at" bson:"created_at"`
}

func checkUserExists(collection *mongo.Collection, mobileNumber string) (bool, error) {
	var user map[string]interface{}
	var err error
	filter := bson.M{"mobile_number": mobileNumber}
	if err = collection.FindOne(context.TODO(), filter).Decode(&user); err != nil {
		if err == mongo.ErrNoDocuments {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
