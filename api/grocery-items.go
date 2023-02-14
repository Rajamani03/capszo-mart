package api

import (
	"capszo-mart/database"
	"capszo-mart/token"
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

func (server *Server) getAllItems(ctx *gin.Context) {
	var items []database.Item
	db := server.mongoDB.Database("capszo")
	groceriesColl := db.Collection(string(database.GroceryColl))

	// get mart_id from params
	martID := ctx.Param("mart-id")

	// get all grocery items from groceries collections
	filter := bson.D{{Key: "mart_id", Value: martID}}
	cursor, err := groceriesColl.Find(context.TODO(), filter)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	if err = cursor.All(context.TODO(), &items); err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// response
	ctx.JSON(http.StatusOK, items)
}

func (server *Server) addItems(ctx *gin.Context) {
	var request []database.Item
	var err error
	db := server.mongoDB.Database("capszo")
	groceriesColl := db.Collection(string(database.GroceryColl))

	// get token payload
	tokenPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	// get request data
	if err = ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// change the datatype
	var items []interface{}
	for _, item := range request {
		if tokenPayload.TokenFor == token.MartAccess {
			item.MartID = tokenPayload.UserID
		}
		items = append(items, item)
	}

	// insert all the items
	result, err := groceriesColl.InsertMany(context.TODO(), items)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// response
	ctx.JSON(http.StatusCreated, gin.H{"message": "items added successfully", "item_ids": result.InsertedIDs})
}
