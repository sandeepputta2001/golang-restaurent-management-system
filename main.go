package main

import (
	"go-restaurent-management-system/database"
	middleware "go-restaurent-management-system/middleware"
	"go-restaurent-management-system/routes"
	"os"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

var foodCollection *mongo.Collection = database.OpenCollection(database.Client, "food")

func main() {

	port := os.Getenv("PORT")

	if port == "" {
		port = "8000"
	}

	router := gin.New()
	router.Use(gin.Logger()) //gin.Logger(): This is a predefined middleware provided by the gin framework. It's a logging middleware that automatically logs information about incoming requests and outgoing responses. When this middleware is used, it will log details such as the HTTP method, URL, status code, and request processing time for each request.
	router.Use(middleware.Authentication())

	routes.UserRoutes(router)
	routes.FoodRoutes(router)
	routes.MenuRoutes(router)
	routes.TableRoutes(router)
	routes.OrderRoutes(router)
	routes.OrderItemRoutes(router)
	routes.InvoiceRoutes(router)

	router.Run(":" + port)

}
