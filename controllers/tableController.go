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

// @Summary        Returns slice of tables
// @Description    Returns an array of tables placed from table collectios in restaurent database.
// @Tags           table
// @Produce        application/json
// @Security       @Security.require(true)
// @Success        200 {array} models.Table	"slice of tables "
// @Router         /tables [get] 
func GetTables() gin.HandlerFunc {
	return func(c *gin.Context) { 

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
        defer cancel() 
		result, err := tableCollection.Find(ctx, bson.M{}) // empty bson.M{} indicates querying for all the records present in the collection .

		if err != nil {
			msg := fmt.Sprintf("error occured while finding the tables in the orderItem Collection")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}
		

		var allTables []bson.M

		if err := result.All(ctx, &allTables); err != nil {
			log.Fatal(err)
		}

		c.JSON(http.StatusOK, allTables)

	}
}

// @Summary        Retrieves a table  with specific table id
// @Description    Retrieves a table with specific table id from the tables collection
// @Tags           table
// @Accept         application/json
// @Produce        application/json
// @Param          table_id path string true "table_id"
// @Security       @Security.require(true)
// @Success        200 {object} models.Table "table details of specific userid"
// @Failure        500 {string} http.StatusInternalServerError "Internal Server Error in mongodb"
// @Router         /tables/{table_id} [get] 
func GetTable() gin.HandlerFunc {
	return func(c *gin.Context) {

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
        defer cancel()
		tableId := c.Param("table_id")

		var table models.Table

		err := tableCollection.FindOne(ctx, bson.M{"table_id": tableId}).Decode(&table) // empty bson.M{} indicates querying for all the records present in the collection .

		if err != nil {
			msg := fmt.Sprintf("error occured while finding the table in the table 	 Collection")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}
		c.JSON(http.StatusOK, table)

	}
}

// @Summary        Creates a table  resource
// @Description    Creates a new table resource on the server
// @Tags           table
// @Accept         application/json
// @Produce        application/json
// @Param          table body models.Table true  "Table object"
// @Security       @Security.require(true)
// @Success        201 {object} models.Table "New table created"
// @Failure        500 {string} http.StatusInternalServerError "Internal Server Error while creating a new table"
// @Router         /tables [post]
func CreateTable() gin.HandlerFunc {
	return func(c *gin.Context) {

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var table models.Table

		if err := c.BindJSON(&table); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			
			return
		}

		table.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		table.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

		table.ID = primitive.NewObjectID()
		table.Table_id = table.ID.Hex()

		result, insertErr := tableCollection.InsertOne(ctx, table)

		if insertErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while inserting document into the collection "})
			return

		}

		c.JSON(http.StatusOK, result)

	}
}

// @Summary        Updates an table resource
// @Description    Updates an existing table resource in the table collection
// @Tags           table
// @Accept         application/json
// @Produce        application/json 
// @Param          table_id path string true "ID of the table resource to update"
// @Param          table body models.Table true  "Table object"
// @Security       @Security.require(true)
// @Success        200 {object} models.Table "table got updated with new body"
// @Failure        500 {string} http.StatusInternalServerError "Internal Server Error while updating table"
// @Failure        404 {string} http.StatusBadRequest "Bad Request"
// @Router         /tables/{table_id} [patch]
func UpdateTable() gin.HandlerFunc {
	return func(c *gin.Context) {

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var table models.Table

		tableId := c.Param("table_id") 
		filter := bson.M{"table_id": tableId}

		var updateObj primitive.D

		if err := c.BindJSON(&table); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			
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
