package helpers

import (
	"context"
	"fmt"
	"go-restaurent-management-system/database"
	"log"
	"os"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
) 

type SignedDetails struct {
	Email      string
	First_name string
	Last_name  string
	Uid        string
	jwt.StandardClaims
}

//JWTs have three main parts: Header, Payload, and Signature.

var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user") 

var SECRET_KEY string = os.Getenv("SECRET_KEY")

func GenerateAllTokens(email string, firstName string, lastName string, uid string) (signedToken string, signedRefreshToken string, err error) {
	claims := &SignedDetails{  // claims are part of payload , which contains information related to user and roles to which he is assigned.
		Email:      email,
		First_name: firstName,
		Last_name:  lastName,
		Uid:        uid,
		StandardClaims: jwt.StandardClaims{ 
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(24)).Unix(),
		},
	} 

	refreshClaims := &SignedDetails{ 
		StandardClaims: jwt.StandardClaims{ //StandardClaims is a struct provided by jwt-go that includes standard JWT claims like ExpiresAt.
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(168)).Unix(), //Unix timestamp (seconds since January 1, 1970, UTC).
		},
	} 

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(SECRET_KEY))
	if err != nil {
		log.Panic(err)
	}
	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(SECRET_KEY))
	if err != nil {
		log.Panic(err)
	}
	//jwt.SigningMethodHS256: This indicates that the JWT will be signed using the HMAC-SHA256 algorithm. HMAC (Hash-based Message Authentication Code) is a method for creating a cryptographic hash of the token's content and signing it with a secret key to ensure its integrity and authenticity.
	//Signing: When a JWT is created and signed with this method, a secret key (or shared secret) is used to create a digital signature of the token's content. This signature is added to the JWT, ensuring that the token has not been tampered with.
	//Verification: When someone receives a JWT and wants to verify its authenticity, they must also have access to the same secret key. They can use this key to verify the signature and confirm that the token haodis not been mfied since it was signed.
	//SigningMethodHS256 : is a term related to JWT (JSON Web Tokens) and is typically used in Go libraries like "github.com/dgrijalva/jwt-go" for signing and verifying JWTs using the HMAC-SHA256 algorithm. HMAC-SHA256 is a widely used cryptographic algorithm for creating and verifying digital signatures.
	//SignedString([]byte(SECRET_KEY)): This is a method call that signs a JWT with the secret key. Here's how it works:

	//The SignedString method takes the secret key as a byte slice.

	//It uses a specific signing algorithm (e.g., HMAC-SHA256) and the provided secret key to create a digital signature of the JWT's content.

	//The digital signature is added to the JWT, ensuring that the token's content has not been tampered with.

	//The resulting JWT, which now includes the digital signature, can be shared with other parties.

	return token, refreshToken, err
} 

func UpdateAllTokens(signedToken string, signedRefreshToken string, userId string) {

	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	var updateObj primitive.D //In Go, the primitive.D type is used to represent a BSON document (Binary JSON) in MongoDB. The D stands for "Document," and it is a slice of primitive.E values, where each E represents a BSON element (field and value pair).

	updateObj = append(updateObj, bson.E{Key: "token", Value: signedToken})
	updateObj = append(updateObj, bson.E{Key: "refreshToken", Value: signedRefreshToken})

	Updated_at, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339)) 

	updateObj = append(updateObj, bson.E{Key: "updated_at", Value: Updated_at}) 

	upsert := true

	filter := bson.M{"user_id": userId}

	opts := options.UpdateOptions{ //options.UpdateOptions is a struct that holds various options for the update operation.
		Upsert: &upsert, // Upsert is an option in the UpdateOptions struct that specifies whether to perform an upsert operation. An upsert operation updates an existing document or inserts a new document if no matching document is found.
	} 

	_, err := userCollection.UpdateOne(
		ctx,
		filter, // filtering the document to be updated
		bson.D{
			{Key: "$set", Value: updateObj}, // setting / updating the data in the documents.
		},
		&opts, // update options
	)

	if err != nil {
		log.Panic(err)
		return
	}

	return
}

func ValidateToken(signedToken string) (claims *SignedDetails, msg string) {  

	token, err := jwt.ParseWithClaims(
		signedToken,      // jwt token to be parsed.
		&SignedDetails{}, // This is a pointer to an instance of the SignedDetails struct, which is used to store the claims extracted from the JWT. The claims will be populated after a successful parsing.

		func(token *jwt.Token) (interface{}, error) {    //n Go, returning an interface{} type from a function is a way to return a value of unspecified type. It's a very flexible feature but should be used with caution because it can make your code less type-safe and harder to understand. It's often used in certain situations, such as when working with data of different types dynamically or when you want to create generic functions. Here are some common scenarios where you might return an interface{} type from a function:
			return []byte(SECRET_KEY), nil                  // secret key is used by the server to access the claims of the json token .
		},
	) 

	//The purpose of this key function is to provide the secret key that is necessary to verify the JWT signature.
	//The secret key is used by the jwt library to check whether the JWT has been tampered with or generated by an unauthorized party.

	// if the token is invalid

	claims, ok := token.Claims.(*SignedDetails) // Type Assertion : In Go, a type assertion is a mechanism that allows you to convert an interface value to a concrete type. Type assertions are used when you have an interface value that could potentially hold a value of a specific type, and you want to access the value as that specific type. If the conversion is not possible, a type assertion may return an error or panic.
	// the above line checking that whether claims is of type signeddetailes

	if !ok {
		msg = fmt.Sprintf("the token is invalid")
		msg = err.Error()
		return
	}

	// the token is expired

	if claims.ExpiresAt < time.Now().Local().Unix() { // if token expiration time is less than now , then it is expired

		msg = fmt.Sprintf("token is expired")
		msg = err.Error()
		return

	}

	return claims, msg

} 
