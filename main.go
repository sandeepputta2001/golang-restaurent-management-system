package main

import (
	middleware "go-restaurent-management-system/middleware"
	"go-restaurent-management-system/routes"
	"os"

	"github.com/gin-gonic/gin"

	//swagger
	_ "go-restaurent-management-system/docs"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title       go-backend
// @version     1.0
// @description CRUP api's for go backend

// @host       localhost:8000
// @BasePath   /
// @SecurityDefinitions.apiKey
//   @in header
//   @name token
//   @description JWT token for authentication
//   @type string
func main() {  

	port := os.Getenv("PORT") 

	if port == "" {
		port = "8000"
	} 
     
	router := gin.New()
	router.Use(gin.Logger())  

	router.LoadHTMLGlob("templates/*")

	router.GET("/swagger/*any",ginSwagger.WrapHandler(swaggerFiles.Handler)) 

	routes.HomeRoutes(router)
	routes.UserRoutes(router)

	// middleware
	router.Use(middleware.Authentication()) 

	routes.FoodRoutes(router) 
	routes.MenuRoutes(router)
	routes.TableRoutes(router)
	routes.OrderRoutes(router)
	routes.OrderItemRoutes(router)
	routes.InvoiceRoutes(router)

	router.Run(":" + port)

}
