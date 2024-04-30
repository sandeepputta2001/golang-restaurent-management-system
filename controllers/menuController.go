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

var menuCollection *mongo.Collection = database.OpenCollection(database.Client, "menu")

// @Summary        Returns slice of menus 
// @Description    Returns an array of menus from the menucollection in restaurent database.
// @Tags           menu
// @Produce        application/json
// @Security        @Security.require(true)
// @Success        200 {array} models.Menu "slice of menus "
// @Router         /menus [get]
func GetMenus() gin.HandlerFunc {
	return func(c *gin.Context) {

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		result, err := menuCollection.Find(context.TODO(), bson.M{})
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while listing the menu items"})
			return
		}

		var allMenus []bson.M

		if err := result.All(ctx, &allMenus); err != nil {
			log.Fatal(err)
		}

		c.JSON(http.StatusOK, allMenus)

	}
}

// @Summary        Retrieves a menu  with specific menu id
// @Description    Retrieves a menu with specific menu id from the menu collection
// @Tags           menu
// @Accept         application/json
// @Produce        application/json
// @Param          menu_id path string true "menu_id"
// @Security       @Security.require(true)
// @Success        200 {object} models.Menu "Details of a specific menu"
// @Failure        500 {string} http.StatusInternalServerError "Internal Server Error in mongodb"
// @Router         /menus/{menu_id} [get]
func GetMenu() gin.HandlerFunc {
	return func(c *gin.Context) {

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		menuId := c.Param("menu_id")
		var menu models.Menu
		err := menuCollection.FindOne(ctx, bson.M{"menu_id": menuId}).Decode(&menu)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while getting the menu item"})
			defer cancel()
			return
		}

		c.JSON(http.StatusOK, menu)

		defer cancel()

	}
}


// @Summary        Creates a new menu 
// @Description    Creates a menu resource on the server
// @Tags           menu
// @Accept         application/json
// @Produce        application/json
// @Param          Menu body models.Menu true  "Menu object"
// @Security       @Security.require(true)
// @Success        201 {object} models.Menu "New menu created" 
// @Failure        500 {string} http.StatusInternalServerError "Internal Server Error while creating a new menu"
// @Router         /menus  [post]
func CreateMenu() gin.HandlerFunc {
	return func(c *gin.Context) {

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		var menu models.Menu

		if err := c.BindJSON(&menu); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			defer cancel()
			return
		}

		validationErr := validate.Struct(menu)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			defer cancel()
			return

		}

		menu.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		menu.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		menu.ID = primitive.NewObjectID()
		menu.Menu_id = menu.ID.Hex()

		result, insertErr := menuCollection.InsertOne(ctx, menu)
		if insertErr != nil {
			msg := fmt.Sprintf("error occuring while creating a menu item in the menu collection")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			defer cancel()
			return
		}
		defer cancel()

		c.JSON(http.StatusOK, result)

	}
}

func inTimeSpan(start, end, check time.Time) bool {
	return start.After(time.Now()) && end.After(start)
}

// @Summary        Updates an menu resource
// @Description    Updates an existing menu resource in the menu collection
// @Tags           menu
// @Accept         application/json
// @Produce        application/json 
// @Param          menu_id path string true "ID of the menu resource to update"
// @Param          Menu body models.Menu true  "Menu object"
// @Security       @Security.require(true)
// @Success        200 {object} models.Menu "menu got updated with new body"
// @Failure        500 {string} http.StatusInternalServerError "Internal Server Error while updating food"
// @Failure        404 {string} http.StatusBadRequest "Bad Request"
// @Router         /menu/{menu_id} [patch]
func UpdateMenu() gin.HandlerFunc {
	return func(c *gin.Context) {

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		var menu models.Menu

		if err := c.BindJSON(&menu); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error ": err.Error()})
			defer cancel()
			return
		}

		menuId := c.Param("menu_id")

		filter := bson.M{"menu_id": menuId}

		var updateObj primitive.D //he primitive package is used for working with BSON (Binary JSON) data types, which are used for representing data in MongoDB documents.
		//The primitive.D type you've mentioned is used to represent a BSON document as an ordered list of key-value pairs, similar to a dictionary or map in other programming languages.

		if menu.Start_Date != nil && menu.End_Date != nil {
			if !inTimeSpan(*menu.Start_Date, *menu.End_Date, time.Now()) {
				msg := "Kindly retype the time"

				c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
				defer cancel()
				return
			}
		}

		updateObj = append(updateObj, bson.E{Key: "start_date", Value: menu.Start_Date})
		updateObj = append(updateObj, bson.E{Key: "end_date", Value: menu.End_Date})

		if menu.Name != "" {
				updateObj = append(updateObj, bson.E{Key: "name", Value: menu.Name})
			}

		if menu.Category != "" {
				updateObj = append(updateObj, bson.E{Key: "category", Value: menu.Category})
			}

		menu.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		updateObj = append(updateObj, bson.E{Key: "updated_at", Value: menu.Updated_at})

		upsert := true

		opt := options.UpdateOptions{
				Upsert: &upsert,
		}

		result, err := menuCollection.UpdateOne(
				ctx,
				filter,
				bson.D{
					{Key: "$set", Value: updateObj},
				},
				&opt,
		)

		if err != nil {
				msg := "Menu Update Failed"

				c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
				defer cancel()
				return
		}

	   c.JSON(http.StatusOK, result)
		

	}
}
