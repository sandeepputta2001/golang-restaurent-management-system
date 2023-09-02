package routes

import (
	controller "go-restaurent-management-system/controllers"

	"github.com/gin-gonic/gin"
)

func OrderItemRoutes(incomingRoutes *gin.Engine) {

	incomingRoutes.GET("/orderitems", controller.GetOrderItems())
	incomingRoutes.GET("/orderitems/:orderitem_id", controller.GetOrderItem())
	incomingRoutes.POST("/orderitems", controller.CreateOrderItem())
	incomingRoutes.PATCH("/orderitems/:orderitem_id", controller.UpdateOrderItem())
	incomingRoutes.GET("/orderItems-order/:order_id", controller.GetOrderItemsByOrder())
}
