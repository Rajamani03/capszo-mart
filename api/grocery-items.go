package api

import (
	"capszo-mart/blueprint"
	"capszo-mart/database"
	"capszo-mart/token"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (server *Server) getAllItems(ctx *gin.Context) {
	var items []database.Item
	db := server.mongoDB.Database("capszo")
	groceriesColl := db.Collection(string(database.GroceryColl))

	// get mart_id from params
	martID := ctx.Param("mart-id")

	// get all grocery items from groceries collections
	filter := bson.M{"mart_id": martID}
	cursor, err := groceriesColl.Find(context.TODO(), filter)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	if err = cursor.All(context.TODO(), &items); err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// transform items
	var transformedItems []map[string]interface{}
	for _, item := range items {
		transformedItem, err := blueprint.ItemTransform(item, blueprint.CustomerItem)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}
		transformedItems = append(transformedItems, transformedItem)
	}

	// response
	ctx.JSON(http.StatusOK, transformedItems)
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
		item.CreatedAt = time.Now()
		item.UpdatedAt = time.Now()
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

func (server *Server) updateItem(ctx *gin.Context) {
	var request database.Item
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

	// convert itemID string to objectID
	objectID, err := primitive.ObjectIDFromHex(toString(request.ID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// set filters for query
	var filter primitive.M
	if tokenPayload.TokenFor == token.AdminAccess {
		filter = bson.M{"_id": objectID}
	} else {
		filter = bson.M{"_id": objectID, "mart_id": tokenPayload.UserID}
		request.MartID = tokenPayload.UserID
	}

	// update grocery items
	request.ID = objectID
	request.UpdatedAt = time.Now()
	update := bson.M{"$set": request}
	result, err := groceriesColl.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	fmt.Printf("result.MatchedCount: %v\n", result.MatchedCount)

	// response
	ctx.JSON(http.StatusOK, gin.H{"message": "items updated successfully", "item_id": request.ID})
}

func (server *Server) searchItem(ctx *gin.Context) {
	var martID string
	var query string
	var items []database.Item
	var err error
	db := server.mongoDB.Database("capszo")
	groceriesColl := db.Collection(string(database.GroceryColl))

	// get query parameters
	martID = ctx.Query("mart-id")
	query = ctx.Query("q")

	// get items with the query from DB
	filter := bson.M{"mart_id": martID, "$text": bson.M{"$search": query}}
	opts := options.Find().SetLimit(20)
	cursor, err := groceriesColl.Find(context.TODO(), filter, opts)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	if err = cursor.All(context.TODO(), &items); err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// transform items
	var transformedItems []map[string]interface{}
	for _, item := range items {
		transformedItem, err := blueprint.ItemTransform(item, blueprint.CustomerItem)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}
		transformedItems = append(transformedItems, transformedItem)
	}

	// response
	ctx.JSON(http.StatusOK, transformedItems)
}
