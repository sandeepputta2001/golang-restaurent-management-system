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

var tableCollection *mongo.Collection = database.OpenCollection(database.Client, "tables")

func GetTables() gin.HandlerFunc {
	return func(c *gin.Context) {

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		result, err := tableCollection.Find(ctx, bson.M{}) // empty bson.M{} indicates querying for all the records present in the collection .

		if err != nil {
			msg := fmt.Sprintf("error occured while finding the tables in the orderItem Collection")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			defer cancel()
			return
		}
		defer cancel()

		var allTables []bson.M

		if err := result.All(ctx, &allTables); err != nil {
			log.Fatal(err)
		}

		c.JSON(http.StatusOK, allTables)

	}
}

func GetTable() gin.HandlerFunc {
	return func(c *gin.Context) {

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		tableId := c.Param("table_id")

		var table models.Table

		err := tableCollection.FindOne(ctx, bson.M{"table_id": tableId}).Decode(&table) // empty bson.M{} indicates querying for all the records present in the collection .

		if err != nil {
			msg := fmt.Sprintf("error occured while finding the table in the table 	 Collection")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			defer cancel()
			return
		}
		defer cancel()

		c.JSON(http.StatusOK, table)

	}
}

func CreateTable() gin.HandlerFunc {
	return func(c *gin.Context) {

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		var table models.Table

		if err := c.BindJSON(&table); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			defer cancel()
			return
		}

		table.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		table.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

		table.ID = primitive.NewObjectID()
		table.Table_id = table.ID.Hex()

		result, insertErr := tableCollection.InsertOne(ctx, table)
		defer cancel()

		if insertErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while inserting document into the collection "})
			return

		}

		c.JSON(http.StatusOK, result)

	}
}

func UpdateTable() gin.HandlerFunc {
	return func(c *gin.Context) {

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		var table models.Table

		tableId := c.Param("table_id")
		filter := bson.M{"table_id": tableId}

		var updateObj primitive.D

		if err := c.BindJSON(&table); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			defer cancel()
			return
		}

		table.ID = primitive.NewObjectID()
		table.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

		if table.Number_of_guests != nil {
			updateObj = append(updateObj, bson.E{Key: "number_of_guests", Value: *table.Number_of_guests})

		}

		if table.Table_number != nil {
			updateObj = append(updateObj, bson.E{Key: "table_number", Value: table.Table_number})
		}

		upsert := true

		opts := options.UpdateOptions{
			Upsert: &upsert,
		}

		result, err := tableCollection.UpdateOne(
			ctx,
			filter,
			bson.D{
				{Key: "$set", Value: updateObj},
			},
			&opts,
		)

		defer cancel()

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while updating the document"})
			return
		}

		c.JSON(http.StatusOK, result)

	}
}
