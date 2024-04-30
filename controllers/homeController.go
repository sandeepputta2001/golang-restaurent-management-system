package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// @Summary                Returns home page
// @Description            Returns home page with introduction and guide
// @Produce                html
// @Success                200 {string}  "HTML content"
// @Router                 /home [get]
func HomeController() gin.HandlerFunc{

	return func(c  *gin.Context) {
 

		c.HTML(http.StatusOK,"index.html", gin.H{
			"title":"Hello Amigos!!!!!!",
			"description":"Looking for API's to build user interface for a restaurant Management system",
		}) 

	}
} 