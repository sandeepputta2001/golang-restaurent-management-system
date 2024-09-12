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

// @Summary       Returns slice of users
// @Description   Returns an array of users placed from the users collection  in restaurent database.
// @Tags          user
// @Produce       application/json
// @Success       200 {array} models.User "slice of tables "
// @Router        /users [get]
func GetUsers() gin.HandlerFunc {
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
					{Key: "users", Value: bson.D{{Key: "$slice", Value: []interface{}{"$data", startIndex, recordPerPage}}}},
					{Key: "data" , Value: 1},
				}}}

		result, err := userCollection.Aggregate(ctx, mongo.Pipeline{
			matchStage, groupStage, projectStage,
		})

		
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while listing users"})
			return
		}

		var allUsers []bson.M

		if err := result.All(ctx, &allUsers); err != nil {
			log.Fatal(err)
		}

		c.JSON(http.StatusOK, allUsers)
	}
}

// @Summary       Retrieves a user with specific user id
// @Description   Retrieves a user with specific user id from the users collection
// @Tags          user
// @Accept        application/json
// @Produce       application/json
// @Param         user_id path string true "user_id"
// @Success       200 {object} models.User "Details of user with specific user_id"
// @Failure       500 {string} http.StatusInternalServerError "Internal Server Error in mongodb"
// @Router        /users/{user_id} [get]
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

// @Summary       Creates a new login for a user 
// @Description   User gets logged in 
// @Tags          user
// @Accept        application/json
// @Produce       application/json
// @Param         user body models.User true  "User object"
// @Success       200 {object} models.User "User logged in "
// @Failure       500 {string} http.StatusInternalServerError "Internal Server Error while logging  a new user"
// @Router        /user/login [post]
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

// @Summary       Creates a new resource
// @Description   Creates a new users resource on the server
// @Tags          user
// @Accept        application/json
// @Produce       application/json
// @Param         user body models.User true  "User object"
// @Success       201 {object} models.User "New user created"
// @Failure       500 {string} http.StatusInternalServerError "Internal Server Error while creating a new user"
// @Router        /users/signup [post]
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

		

		phoneCount, err := userCollection.CountDocuments(ctx, bson.M{"phone": user.Phone})

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error ": "error occurre while checking for the phone "})
			return
		}

		if emailCount > 0 && phoneCount > 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "This email or phone already exists in the database"})
			return
		}

		// hash password
		password := HashPassword(*user.Password)
		user.Password = &password

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
