package api

import (
	"capszo-mart/middleware"
	"capszo-mart/token"
	"capszo-mart/util"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	authorizationPayloadKey = "authorization_payload"
)

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

	// general routes
	router.GET("/", server.home)

	// auth routes
	router.POST("/access-token/renew", server.renewToken)
	router.POST("/admin/login", server.getAdminLoginOTP)
	router.POST("/admin/login/otp", server.adminLogin)
	router.POST("/customer/signup", server.getCustomerSignupOTP)
	router.POST("/customer/signup/otp", server.customerSignup)
	router.POST("/customer/login", server.getCustomerLoginOTP)
	router.POST("/customer/login/otp", server.customerLogin)
	router.POST("/mart/login", server.getMartLoginOTP)
	router.POST("/mart/login/otp", server.martLogin)
	router.POST("/truck/login", server.getTruckLoginOTP)
	router.POST("/truck/login/otp", server.truckLogin)
	router.POST("/logout", server.logout)

	// customer routes
	authMiddleware := middleware.CustomerAuthMiddleware(server.token)
	customerRouter := router.Group("/").Use(authMiddleware)
	customerRouter.GET("/customer", server.getCustomerInfo)
	customerRouter.GET("/item/:id", server.getItem)
	customerRouter.GET("/items", server.getMartItems)
	customerRouter.GET("/items/search", server.searchItem)
	customerRouter.GET("/mart/order-preferences", server.getMartOrderPreference)
	customerRouter.GET("customer/order/:id", server.getOrder)
	customerRouter.GET("customer/orders", server.getOrders)
	customerRouter.POST("/order", server.order)
	// customerRouter.GET("/basket",)
	customerRouter.PUT("customer/basket", server.updateBasket)
	customerRouter.PUT("/customer/address", server.updateCustomerAddress)
	// customerRouter.PUT("/email", server.updateCustomerEmail)

	// mart routes
	authMiddleware = middleware.MartAuthMiddleware(server.token)
	martRouter := router.Group("/mart").Use(authMiddleware)
	martRouter.POST("/items", server.addItems)
	martRouter.PUT("/item", server.updateItem)
	martRouter.GET("/orders", server.getOrders)
	martRouter.PUT("/order/deliver", server.deliverOrder)

	// truck routes
	authMiddleware = middleware.TruckAuthMiddleware(server.token)
	truckRouter := router.Group("/truck").Use(authMiddleware)
	truckRouter.GET("/orders", server.getOrders)
	truckRouter.PUT("/order")

	// admin routes
	authMiddleware = middleware.AdminAuthMiddleware(server.token)
	adminRouter := router.Group("/admin").Use(authMiddleware)
	adminRouter.GET("/test-token", server.getTestToken)
	adminRouter.POST("/products", server.addProducts)
	adminRouter.GET("/product/:id", server.getProduct)
	adminRouter.GET("/products", server.searchProduct)
	adminRouter.PUT("/product", server.updateProduct)
	adminRouter.POST("/mart/signup", server.martSignup)
	adminRouter.POST("/truck/signup", server.truckSignup)
	adminRouter.GET("/order/:id", server.getOrder)
	adminRouter.GET("/orders", server.getOrders)
	adminRouter.POST("/items", server.addItems)
	adminRouter.PUT("/item", server.updateItem)
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
