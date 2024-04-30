package routes

import (
	controllers "go-restaurent-management-system/controllers"

	"github.com/gin-gonic/gin"
)


func HomeRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.GET("/",controllers.HomeController())
} 