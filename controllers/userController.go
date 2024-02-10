package controllers

import (
	"context"
	"fmt"
	"go-restaurent-management-system/database"
	helper "go-restaurent-management-system/helpers"
	"go-restaurent-management-system/models"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")

func GetUsers() gin.HandlerFunc {
	return func(c *gin.Context) {

		recordPerPage, err := strconv.Atoi(c.Query("recordPerPage")) // this is related to the skip and limit concept , which is mainly used in pagination .
		// when we are getting bulk amount of data from database , we use pagination concept to send data to the frontend , in the pagination concept skip tells from where the data should be sent and limit restrcits the amount of data to be sent to the frontend.
		if err != nil || recordPerPage < 1 {

			recordPerPage = 10

		}

		page, err := strconv.Atoi(c.Query("page"))

		if err != nil || page < 1 {
			page = 1
		}

		startIndex := (page - 1) * recordPerPage
		startIndex, _ = strconv.Atoi(c.Query("startindex"))

		matchStage := bson.D{{Key: "$match ", Value: bson.D{{}}}}
		projectStage := bson.D{
			{Key: "$project", Value: bson.D{
				{Key: "id", Value: 0},
				{Key: "total_count", Value: 1},
				{Key: "user_items", Value: bson.D{{Key: "$slice", Value: []interface{}{"$data", startIndex, recordPerPage}}}}, //{"user_items", bson.D{{"$slice", []interface{}{"$data", startIndex, recordPerPage}}}}: This projection includes the user_items field in the output documents, and it uses the $slice operator to limit the array contained in the user_items field to a specific range. The $slice operator takes three arguments: the source array (in this case, $data), the starting index (specified by startIndex), and the number of elements to include (specified by recordPerPage). This is used for paginating the user_items array.
			}}}
		//The $slice operator in MongoDB is used to limit the number of elements in an array that is projected in a query or aggregation pipeline.
		//This specific line of code is defining a projection for a MongoDB query or aggregation operation, and it's using the $slice operator to limit the elements in an array field called user_items
		//Key: "$slice": This part specifies that you want to use the $slice operator to modify the user_items field. The $slice operator is used to retrieve a portion (or slice) of an array.
		//Value: []interface{}{"$data", startIndex, recordPerPage}: This part specifies the arguments for the $slice operator. It's an array of three elements:

		//"data": This is the source array from which you want to retrieve a slice. It appears to be a reference to another field named data. In MongoDB, $data suggests that it's a field reference and not a literal array.
		//startIndex: This variable represents the starting index of the slice within the data array. It indicates where the slice should begin.
		//startIndex: This variable represents the starting index of the slice within the data array. It indicates where the slice should begin.
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		result, err := userCollection.Aggregate(ctx, mongo.Pipeline{
			matchStage,
			projectStage,
		})

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while listing  all the user documents in the usercolleciton"})
			return
		}
		// gin.H is of type map[string]interface{} , which is map data structure of key value pairs
		// bson.M = document with unordered key value pairs
		// bson.D = document with ordered key value pairs.

		var allUsers []bson.M //collection.Find: This method initiates a query on a MongoDB collection. It returns a cursor pointing to the result set that matches the specified query.

		if err := result.All(ctx, &allUsers); err != nil {
			log.Fatal(err)
		}

		c.JSON(http.StatusOK, allUsers[0])

	}
}

func GetUser() gin.HandlerFunc {
	return func(c *gin.Context) {

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		var user models.User

		userId := c.Param("user_id") 

		err := userCollection.FindOne(ctx, bson.M{"user_id": userId}).Decode(&user)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, user)
	} 
}

func Login() gin.HandlerFunc {
	return func(c *gin.Context) {

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel() 

		var user models.User

		var foundUser models.User

		// convert the incoming request's json into golang readable format
		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// find a user with the emailid and check whether the user already exits

		err := userCollection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&foundUser)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while finding the document"})
			return
		}

		// verifying the password of the user

		passwordIsValid, msg := VerifyPassword(*user.Password, *foundUser.Password)

		if passwordIsValid != true {
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		// generate all tokens

		token, refreshToken, _ := helper.GenerateAllTokens(*foundUser.Email, *foundUser.First_name, *foundUser.Last_name, foundUser.User_id)
		fmt.Println("token is ", token)
		fmt.Println("refresh token", refreshToken)
		// update all the tokens

		helper.UpdateAllTokens(token, refreshToken, foundUser.User_id) 

		// return status ok

		fmt.Println("the logined user is ", foundUser)

		c.JSON(http.StatusOK, foundUser)

	}
}

func SignUp() gin.HandlerFunc {
	return func(c *gin.Context) {

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var user models.User

		//convert the incoming data from the http request into golang readable format

		if err := c.BindJSON(&user); err != nil { 
			c.JSON(http.StatusBadRequest, gin.H{"error": err})
			return
		}

		fmt.Println(user)

		// validate the data using the struct method

		validationErr := validate.Struct(user)

		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

		// checking whether the email is already existing in the database

		emailCount, err := userCollection.CountDocuments(ctx, bson.M{"email": user.Email})

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while checking for the email"})
			return
		}	

		// hash password
		password := HashPassword(*user.Password)
		user.Password = &password

		phoneCount, err := userCollection.CountDocuments(ctx, bson.M{"phone": user.Phone})

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error ": "error occurre while checking for the phone "})
			return
		}

		if emailCount > 0 && phoneCount > 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "This email or phone already exists in the database"})
			return
		}

		user.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

		user.ID = primitive.NewObjectID() // ObjectID, which is a unique identifier commonly used in MongoDB. Each ObjectID is typically 12 bytes,
		user.User_id = user.ID.Hex()     //The Hex() method converts the ObjectID to its hexadecimal string representation

		// generate token and refresh token

		token, refreshToken, _ := helper.GenerateAllTokens(*user.Email, *user.First_name, *user.Last_name, user.User_id)

		user.Token = &token
		user.RefreshToken = &refreshToken

		// inserting the new user into the database

		result, insertionErr := userCollection.InsertOne(ctx, user)

		if insertionErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while inserting user data into the database "})
			return
		}

		fmt.Println("The new user details are ", result)

		c.JSON(http.StatusOK, result)

	}
}

func HashPassword(password string) string {

	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14) //The 14 signifies the cost factor for the bcrypt password hashing algorithm. Bcrypt is designed to be slow and computationally expensive, which makes it resistant to brute-force and dictionary attacks. The cost factor determines how many iterations of the underlying Blowfish encryption algorithm are applied to hash the password. Higher cost factors result in more iterations and, therefore, slower hash generation.

	if err != nil { // n Go, a string can be converted to a byte slice using the []byte type conversion. The []byte conversion treats the string as a sequence of bytes, where each character in the string is represented by its ASCII value.
		log.Panic(err)

	}

	return string(bytes)

}

func VerifyPassword(userPassword string, providedPassword string) (bool, string) {

	err := bcrypt.CompareHashAndPassword([]byte(providedPassword), []byte(userPassword))
	check := true
	msg := ""

	if err != nil {
		msg = fmt.Sprintf("login or password is incorrect")
		check = false
	}
	return check, msg
}
