package api

import (
	"capszo-mart/database"
	"capszo-mart/token"
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (server *Server) updateGroceryBasket(ctx *gin.Context) {
	var request []database.BasketItem
	var err error
	db := server.mongoDB.Database("capszo")
	customerColl := db.Collection("customers")

	// get token payload
	tokenPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	// get request data
	if err = ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// convert userID string to objectID
	objectID, err := primitive.ObjectIDFromHex(tokenPayload.UserID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	update := bson.D{{Key: "$set", Value: bson.D{{Key: "grocery_basket", Value: request}}}}
	_, err = customerColl.UpdateByID(context.TODO(), objectID, update)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// response
	ctx.JSON(http.StatusOK, gin.H{"message": "basket updated successfully"})
}
