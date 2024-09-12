package controllers

import (
	"context"
	"fmt"
	"go-restaurent-management-system/database"
	"go-restaurent-management-system/models"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var foodCollection *mongo.Collection = database.OpenCollection(database.Client, "foods")
var validate = validator.New() 

// @Summary              Returns slice of foods
// @Description          Returns an array of foods from the ordercollection in restaurent database.
// @Tags                 foods
// @Security             @Security.require(true)
// @Produce              application/json
// @Success              200 {array} models.Food "slice of orders"
// @Router               /foods [get]
func GetFoods() gin.HandlerFunc {
	return func(c *gin.Context) { 

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		recordPerPage, err := strconv.Atoi(c.Query("recordPerPage"))
		if err != nil || recordPerPage < 1 { 
			recordPerPage = 10

		} 

		page, err := strconv.Atoi(c.Query("page"))
		if err != nil {
			page = 1
		} 

		startIndex := (page - 1) * recordPerPage

		matchStage := bson.D{{Key: "$match", Value: bson.D{{}}}}
		groupStage := bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: bson.D{
				{Key: "_id", Value: "null"}}}, 
			{Key: "total_count", Value: bson.D{{Key: "$sum", Value: 1}}}, 
			{Key: "data", Value: bson.D{{Key: "$push", Value: "$$ROOT"}}}}}} 
			
		projectStage := bson.D{
			{
				Key: "$project", Value: bson.D{
					{Key: "_id", Value: 0},
					{Key: "total_count", Value: 1},
					{Key: "food_items", Value: bson.D{{Key: "$slice", Value: []interface{}{"$data", startIndex, recordPerPage}}}},
					{Key: "data" , Value: 1},
				}}}

		result, err := foodCollection.Aggregate(ctx, mongo.Pipeline{
			matchStage, groupStage, projectStage,
		})

		
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while listing food items"})
			return
		}

		var allFoods []bson.M 

		if err := result.All(ctx, &allFoods); err != nil {
			log.Fatal(err)
		}

		c.JSON(http.StatusOK, allFoods) 
	} 
}

// @Summary              Retrieves a food with specific food id
// @Description          Retrieves  a  food with specific food id from the  orders collection
// @Tags                 foods
// @Accept               application/json
// @Produce              application/json
// @Param                food_id path string true "food_id"
// @Security             @Security.require(true)
// @Success              200 {object} models.Food "Details of a speicific food"
// @Failure              500 {string} http.StatusInternalServerError "Internal Server Error in mongodb"
// @Router               /foods/{food_id} [get]
func GetFood() gin.HandlerFunc {
	return func(c *gin.Context) {

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		foodid := c.Param("food_id")

		var food models.Food

		err := foodCollection.FindOne(ctx, bson.M{"food_id": foodid}).Decode(&food)
		defer cancel()

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while fetching the food item"})
			return
		}

		c.JSON(http.StatusOK, food)

	}
}

// @Summary              Creates a food  resource
// @Description          Creates a food resource on the server
// @Tags                 foods
// @Accept               application/json
// @Produce              application/json
// @Param                food body models.Food true "Food object"
// @Security             @Security.require(true)
// @Success              201 {object} models.Food "New food created" 
// @Failure              500 {string} http.StatusInternalServerError "Internal Server Error while creating a new food"
// @Router               /foods  [post]
func CreateFood() gin.HandlerFunc {	
	return func(c *gin.Context) {

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)

		var food models.Food
		var menu models.Menu

		if err := c.BindJSON(&food); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		}

		validationErr := validate.Struct(food)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})

		}

		err := menuCollection.FindOne(ctx, bson.M{"menu_id": food.Menu_id}).Decode(&menu)
		defer cancel()
		if err != nil {
			msg := fmt.Sprintf("menu was not found")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		food.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339)) //time.RFC3339: This is a predefined constant in the time package that represents the layout of a timestamp in the RFC3339 format. RFC3339 is a standardized timestamp format that includes both date and time information.
		//time.Parse(time.RFC3339, time.Now()): Here, time.Parse is used to parse a time string using the specified layout (RFC3339) and the current time as a string. The result of this operation is a time.Time object representing the parsed timestamp.
		//.Format(time.RFC3339): The Format method is used on the time.Time object to convert it back to a string representation in the RFC3339 format. This is done to ensure that the timestamp is in the expected format before assigning it to the Created_at and Updated_at fields.
		food.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		food.ID = primitive.NewObjectID()
		food.Food_id = food.ID.Hex() // converting the ID created in the collection into understandable hexadecimal which comprises of 0 to 9 numbers and a to f alphabets
		var num = toFixed(*food.Price, 2)
		food.Price = &num

		result, insertErr := foodCollection.InsertOne(ctx, food)
		if insertErr != nil {
			msg := fmt.Sprintf("food item was not created ")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		defer cancel()
		c.JSON(http.StatusOK, result)

	}
}

// @Summary              Updates an food  resource
// @Description          Updates an existing fodd  resource in the foods collection
// @Tags                 foods
// @Accept               application/json
// @Produce              application/json 
// @Param                food_id path string true "ID of the food resource to update"
// @Security             @Security.require(true)
// @Param                food body models.Food true "Food object"
// @Success              200 {object} models.Food "food got updated with new body"
// @Failure              500 {string} http.StatusInternalServerError "Internal Server Error while updating food"
// @Failure              404 {string} http.StatusBadRequest "Bad Request"
// @Router               /foods/{food_id} [patch]
func UpdateFood() gin.HandlerFunc {
	return func(c *gin.Context) {

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		var menu models.Menu
		var food models.Food

		foodId := c.Param("food_id")

		if err := c.BindJSON(&food); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error ": err.Error()})
			defer cancel()
			return
		}

		var updateObj primitive.D

		if food.Name != nil {
			updateObj = append(updateObj, bson.E{Key: "name", Value: food.Name})
		}

		if food.Price != nil {
			updateObj = append(updateObj, bson.E{Key: "price", Value: food.Price})
		}

		if food.Food_image != nil {
			updateObj = append(updateObj, bson.E{Key: "food_image", Value: food.Food_image})
		}

		if food.Menu_id != nil {
			err := menuCollection.FindOne(ctx, bson.M{"menu_id": food.Menu_id}).Decode(&menu)
			defer cancel()
			if err != nil {
				msg := fmt.Sprintf("message : menu was not found")
				c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
				return
			}
			updateObj = append(updateObj, bson.E{Key: "menu_id", Value: food.Menu_id})
		}

		food.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		updateObj = append(updateObj, bson.E{Key: "updated_at", Value: food.Updated_at})

		upsert := true

		filter := bson.M{"food_id": foodId}

		opt := options.UpdateOptions{
			Upsert: &upsert,
		}

		result, err := foodCollection.UpdateOne(
			ctx,
			filter,
			bson.D{
				{Key: "$set", Value: updateObj},
			},
			&opt,
		)

		if err != nil {
			msg := fmt.Sprintf("error occured while updating the document in the collection")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			defer cancel()
			return
		}

		defer cancel()

		c.JSON(http.StatusOK, result)

	}
}

func round(num float64) int { 

	return int(num + math.Copysign(0.5, num)) //math.Copysign(0.5, num) returns a positive or negative 0.5 based on the sign of num. If num is positive, it returns 0.5; if num is negative, it returns -0.5.

}

func toFixed(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision)) //precision is the desired number of decimal places to keep after rounding.
	return float64(round(num*output)) / output

} 

