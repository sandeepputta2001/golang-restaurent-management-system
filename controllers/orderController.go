package controllers

import (
	"context"
	"fmt"
	"go-restaurent-management-system/database"
	"go-restaurent-management-system/models"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var orderCollection *mongo.Collection = database.OpenCollection(database.Client, "orders")


// @Summary       Returns slice of orders
// @Description   Returns an array of orders placed from the ordercollection in restaurent database.
// @Tags          order
// @Security      @Security.require(true)
// @Produce       application/json 
// @Success       200 {array} models.Order	"slice of orders "
// @Router        /orders [get]
func GetOrders() gin.HandlerFunc {	
	return func(c *gin.Context) {

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		result, err := orderCollection.Find(ctx, bson.M{})
		if err != nil {
			msg := fmt.Sprintf("error occured while finding the orders in the orderCollection")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})

		}
		defer cancel()

		var allOrders []bson.M

		if err := result.All(ctx, &allOrders); err != nil {
			log.Fatal(err)
		}

		c.JSON(http.StatusOK, allOrders)

	}
}


// @Summary       Retrieves a order with specific order id
// @Description   Retrieves a order with specific order id from the orders collection
// @Tags          order
// @Accept        application/json
// @Produce       application/json 
// @Param         order_id path string true "order_id"
// @Security      @Security.require(true)
// @Success       200 {object} models.Order "Details of a specific order"
// @Failure       500 {string} http.StatusInternalServerError "Internal Server Error in mongodb"
// @Router        /orders/{order_id} [get]
func GetOrder() gin.HandlerFunc {
	return func(c *gin.Context) {

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		orderId := c.Param("order_id")

		var order models.Order

		err := orderCollection.FindOne(ctx, bson.M{"order_id": orderId}).Decode(&order)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while finding for the order "})
			return

		}

		c.JSON(http.StatusOK, order)

	}
}


// @Summary       Creates a order  resource
// @Description   Creates a order resource on the server
// @Tags          order
// @Accept        application/json
// @Produce       application/json
// @Param         Order body models.Order true "Order object"
// @Security      @Security.require(true)
// @Success       201 {object} models.Order "New order created"
// @Failure       500 {string} http.StatusInternalServerError "Internal Server Error while creating a new order"
// @Router        /orders [post]
func CreateOrder() gin.HandlerFunc {
	return func(c *gin.Context) {
		var table models.Table
		var order models.Order

		ctx , cancel := context.WithTimeout(context.Background() , 100*time.Second)
		defer cancel()

		if err := c.BindJSON(&order); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		validationErr := validate.Struct(order)

		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

		if order.Table_id != nil {
			err := tableCollection.FindOne(ctx, bson.M{"table_id": order.Table_id}).Decode(&table)
			defer cancel()
			if err != nil {
				msg := fmt.Sprintf("message:Table was not found")
				c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
				return
			}
		}

		order.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		order.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

		order.ID = primitive.NewObjectID()
		order.Order_id = order.ID.Hex()

		result, insertErr := orderCollection.InsertOne(ctx, order)

		if insertErr != nil {
			msg := fmt.Sprintf("order item was not created")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		defer cancel()
		c.JSON(http.StatusOK, result)
	}
}

// @Summary       Updates an order  resource
// @Description   Updates an existing order resource in the orders collection
// @Tags          order
// @Accept        application/json
// @Produce       application/json 
// @Param         order_id path string true "ID of the order resource to update"
// @Security      @Security.require(true)
// @Param         Order body models.Order true  "Order object"
// @Success       200 {object} models.Order "orders got updated with new body"
// @Failure       500 {string} http.StatusInternalServerError "Internal Server Error while updating order"
// @Failure       404 {string} http.StatusBadRequest "Bad Request"
// @Router        /orders/{order_id} [patch]
func UpdateOrder() gin.HandlerFunc {
	return func(c *gin.Context) {
		var table models.Table
		var order models.Order

		var updateObj primitive.D 

		ctx , cancel := context.WithTimeout(context.Background() , 100*time.Second)
		defer cancel()

		orderId := c.Param("order_id")
		if err := c.BindJSON(&order); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if order.Table_id != nil {
			err := tableCollection.FindOne(ctx, bson.M{"table_id": order.Table_id}).Decode(&table)
			if err != nil {
				msg := fmt.Sprintf("Table was not found") 
				c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
				return
			}
			updateObj = append(updateObj, bson.E{Key: "table_id", Value: order.Table_id})
		}

		order.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		updateObj = append(updateObj, bson.E{Key: "updated_at", Value: order.Updated_at})

		upsert := true

		filter := bson.M{"order_id": orderId}
		opt := options.UpdateOptions{
			Upsert: &upsert,
		}

		result, err := orderCollection.UpdateOne(
			ctx,
			filter,
			bson.D{
				{Key: "$set", Value: updateObj},
			},
			&opt,
		)

		if err != nil {
			msg := fmt.Sprintf("order item update failed")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

func OrderItemOrderCreator(order models.Order) string {   

	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	order.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	order.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

	order.ID = primitive.NewObjectID() 
	order.Order_id = order.ID.Hex()

	orderCollection.InsertOne(ctx, order) 

	return order.Order_id 
} 
