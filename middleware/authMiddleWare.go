package middleware

import (
	"fmt"
	"net/http"

	helper "go-restaurent-management-system/helpers"

	"github.com/gin-gonic/gin"
)

// authentication function is used to authenticate jwt token and set keys from jwt signed details in this request for
//for further use in request handlers.
func Authentication() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientToken := c.Request.Header.Get("token")
		if clientToken == "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error ": fmt.Sprintf("No authorisation Header Provided")})
			c.Abort()
			return
		}  

		claims, err := helper.ValidateToken(clientToken)

		if err != "" { 
			c.JSON(http.StatusInternalServerError, gin.H{"error": err}) 
			c.Abort() //In Go, c.Abort() is commonly associated with web application frameworks like Gin and Echo, and it's used to prematurely terminate the processing of a request and immediately return a response to the client without allowing further middleware functions or request handlers to execute.
			return
		}

		c.Set("email", claims.Email) 
		c.Set("first_name", claims.First_name)
		c.Set("last_name", claims.Last_name)
		c.Set("uid", claims.Uid)

		c.Next() // In the context of web frameworks like Gin or Echo in Go, c.Next() is used to instruct the framework to continue processing the current HTTP request by calling the next middleware function or the next route handler in the chain. It allows you to delegate control to the next piece of middleware or the next handler in line.

	}
}
