package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Person struct {
	ID       int    `json:"id" bson:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RequestBody struct {
	Person Person `json:"person"`
	Status string `json:"status"`
}

type Response struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func showPageRegister(w http.ResponseWriter, r *http.Request) {
	page, err := os.ReadFile("register.html")
	if err != nil {
		log.Fatal(err)
	}
	w.Header().Set("Content-Type", "text/html")
	w.Write(page)
	return
}

func showPageUsers(w http.ResponseWriter, r *http.Request) {
	page, err := os.ReadFile("users.html")
	if err != nil {
		log.Fatal(err)
	}
	w.Header().Set("Content-Type", "text/html")
	w.Write(page)
	return
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleGetRequest(w, r)
		case http.MethodPost:
			handlePostRequest(w, r)
		case http.MethodDelete:
			handleDeleteRequest(w, r)
		case http.MethodPut:
			handlePutRequest(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/main", showPageRegister)
	http.HandleFunc("/users", handleGetAllUsers)

	fmt.Println("Server is running on port 8080...")
	http.ListenAndServe(":8080", nil)
}

func handleGetRequest(w http.ResponseWriter, r *http.Request) {
	// Parse the URL query parameters
	params := r.URL.Query()
	id := params.Get("id")

	if id == "" {
		response := Response{
			Status:  "400",
			Message: "Missing 'id' parameter",
		}
		sendResponse(w, response, http.StatusBadRequest)
		return
	}

	// Convert the id parameter to integer
	personID, err := strconv.Atoi(id)
	if err != nil {
		response := Response{
			Status:  "400",
			Message: "Invalid 'id' parameter",
		}
		sendResponse(w, response, http.StatusBadRequest)
		return
	}

	// Connect to MongoDB
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context.Background())

	// Access the "users" collection in the "golangDatabase" database
	collection := client.Database("golangDatabase").Collection("users")

	// Find the user by ID in the collection
	var user Person
	err = collection.FindOne(context.Background(), bson.M{"id": personID}).Decode(&user)
	if err != nil {
		response := Response{
			Status:  "404",
			Message: "User not found",
		}
		sendResponse(w, response, http.StatusNotFound)
		return
	}

	// Respond with the user data in a message
	response := Response{
		Status:  "success",
		Message: fmt.Sprintf("User found with ID %s and Name %s", id, user.Name),
	}
	sendResponse(w, response, http.StatusOK)
}
func handlePostRequest(w http.ResponseWriter, r *http.Request) {
	var requestBody RequestBody
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&requestBody)
	if err != nil {
		response := Response{
			Status:  "400",
			Message: "Invalid JSON message",
		}
		sendResponse(w, response, http.StatusBadRequest)
		return
	}

	// Connect to MongoDB
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context.Background())

	// Access the "users" collection in the "golangDatabase" database
	collection := client.Database("golangDatabase").Collection("users")

	// Count the number of documents in the collection
	count, err := collection.CountDocuments(context.Background(), bson.D{})
	if err != nil {
		response := Response{
			Status:  "500",
			Message: "Error counting documents in the database",
		}
		sendResponse(w, response, http.StatusInternalServerError)
		return
	}

	// Create a User instance with the calculated ID
	newUser := Person{
		ID:       int(count) + 1, // Assign the next available ID
		Name:     requestBody.Person.Name,
		Email:    requestBody.Person.Email,
		Password: requestBody.Person.Password,
	}

	// Insert the new user into the MongoDB collection
	_, err = collection.InsertOne(context.Background(), newUser)
	if err != nil {
		response := Response{
			Status:  "500",
			Message: "Error inserting data into the database",
		}
		sendResponse(w, response, http.StatusInternalServerError)
		return
	}

	response := Response{
		Status:  "success",
		Message: "Successfully registered",
	}
	fmt.Println("Received data:", requestBody)
	sendResponse(w, response, http.StatusOK)
}

func handleGetAllUsers(w http.ResponseWriter, r *http.Request) {
	// Connect to MongoDB
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context.Background())

	// Access the "users" collection in the "golangDatabase" database
	collection := client.Database("golangDatabase").Collection("users")

	// Find all documents in the collection
	cursor, err := collection.Find(context.Background(), bson.D{})
	if err != nil {
		response := Response{
			Status:  "500",
			Message: "Error querying data from the database",
		}
		sendResponse(w, response, http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.Background())

	// Slice to store user objects
	var users []Person

	// Loop through the cursor and append user objects to the slice
	for cursor.Next(context.Background()) {
		var user Person
		if err := cursor.Decode(&user); err != nil {
			log.Println("Error decoding user:", err)
			continue
		}
		users = append(users, user)
	}

	// Encode the JSON response and send it directly
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(users)
	if err != nil {
		response := Response{
			Status:  "500",
			Message: "Error encoding JSON response",
		}
		sendResponse(w, response, http.StatusInternalServerError)
		return
	}
}

func sendResponse(w http.ResponseWriter, response Response, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

func handleDeleteRequest(w http.ResponseWriter, r *http.Request) {
	// Parse the URL query parameters
	params := r.URL.Query()
	id := params.Get("id")

	if id == "" {
		response := Response{
			Status:  "400",
			Message: "Missing 'id' parameter",
		}
		sendResponse(w, response, http.StatusBadRequest)
		return
	}

	// Convert the id parameter to integer
	personID, err := strconv.Atoi(id)
	if err != nil {
		response := Response{
			Status:  "400",
			Message: "Invalid 'id' parameter",
		}
		sendResponse(w, response, http.StatusBadRequest)
		return
	}

	// Connect to MongoDB
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context.Background())

	// Access the "users" collection in the "golangDatabase" database
	collection := client.Database("golangDatabase").Collection("users")

	// Delete the document with the specified ID
	result, err := collection.DeleteOne(context.Background(), bson.M{"id": personID})
	if err != nil {
		response := Response{
			Status:  "500",
			Message: "Error deleting data from the database",
		}
		sendResponse(w, response, http.StatusInternalServerError)
		return
	}

	if result.DeletedCount == 0 {
		response := Response{
			Status:  "404",
			Message: "Document not found",
		}
		sendResponse(w, response, http.StatusNotFound)
		return
	}

	response := Response{
		Status:  "success",
		Message: fmt.Sprintf("Successfully deleted document with ID %s", id),
	}
	sendResponse(w, response, http.StatusOK)
}

func handlePutRequest(w http.ResponseWriter, r *http.Request) {
	// Parse the URL query parameters
	params := r.URL.Query()
	id := params.Get("id")
	newName := params.Get("newName")

	if id == "" || newName == "" {
		response := Response{
			Status:  "400",
			Message: "Missing 'id' or 'newName' parameter",
		}
		sendResponse(w, response, http.StatusBadRequest)
		return
	}

	// Convert the id parameter to integer
	personID, err := strconv.Atoi(id)
	if err != nil {
		response := Response{
			Status:  "400",
			Message: "Invalid 'id' parameter",
		}
		sendResponse(w, response, http.StatusBadRequest)
		return
	}

	// Connect to MongoDB
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context.Background())

	// Access the "users" collection in the "golangDatabase" database
	collection := client.Database("golangDatabase").Collection("users")

	// Update the name of the user with the specified ID
	result, err := collection.UpdateOne(
		context.Background(),
		bson.M{"id": personID},
		bson.D{{Key: "$set", Value: bson.D{{Key: "name", Value: newName}}}},
	)
	if err != nil {
		response := Response{
			Status:  "500",
			Message: "Error updating data in the database",
		}
		sendResponse(w, response, http.StatusInternalServerError)
		return
	}

	if result.ModifiedCount == 0 {
		response := Response{
			Status:  "404",
			Message: "User not found",
		}
		sendResponse(w, response, http.StatusNotFound)
		return
	}

	response := Response{
		Status:  "success",
		Message: fmt.Sprintf("Successfully updated name for user with ID %s", id),
	}
	sendResponse(w, response, http.StatusOK)
}
