package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	IpConnectMongo   = "127.0.0.1"
	PortConnectMongo = "27017"
	Login            = "admin"
	Password         = "password"
)

func createMongoDBClient() (*mongo.Client, error) {
	clientOptions := options.Client().
		ApplyURI("mongodb://" + Login + ":" + Password + "@" +
			IpConnectMongo + ":" + PortConnectMongo)
	fmt.Print("mongodb://" + Login + ":" + Password + "@" +
		IpConnectMongo + ":" + PortConnectMongo)
	// "mongodb://admin:password@mongo6.6:27017"

	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		return nil, err
	}

	err = client.Ping(context.Background(), nil)
	if err != nil {
		return nil, err
	}

	return client, nil
}

type Request struct {
	Method     string              `bson:"method"`
	Path       string              `bson:"path"`
	GetParams  map[string][]string `bson:"get_params"`
	Headers    http.Header         `bson:"headers"`
	Cookies    []http.Cookie       `bson:"cookies"`
	PostParams map[string][]string `bson:"post_params"`
}

type Response struct {
	Code    int         `bson:"code"`
	Message string      `bson:"message"`
	Headers http.Header `bson:"headers"`
	Body    string      `bson:"body"`
}

type RequestRepository struct {
	ID         primitive.ObjectID  `bson:"_id,omitempty"`
	Method     string              `bson:"method"`
	Path       string              `bson:"path"`
	GetParams  map[string][]string `bson:"get_pa	rams"`
	Headers    http.Header         `bson:"headers"`
	Cookies    []http.Cookie       `bson:"cookies"`
	PostParams map[string][]string `bson:"post_params"`
	Timestamp  time.Time           `bson:"timestamp"`
}

type ResponseRepository struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Code      int                `bson:"code"`
	Message   string             `bson:"message"`
	Headers   http.Header        `bson:"headers"`
	Body      string             `bson:"body"`
	IdRequest primitive.ObjectID `bson:"request_id"`
	Timestamp time.Time          `bson:"timestamp"`
}

type BD interface {
	SaveResponseRequest(Response, Request) error
	GetRequestByID(string) (Request, error)
	GetAllRequests() ([]Request, error)
}

type MongoDB struct {
}

func (m MongoDB) SaveResponseRequest(resp Response, req Request) error {
	client, err := createMongoDBClient()
	if err != nil {
		return err
	}
	defer client.Disconnect(context.Background())

	db := client.Database("http_logs")
	requestsCollection := db.Collection("requests")
	responsesCollection := db.Collection("responses")

	timeNow := time.Now()

	requestMongo := RequestRepository{
		Method:     req.Method,
		Path:       req.Path,
		GetParams:  req.GetParams,
		Headers:    req.Headers,
		Cookies:    req.Cookies,
		PostParams: req.PostParams,
		Timestamp:  timeNow,
	}

	responseMongo := ResponseRepository{
		Code:      resp.Code,
		Message:   resp.Message,
		Headers:   resp.Headers,
		Body:      resp.Body,
		Timestamp: timeNow,
	}

	idOfReq, err := requestsCollection.InsertOne(context.Background(), requestMongo)
	if err != nil {
		return err
	}

	responseMongo.IdRequest = idOfReq.InsertedID.(primitive.ObjectID)
	_, err = responsesCollection.InsertOne(context.Background(), responseMongo)
	if err != nil {
		return err
	}

	return nil
}

func (m MongoDB) GetRequestByID(id string) (Request, error) {
	client, err := createMongoDBClient()
	if err != nil {
		return Request{}, err
	}
	defer client.Disconnect(context.Background())

	db := client.Database("http_logs")
	requestsCollection := db.Collection("requests")

	var retrievedRequest RequestRepository
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return Request{}, err
	}

	err = requestsCollection.FindOne(context.Background(), bson.M{"_id": objectID}).Decode(&retrievedRequest)
	if err != nil {
		return Request{}, err
	}

	return Request{
		Method:     retrievedRequest.Method,
		Path:       retrievedRequest.Path,
		GetParams:  retrievedRequest.GetParams,
		Headers:    retrievedRequest.Headers,
		Cookies:    retrievedRequest.Cookies,
		PostParams: retrievedRequest.PostParams,
	}, nil
}

func (m MongoDB) GetAllRequests() ([]Request, error) {
	client, err := createMongoDBClient()
	if err != nil {
		return nil, err
	}
	defer client.Disconnect(context.Background())

	db := client.Database("http_logs")
	collection := db.Collection("requests")

	cursor, err := collection.Find(context.Background(), bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var requests []Request
	for cursor.Next(context.Background()) {
		var request RequestRepository
		if err := cursor.Decode(&request); err != nil {
			return nil, err
		}
		requests = append(requests, Request{
			Method:     request.Method,
			Path:       request.Path,
			GetParams:  request.GetParams,
			Headers:    request.Headers,
			Cookies:    request.Cookies,
			PostParams: request.PostParams,
		})
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return requests, nil
}