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

	// customer unauthorized routes
	router.POST("/customer/signup/get-otp", server.getCustomerSignupOTP)
	router.POST("/customer/signup", server.customerSignup)
	router.POST("/customer/login/get-otp", server.getCustomerLoginOTP)
	router.POST("/customer/login", server.customerLogin)

	// customer authorized routes
	authMiddleware := middleware.CustomerAuthMiddleware(server.token)
	customerRouter := router.Group("/").Use(authMiddleware)
	customerRouter.GET("/items/:mart-id", server.getAllItems)
	customerRouter.POST("/order", server.order)
	customerRouter.PUT("/customer/basket", server.updateBasket)
	customerRouter.PUT("/customer/address", server.updateCustomerAddress)
	// customerRouter.PUT("/email", server.updateCustomerEmail)

	// mart routes
	authMiddleware = middleware.MartAuthMiddleware(server.token)
	martRouter := router.Group("/mart")
	martRouter.POST("/items", server.addItems).Use(authMiddleware)
	// martRouter.GET("/orders").Use(authMiddleware)

	// truck routes
	// authMiddleware = middleware.TruckAuthMiddleware(server.token)
	// truckRouter := router.Group("/truck")
	// truckRouter.POST("/truck/login/get-otp", server.getCustomerLoginOTP)
	// truckRouter.POST("/truck/login", server.customerLogin)
	// truckRouter.PATCH("/order").Use(authMiddleware)

	// mart routes
	// haulerRouter := router.Group("/hauler")
	// haulerRouter.POST()

	// admin routes
	authMiddleware = middleware.AdminAuthMiddleware(server.token)
	adminRouter := router.Group("/admin")
	adminRouter.GET("/test-token", server.getTestToken)
	adminRouter.POST("/items", server.addItems).Use(authMiddleware)
	// adminRouter.POST("/truck/signup", server.customerSignup)
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
