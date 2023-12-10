package api

import (
	"capszo-mart/database"
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (server *Server) addProducts(ctx *gin.Context) {
	var request []database.Product
	var err error
	db := server.mongoDB.Database("capszo")
	productColl := db.Collection(string(database.ProductColl))

	// get request data
	if err = ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// set timings and change datatype to insert
	var products []interface{}
	currentTime := time.Now()
	for _, product := range request {
		product.CreatedAt = currentTime
		product.UpdatedAt = currentTime
		products = append(products, product)
	}

	// insert all the products
	result, err := productColl.InsertMany(context.TODO(), products)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// create text index
	model := mongo.IndexModel{
		Keys: bson.D{{Key: "name", Value: "text"}, {Key: "other_names", Value: "text"}},
	}
	_, err = productColl.Indexes().CreateOne(context.TODO(), model)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// response
	ctx.JSON(http.StatusCreated, gin.H{"message": "products added successfully", "product_ids": result.InsertedIDs})
}

func (server *Server) searchProduct(ctx *gin.Context) {
	var query string
	var products []database.Product
	var err error
	db := server.mongoDB.Database("capszo")
	productColl := db.Collection(string(database.ProductColl))

	// get query parameters
	query = ctx.Query("q")

	// get items with the query from DB
	filter := bson.M{"$text": bson.M{"$search": query}}
	opts := options.Find().SetLimit(20)
	cursor, err := productColl.Find(context.TODO(), filter, opts)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	if err = cursor.All(context.TODO(), &products); err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// response
	ctx.JSON(http.StatusOK, products)
}

func (server *Server) getProduct(ctx *gin.Context) {
	var product database.Product
	db := server.mongoDB.Database("capszo")
	productColl := db.Collection(string(database.ProductColl))

	// get id from query params
	productID := ctx.Param("id")

	// get order using _id
	objectID, err := primitive.ObjectIDFromHex(productID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// get product by id
	filter := bson.M{"_id": objectID}
	if err = productColl.FindOne(context.TODO(), filter).Decode(&product); err != nil {
		if mongo.ErrNoDocuments == err {
			err = errors.New("NO PRODUCT FOUND")
			ctx.JSON(http.StatusBadRequest, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// response
	ctx.JSON(http.StatusOK, product)
}

func (server *Server) updateProduct(ctx *gin.Context) {
	var request database.Product
	var err error
	db := server.mongoDB.Database("capszo")
	productColl := db.Collection(string(database.ProductColl))

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

	// update grocery items
	request.ID = objectID
	request.UpdatedAt = time.Now()
	update := bson.M{"$set": request}
	filter := bson.M{"_id": objectID}
	if _, err = productColl.UpdateOne(context.TODO(), filter, update); err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// response
	ctx.JSON(http.StatusOK, gin.H{"message": "product updated successfully", "product_id": request.ID})
}
