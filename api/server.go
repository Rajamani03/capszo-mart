package api

import (
	"capszo-mart/middleware"
	"capszo-mart/token"
	"capszo-mart/util"
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	authorizationPayloadKey = "authorization_payload"
)

var bgctx = context.Background()

type Server struct {
	router  *gin.Engine
	mongoDB *mongo.Client
	token   token.Maker
	config  util.Config
}

func NewServer(mongoDB *mongo.Client, config util.Config) (*Server, error) {
	var server *Server
	var err error

	tokenMaker, err := token.NewPasetoMaker(config.TokenKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}

	server = &Server{
		mongoDB: mongoDB,
		token:   tokenMaker,
		config:  config,
	}

	server.SetupRouter()
	return server, nil
}

func (server *Server) SetupRouter() {
	var router *gin.Engine = gin.Default()

	// unauthorized routes
	router.GET("/", server.home)
	router.POST("/access-token/renew", server.renewToken)
	router.POST("/customer/signup/get-otp", server.getCustomerSignupOTP)
	router.POST("/customer/signup", server.customerSignup)
	router.POST("/customer/login/get-otp", server.getCustomerLoginOTP)
	router.POST("/customer/login", server.customerLogin)

	// customer authorized routes
	customerRouter := router.Group("/customer").Use(middleware.CustomerAuthMiddleware(server.token))
	customerRouter.GET("/grocery-items/:mart-id", server.getAllGroceryItems)
	// customerRouter.PUT("/grocery-basket", server.updateGroceryBasket)
	// customerRouter.PUT("/address", server.updateCustomerAddress)
	// customerRouter.POST("/grocery-order", server.groceryOrder)
	// // customerRouter.PUT("/email", server.updateCustomerEmail)

	// admin routes
	adminRouter := router.Group("/admin")
	adminRouter.GET("/test-token", server.getTestToken)
	// adminRouter.GET("/customers", server.getAllCustomers)

	// add router to server struct
	server.router = router
}

func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}

func toString(data interface{}) string {
	return fmt.Sprintf("%v", data)
}

func getID(objectID interface{}) string {
	return objectID.(primitive.ObjectID).Hex()
}

func (server *Server) home(ctx *gin.Context) {
	// response
	ctx.JSON(http.StatusOK, gin.H{"message": "Welcome to Capszo Mart", "environment": gin.Mode()})
}
