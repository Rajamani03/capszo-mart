package api

import (
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

	// response
	ctx.JSON(http.StatusOK, items)
}

type inputItem struct {
	ID              string    `json:"item_id" bson:"_id,omitempty"`
	MartID          string    `json:"mart_id" bson:"mart_id"`
	Name            string    `json:"name" bson:"name"`
	ImageURLs       []string  `json:"image_urls" bson:"image_urls"`
	Mrp             float64   `json:"mrp" bson:"mrp"`
	SellingPrice    float64   `json:"selling_price" bson:"selling_price"`
	CostPrice       float64   `json:"cost_price" bson:"cost_price"`
	Quantity        float64   `json:"quantity" bson:"quantity"`
	Unit            string    `json:"unit" bson:"unit"`
	StepQuantity    float32   `json:"step_quantity" bson:"step_quantity"`
	IndividualLimit float64   `json:"individual_limit" bson:"individual_limit"`
	StockQuantity   float64   `json:"stock_quantity" bson:"stock_quantity"`
	Brand           string    `json:"brand" bson:"brand"`
	Category        string    `json:"category" bson:"category"`
	SubCategory     string    `json:"sub_category" bson:"sub_category"`
	OtherNames      []string  `json:"other_names" bson:"other_names"`
	CreatedAt       time.Time `json:"-" bson:"created_at"`
	UpdatedAt       time.Time `json:"-" bson:"updated_at"`
}

func (server *Server) addItems(ctx *gin.Context) {
	var request []inputItem
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
	var request inputItem
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
	objectID, err := primitive.ObjectIDFromHex(request.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	filter := bson.M{"_id": objectID, "mart_id": tokenPayload.UserID}
	if tokenPayload.TokenFor == token.AdminAccess {
		filter = bson.M{"_id": objectID}
	}

	// update grocery items
	fmt.Printf("filter: %v\n", filter)
	request.UpdatedAt = time.Now()
	update := bson.M{"$set": request}
	result, err := groceriesColl.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	fmt.Printf("result.MatchedCount: %v\n", result.MatchedCount)

	// response
	ctx.JSON(http.StatusCreated, gin.H{"message": "items updated successfully", "item_id": result.UpsertedID})
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
	// response
	ctx.JSON(http.StatusOK, items)
}
