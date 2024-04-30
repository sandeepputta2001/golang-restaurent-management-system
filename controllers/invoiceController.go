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

type InvoiceViewFormat struct {
	Invoice_id       string
	Payment_method   string
	Order_id         string
	Payment_status   *string 
	Payment_due      interface{}
	Table_number     interface{}
	Payment_due_date time.Time
	Order_details    interface{}
}

var invoiceCollection *mongo.Collection = database.OpenCollection(database.Client, "invoices")

// @Summary         Returns slice of invoices 
// @Description     Returns an array of invoices from the invoiceCollection in restaurent database.
// @Tags            invoice
// @Security        @Security.require(true)
// @Produce         application/json
// @Success         200 {array} models.Invoice "slice of invoices "
// @Router          /invoices [get]
func GetInvoices() gin.HandlerFunc {
	return func(c *gin.Context) {

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
        defer cancel()
		result, err := invoiceCollection.Find(ctx, bson.M{})

		

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while finding the invoice documents"})
			return
		}

		var allInvoices []bson.M

		if err := result.All(ctx, &allInvoices); err != nil {
			log.Fatal(err)
		}

		c.JSON(http.StatusOK, allInvoices)

	}
}

// @Summary         Retrieves a invoice  with specific invoice id
// @Description     Retrieves a invoice  with specific inovoice  id from the  invoices collection
// @Tags            invoice
// @Accept          application/json
// @Produce         application/json
// @Param           invoice_id path string true "invoice_id"
// @Security        @Security.require(true)
// @Success         200 {object}  InvoiceViewFormat "Details of a specific invoice created"
// @Failure         500 {string} http.StatusInternalServerError "Internal Server Error in mongodb"
// @Router          /invoices/{invoice_id} [get]
func GetInvoice() gin.HandlerFunc {
	return func(c *gin.Context) {

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
        defer cancel()
		var invoice models.Invoice 

		invoiceId := c.Param("invoice_id")

		err := invoiceCollection.FindOne(ctx, bson.M{"invoice_id": invoiceId}).Decode(&invoice) 
		
		if err != nil { 
			c.JSON(http.StatusInternalServerError, gin.H{"error ": err.Error()})
			return
		}

		var invoiceView InvoiceViewFormat

		allOrderItems, err := ItemsByOrder(invoice.Order_id)

		invoiceView.Order_id = invoice.Order_id 
		invoiceView.Payment_due_date = invoice.Payment_due_date 

		invoiceView.Payment_method = "null"

		if invoice.Payment_method != nil { 
			invoiceView.Payment_method = *invoice.Payment_method
		}

		invoiceView.Invoice_id = invoice.Invoice_id
		invoiceView.Payment_status = invoice.Payment_status
		invoiceView.Payment_due = allOrderItems[0]["payment_due"]
		invoiceView.Table_number = allOrderItems[0]["table_number"]
		invoiceView.Order_details = allOrderItems[0]["order_items"]
		//allOrderItems[0]["payment_due"] accesses the value associated with the key "payment_due" in the first map within allOrderItems.
		//This line assigns the value of "payment_due" from the first order item map to the Payment_due field in the invoiceView struct.

		c.JSON(http.StatusOK, invoiceView)

	}
}

// @Summary         Creates a invoice 
// @Description     Creates a invoice resource on the server
// @Tags            invoice
// @Accept          application/json
// @Produce         application/json
// @Param           Invoice body models.Invoice true "Invoice object"
// @Security        @Security.require(true)
// @Success         201 {object} models.Invoice "New invoice created"
// @Failure         500 {string} http.InternalServerError "Internal Server Error while creating a new food"
// @Router          /invoices  [post]
func CreateInvoice() gin.HandlerFunc {
	return func(c *gin.Context) {

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var invoice models.Invoice
		var order models.Order 

		if err := c.BindJSON(&invoice); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error ": err.Error()}) 
			return
		}

		err := orderCollection.FindOne(ctx, bson.M{"order_id": invoice.Order_id}).Decode(&order)
		if err != nil {
			msg := fmt.Sprintf("Order is not in the order collection ")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		status := "PENDING"

		if invoice.Payment_status == nil { 
			invoice.Payment_status = &status
		}

		invoice.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		invoice.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		invoice.Payment_due_date, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		invoice.ID = primitive.NewObjectID()
		invoice.Invoice_id = invoice.ID.Hex()

		result, err := invoiceCollection.InsertOne(ctx, invoice)
		

		if err != nil {
			msg := fmt.Sprintf("invoice was not created ")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		c.JSON(http.StatusOK, result)

	}
}

// @Summary         Updates an invoice resource
// @Description     Updates an existing invoice  resource in the invoices collection
// @Tags            invoice
// @Accept          application/json
// @Produce         application/json 
// @Param           invoice_id path string true "ID of the invoice resource to update"
// @Param           Invoice body models.Invoice true "Invoice object"
// @Security        @Security.require(true)
// @Success         200 {object} models.Invoice "invoice got updated with new body"
// @Failure         500 {string} http.InternalServerError "Internal Server Error while updating invoice"
// @Failure         404 {string} http.StatusBadRequest "Bad Request"
// @Router          /invoices/{invoice_id} [patch] 
func UpdateInvoice() gin.HandlerFunc {
	return func(c *gin.Context) {

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		var invoice models.Invoice

		if err := c.BindJSON(&invoice); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

			defer cancel()
			return
		}

		invoiceId := c.Param("invoice_id")

		var updateObj primitive.D

		filter := bson.M{"invoice_id": invoiceId}

		if invoice.Payment_method != nil {
			updateObj = append(updateObj, bson.E{Key: "payment_method", Value: invoice.Payment_method})

		}

		if invoice.Payment_status != nil {
			updateObj = append(updateObj, bson.E{Key: "payment_status", Value: invoice.Payment_status})

		}

		invoice.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

		updateObj = append(updateObj, bson.E{Key: "updated_at", Value: invoice.Updated_at})

		upsert := true

		opt := options.UpdateOptions{
			Upsert: &upsert,
		}

		status := "PENDING"

		if invoice.Payment_status == nil {
			invoice.Payment_status = &status
		}

		result, err := invoiceCollection.UpdateOne(
			ctx,
			filter,
			bson.D{
				{Key: "$set", Value: updateObj},
			},
			&opt,
		)

		defer cancel()

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while updating the document"})
			return
		}

		c.JSON(http.StatusOK, result)

	}
}
