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

type OrderItemPack struct { 
	Table_id    *string            `json:"table_id" validate:"required"`
	Order_items []models.OrderItem `json:"order_items" validate:"required"`
} 

var orderItemCollection *mongo.Collection = database.OpenCollection(database.Client, "orderItem")

// @Summary         Returns slice of orderitems
// @Description     Returns an array of orderitems placed  from the orderitemcollection in restaurent database.
// @Tags            orderitem
// @Produce         application/json
// @Security        @Security.require(true)
// @Success         200 {array} models.OrderItem "slice of orderitems "
// @Router          /orderitems [get] 
func GetOrderItems() gin.HandlerFunc { 
	return func(c *gin.Context) { 

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		result, err := orderItemCollection.Find(ctx, bson.M{}) // empty bson.M{} indicates querying for all the records present in the collection .

		if err != nil { 
			msg := fmt.Sprintf("error occured while finding the orders in the orderItem Collection")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		} 

		var allOrderItems []bson.M 

		if err := result.All(ctx, &allOrderItems); err != nil {
			log.Fatal(err)
		}

		c.JSON(http.StatusOK, allOrderItems)

	} 
}

// @Summary         Retrieves orderitems placed from the specific order
// @Description     Retrieves orderitems placed from the specific order
// @Tags            orderitem
// @Accept          application/json
// @Produce         application/json
// @Param           order_id path string true "order_id"
// @Security        @Security.require(true)
// @Success         200 {object} models.OrderItem "retrieves orderitems placed from specific orderid"
// @Failure         500 {string} http.StatusInternalServerError "Internal Server Error in mongodb"
// @Router          /orderitemsorder/{order_id} [get]
func GetOrderItemsByOrder() gin.HandlerFunc { 
	return func(c *gin.Context) {

		orderId := c.Param("order_id")

		allOrderItems, err := ItemsByOrder(orderId)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		} 

		c.JSON(http.StatusOK, allOrderItems)

	}
}

func ItemsByOrder(orderId string) (OrderItems []bson.M, err error) {

	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second) 
	defer cancel()

	foodMatchStage := bson.D{ 
		{Key: "$match", Value: bson.D{ 
			{Key: "order_id", Value: orderId}, 
		 },
		},
	} 
	
	// matchStage : obtains the documents which are matched with the given query
	foodLookUpStage := bson.D{
		{Key: "$lookup", Value: bson.D{ 
			{Key: "from", Value: "foods"},
			{Key: "localfield", Value: "food_id"}, 
			{Key: "foreignfield", Value: "food_id"}, 
			{Key: "as", Value: "matchedFoods"},
		 }, 
		},
	} 


	// lookup stage : it is used to run the query to obtain the fiels of other document using the foreign key field of current document .
	// from : used to get the data from the other documents
	foodUnwindStage := bson.D{ 
		{Key: "$unwind", Value: bson.D{
			{Key: "path", Value: "$matchedFoods"}, 
			{Key: "preserveNullAndEmptyArrays", Value: true},

		 },
		},
	} 

    // afte lookup stage , the result comes as array , with which mongodb won't run . Therefore  , unwind stage is used to convert array form into some other form
	// path : this keyword gets the document on which unwinding has to be performed

	orderLookUpStage := bson.D{ 
		{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "orders"}, 
			{Key: "localfield", Value: "order_id"}, 
			{Key: "foreignfield", Value: "order_id"}, 
			{Key: "as", Value: "matchedOrders"}, 
		 },
		},
	}


	orderUnwindStage := bson.D{ 
		{Key: "$unwind", Value: bson.D{
			{Key: "path", Value: "$matchedOrders"}, 
			{Key: "preserveNullAndEmptyArrays", Value: true},
		 },
		},
	}

	tableLookUpStage := bson.D{
		{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "tables"}, 
			{Key: "localfield", Value: "matchedOrders.table_id"}, 
			{Key: "foreignfield", Value: "table_id"}, 
			{Key: "as", Value: "matchedTables"}, 
		  },
		},
	}


	tableUnwindStage := bson.D{
		{Key: "$unwind", Value: bson.D{
			{Key: "path", Value: "$matchedTables"}, 
			{Key: "preserveNullAndEmptyArrays", Value: true},
		},
		},
	}


	// projectStage is basically to manage the fields to be sent to frontend


	projectStage := bson.D{ 

		{Key: "$project", Value: bson.D{  
			{Key: "id", Value: 0},
			{Key: "amount", Value: "$matchedFoods.price"},
			{Key: "totat_count", Value: 1},
			{Key: "food_name", Value: "$matchedFoods.name"},
			{Key: "food_image", Value: "$matchedFoods.food_image"},
			{Key: "table_number", Value: "$matchedTables.table_number"},
			{Key: "table_id", Value: "$matchedTables.table_id"},
			{Key: "order_id", Value: "$matchedOrders.order_id"},
			{Key: "price", Value: "$matchedOrders.price"},
			{Key: "quantity", Value: 1},
		  },
	    },
	} 

	groupStage := bson.D{  
		{Key: "$group", Value: bson.D{
			{Key: "_id", Value: bson.D{
				{Key: "order_id", Value: "$order_id"}, 
				{Key: "table_id", Value: "$table_id"}, 
				{Key: "table_number", Value: "$table_number"},
			 },
			}, 
		    {Key: "payment_due", Value: bson.D{
			   {Key: "$sum", Value: "$amount"},
			 },
			},
		    {Key: "total_count", Value: bson.D{
			   {Key: "$sum", Value: 1},
			 },
			}, 
		    {Key: "order_items", Value: bson.D{
			  {Key: "$push", Value: "$$ROOT"},
		     },
	        },
          },
        },
    }

	projectStage2 := bson.D{
		{Key: "$project", Value: bson.D{

			{Key: "id", Value: 0},
			{Key: "payment_due", Value: 1},
			{Key: "total_count", Value: 1},
			{Key: "table_number", Value: "$_id.table_number"},
			{Key: "order_items", Value: 1},
		  },
	    },
    }

	result, err := orderItemCollection.Aggregate(ctx, mongo.Pipeline{ 
		foodMatchStage,
		foodLookUpStage,
		foodUnwindStage,
		orderLookUpStage,
		orderUnwindStage,
		tableLookUpStage,
		tableUnwindStage,
		projectStage,
		groupStage,
		projectStage2,
	})

	if err != nil {
		panic(err)
	}

	var orderItems []bson.M



	if err = result.All(ctx, &orderItems); err != nil {
		panic(err)				
	}

	

	return orderItems, err

} 

// @Summary         Retrieves a orderitem with specific orderitem id
// @Description     Retrieves a orderitem with specific orderitem id from the orderitem collection
// @Tags            orderitem
// @Accept          application/json
// @Produce         application/json
// @Param           orderitem_id path string true "orderitem_id"
// @Security        @Security.require(true)
// @Success         201 {object} models.OrderItem "Details of a specific orderitem id"
// @Failure         500 {string} http.StatusInternalServerError "Internal Server Error in mongodb"
// @Router          /orderitems/{orderitem_id} [get]
func GetOrderItem() gin.HandlerFunc {
	return func(c *gin.Context) {

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		orderItemId := c.Param("order_item_id")


		var orderItem models.OrderItem 

		err := orderItemCollection.FindOne(ctx, bson.M{"order_item_id": orderItemId}).Decode(&orderItem)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "orderitem is not found in the orderitem collection"})
			return
		}

		c.JSON(http.StatusOK, orderItem)

	}
}

// @Summary         Creates a orderitem  resource
// @Description     Creates a new orderitem resource on the server
// @Tags            orderitem
// @Accept          application/json
// @Produce         application/json
// @Param           orderitme body models.OrderItem true "Orderitem object"
// @Security        @Security.require(true)
// @Success         201 {object} models.OrderItem "New orderitem created"
// @Failure         500 {string} http.StatusInternalServerError "Internal Server Error while creating a new orderitem"
// @Router          /orderitems [post]
func CreateOrderItem() gin.HandlerFunc { 
	return func(c *gin.Context) {

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var orderItemPack OrderItemPack 
		var order models.Order 

		if err := c.BindJSON(&orderItemPack); err != nil { 
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
		} 

		orderItemsToBeInserted := []interface{}{} // It appears to be creating an empty slice named orderItemsToBeInserted with the type []interface{
		order.Order_date, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		order.Table_id = orderItemPack.Table_id  

		order_id := OrderItemOrderCreator(order)   

		for _, orderItem := range orderItemPack.Order_items {  
			orderItem.Order_id = order_id 

			validationErr := validate.Struct(orderItem) 

			if validationErr != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": validationErr.Error(),
				})
			}

			orderItem.ID = primitive.NewObjectID()
			orderItem.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
			orderItem.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
			orderItem.Order_item_id = orderItem.ID.Hex()
			var num = toFixed(*orderItem.Unit_price, 2) 

			orderItem.Unit_price = &num 

			orderItemsToBeInserted = append(orderItemsToBeInserted, orderItem) 

		}

		insertedItems, err := orderItemCollection.InsertMany(ctx, orderItemsToBeInserted)
		

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, insertedItems) 

	}
}

// @Summary         Updates an oderitem resource
// @Description     Updates an existing orderitem resource in the orderitem collection
// @Tags            orderitem
// @Accept          application/json
// @Produce         application/json
// @Param           orderitem_id path string true "ID of the orderitem  resource to update"
// @Param           orderitme body models.OrderItem true "Orderitem object"
// @Security        @Security.require(true)
// @Success         200 {object} models.OrderItem "orderitem  got updated with new body"
// @Failure         500 {string} http.StatusInternalServerError "Internal Server Error while updating food"
// @Failure         404 {string} http.StatusBadRequest "Bad Request"
// @Router          /orderitems/{orderitem_id} [patch]
func UpdateOrderItem() gin.HandlerFunc { 
	return func(c *gin.Context) {

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var orderItem models.OrderItem

		orderItemId := c.Param("order_item_id")

		if err := c.BindJSON(&orderItem); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		filter := bson.M{"order_item_id": orderItemId}

		var updateObj primitive.D

		if orderItem.Unit_price != nil {
			updateObj = append(updateObj, bson.E{Key: "unit_price", Value: *orderItem.Unit_price})
		}

		if orderItem.Quantity != nil {
			updateObj = append(updateObj, bson.E{Key: "quantity", Value: *orderItem.Quantity})
		}

		if orderItem.Food_id != nil {
			updateObj = append(updateObj, bson.E{Key: "food_id", Value: orderItem.Food_id})
		}

		orderItem.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		updateObj = append(updateObj, bson.E{Key: "updated_at", Value: orderItem.Updated_at})

		upsert := true

		opt := options.UpdateOptions{
			Upsert: &upsert,
		}

		result, err := orderItemCollection.UpdateOne(
			ctx,
			filter,
			bson.D{
				{Key: "$set", Value: updateObj},
			},
			&opt,
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}
