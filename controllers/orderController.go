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

func CreateOrder() gin.HandlerFunc {
	return func(c *gin.Context) {

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		var table models.Table
		var order models.Order

		if order.Table_id != nil {

			err := tableCollection.FindOne(ctx, bson.M{"table_id": order.Table_id}).Decode(&table)
			defer cancel()
			if err != nil {
				msg := fmt.Sprintf("error occured while finding the table ")
				c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
				return
			}
		}

		order.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		order.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

		order.ID = primitive.NewObjectID()
		order.Order_id = order.ID.Hex()

		result, err := orderCollection.InsertOne(ctx, order)
		defer cancel()

		if err != nil {
			msg := fmt.Sprintf("error occured while inserting document")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		c.JSON(http.StatusOK, result)

	}
}

func UpdateOrder() gin.HandlerFunc {
	return func(c *gin.Context) {

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var order models.Order

		if err := c.BindJSON(&order); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		}

		orderId := c.Param("order_id")

		var updateObj primitive.D

		if order.Table_id != nil {
			err := orderCollection.FindOne(ctx, bson.M{"table_id": order.Table_id}).Decode(&order)
			defer cancel()
			if err != nil {
				msg := fmt.Sprintf("error:Menu was not found")
				c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
				return

			}
			updateObj = append(updateObj, bson.E{Key: "menu", Value: order.Table_id})
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

		defer cancel()

		if err != nil {
			msg := fmt.Sprintf("error occured while updating the document")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		c.JSON(http.StatusOK, result)

	}
}

func OrderItemOrderCreator(order models.Order) string {

	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

	order.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	order.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

	order.ID = primitive.NewObjectID()
	order.Order_id = order.ID.Hex()

	orderCollection.InsertOne(ctx, order)

	defer cancel()

	return order.Order_id
}
