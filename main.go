package main

import (
	"context"
	"encoding/json"
	"fmt"
  "log"
	"net/http"
	"time"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
  "go.mongodb.org/mongo-driver/mongo/options"
  "golang.org/x/crypto/bcrypt"
)

var client *mongo.Client

type Post struct {
	ID  primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Caption string `json:"caption,omitempty" bson:"caption,omitempty"`
	ImageUrl string `json:"imageurl,omitempty" bson:"imageurl,omitempty"`
  PostedTimestamp time.Time `json:"timestamp,omitempty" bson:"timestamp,omitempty"`
}

type User struct {
	ID  primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Name string `json:"name,omitempty" bson:"name,omitempty"`
	Email string `json:"email,omitempty" bson:"email,omitempty"`
  Password string `json:"password,omitempty" bson:"password,omitempty"`
}


func getHash(pwd []byte) string {
  //Generate a salt and apply hash on the password
    hash, err := bcrypt.GenerateFromPassword(pwd, bcrypt.MinCost)

    //Handling errors
    if err != nil {
        log.Println(err)
    }

    //Convert the hash (in byte array) to string
    return string(hash)
}



func getUser(response http.ResponseWriter, request *http.Request) {
  //Set the contentType in response header as application/json
	response.Header().Set("content-type", "application/json")

  //Get the request params using mux
	params := mux.Vars(request)

  //Creates a objectId from hex string id passed as input
	id, _ := primitive.ObjectIDFromHex(params["id"])

  //Create user
	var user User

  //Get the required collection from the database to work on
  collection := client.Database("Instagram").Collection("User")

  /*
   * Prepare the context
   * Setting the timeout as 10 sec for communicating with DB
  */
  ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

  //Get the document (userInfo) using the ID passed and decode it in user var
	err := collection.FindOne(ctx, User{ID: id}).Decode(&user)

  //Handler errors
	if err != nil {
    //Set the internal server error in response header
		response.WriteHeader(http.StatusInternalServerError)

    //Get the error message, Convert to byte array, Set it in response
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}

  //Use the JSON encoder to encode the result
	json.NewEncoder(response).Encode(user)
}

func getPost(response http.ResponseWriter, request *http.Request) {
  //Set the contentType in response header as application/json
	response.Header().Set("content-type", "application/json")

  //Get the request params using mux
	params := mux.Vars(request)

  //Creates a objectId from hex string id passed as input
	id, _ := primitive.ObjectIDFromHex(params["id"])

  //Create post
	var post Post

  //Get the required collection from the database to work on
  collection := client.Database("Instagram").Collection("Post")

  /*
   * Prepare the context
   * Setting the timeout as 10 sec for communicating with DB
  */
  ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

  //Get the document (postInfo) using the ID passed and decode it in post var
	err := collection.FindOne(ctx, Post{ID: id}).Decode(&post)

  //Handler errors
	if err != nil {
    //Set the internal server error in response header
		response.WriteHeader(http.StatusInternalServerError)

    //Get the error message, Convert to byte array, Set it in response
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}

  //Use the JSON encoder to encode the result
	json.NewEncoder(response).Encode(post)
}


func createUser(response http.ResponseWriter, request *http.Request) {
  //Set the contentType in response header as application/json
	response.Header().Set("content-type", "application/json")

  //Create user
	var user User

  //Use the JSON Decoder to decode the request body and store in user var
	json.NewDecoder(request.Body).Decode(&user)

  //Get the user password, Convert to byte array, Apply hash func and restore it back
  user.Password = getHash([]byte(user.Password))

  //Get the required collection from the database to work on
	collection := client.Database("Instagram").Collection("User")

  /*
   * Prepare the context
   * Setting the timeout as 10 sec for communicating with DB
  */
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

  //Insert the document (i.e. user var) in the prepared collection
	result, _ := collection.InsertOne(ctx, user)

  //Use the JSON encoder to encode the result
	json.NewEncoder(response).Encode(result)
}

func createPost(response http.ResponseWriter, request *http.Request) {
  //Set the contentType in response header as application/json
	response.Header().Set("content-type", "application/json")

  //Create post
	var post Post

  //Use the JSON Decoder to decode the request body and store in post var
	json.NewDecoder(request.Body).Decode(&post)

  //Get current time and update in posted timestamp
  post.PostedTimestamp = time.Now()

  //Get the required collection from the database to work on
	collection := client.Database("Instagram").Collection("Post")

  /*
   * Prepare the context
   * Setting the timeout as 10 sec for communicating with DB
  */
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

  //Insert the document (i.e. post var) in the prepared collection
	result, _ := collection.InsertOne(ctx, post)

  //Use the JSON encoder to encode the result
	json.NewEncoder(response).Encode(result)
}

func getAllPostsOfUser(response http.ResponseWriter, request *http.Request) {
  //Set the contentType in response header as application/json
	response.Header().Set("content-type", "application/json")

  //Create list of posts
	var posts []Post

  //Get the required collection from the database to work on
	collection := client.Database("Instagram").Collection("Post")

  /*
   * Prepare the context
   * Setting the timeout as 10 sec for communicating with DB
  */
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

  //Get all posts
	cursor, err := collection.Find(ctx, bson.M{})

  //Handler errors
	if err != nil {
    //Set the internal server error in response header
		response.WriteHeader(http.StatusInternalServerError)

    //Get the error message, Convert to byte array, Set it in response
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}

	//defer cursor.Close(ctx)
  //Iterate each document
	for cursor.Next(ctx) {
    //Create a post doc
		var post Post

    //Decode in post var
		cursor.Decode(&post)

    //Append the post in list
		posts = append(posts, post)
	}

  //Handler errors
	if err := cursor.Err(); err != nil {
    //Set the internal server error in response header
		response.WriteHeader(http.StatusInternalServerError)

    //Get the error message, Convert to byte array, Set it in response
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}

  //Use the JSON encoder to encode the posts
	json.NewEncoder(response).Encode(posts)
}


func main() {
	fmt.Println("Starting the application...")
  /*
   * Prepare the context
   * Setting the timeout as 10 sec for communicating with DB
  */
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

  /*
   * Prepare the clientOptions
   * Preparing the client with localhost URL:mongoDefaultPort
  */
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")

  /*
   * Build the client by connecting with mongo using context & clientOptions prepared above
  */
	client, _ = mongo.Connect(ctx, clientOptions)

  //Get the muxRouter which will used to handle and dispatch the incoming request
	router := mux.NewRouter()

  //Map the endpoint path with appropriate handleFunction using HTTP verbs
  //Create a user
	router.HandleFunc("/users", createUser).Methods("POST")

  //Get a user using the Id
	router.HandleFunc("/users/{id}", getUser).Methods("GET")

  //Create a post
	router.HandleFunc("/posts", createPost).Methods("POST")

  //Get a post using the Id
  router.HandleFunc("/posts/{id}", getPost).Methods("GET")

  //Get all posts of all users
  router.HandleFunc("/posts/users", getAllPostsOfUser).Methods("GET")

  //Listen to the port 12345 and pass on the request to the router
	http.ListenAndServe(":12345", router)
}
