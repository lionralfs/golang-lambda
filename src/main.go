package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func makeWriteModel(now primitive.DateTime, movieId string, movieName string) mongo.WriteModel {
	return mongo.NewUpdateOneModel().SetUpsert(true).SetFilter(bson.D{
		{"mId", movieId},                     // modify a movie by the movieId
		{"lastModified", bson.M{"$lt": now}}, // ensure that the lastModified timestamp is < now
	}).SetUpdate(bson.D{
		{"$set", bson.D{
			{"mName", movieName},   // update the movie name
			{"lastModified", now}}, // update the lastModified
		}},
	)
}

func handleRequest(ctx context.Context) (string, error) {
	uri := "mongodb://mongouser:mongopass@localhost:27017"
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := client.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()

	coll := client.Database("my-database").Collection("movies")

	now := primitive.NewDateTimeFromTime(time.Now())

	// an array of write models
	models := []mongo.WriteModel{
		makeWriteModel(now, "1", "Scarface"),
		makeWriteModel(now, "2", "Back to the future"),
	}
	// use an unordered bulk write, this allows mongo to continue writing items, even if some failed
	// see https://www.mongodb.com/docs/manual/reference/method/db.collection.bulkWrite/#unordered-bulk-write-example
	opts := options.BulkWrite().SetOrdered(false)
	results, err := coll.BulkWrite(context.TODO(), models, opts)

	if err != nil {
		fmt.Printf("Errors during bulk write: %s\n", err)
	}
	fmt.Printf("Inserted %d\n", results.InsertedCount)
	fmt.Printf("Upserted %d\n", results.UpsertedCount)
	fmt.Printf("Modified %d\n", results.ModifiedCount)

	// example output (movie 1 has a timestamp in the future, movie 2 didn't exist):

	// Errors during bulk write: bulk write exception: write errors: [E11000 duplicate key error collection: my-database.movies index: mId_1 dup key: { mId: "1" }]
	// Inserted 0
	// Upserted 1
	// Modified 0

	return "Hello!", nil
}

type MessageBody struct {
	Id string `json:"id"`
}

func handler(ctx context.Context, event events.SQSEvent) (events.SQSEventResponse, error) {
	failures := []events.SQSBatchItemFailure{}
	for _, record := range event.Records {
		if failure := handleMessage(record); failure != nil {
			failures = append(failures, *failure)
		}
	}
	response := events.SQSEventResponse{
		BatchItemFailures: failures,
	}
	return response, nil
}

func handleMessage(message events.SQSMessage) *events.SQSBatchItemFailure {
	var body MessageBody

	// turn the message body into a byte array and parse it
	err := json.Unmarshal([]byte(message.Body), &body)
	if err != nil {
		fmt.Println("Parsing message failed", err)
		return &events.SQSBatchItemFailure{ItemIdentifier: message.MessageId}
	}
	if body.Id == "" {
		fmt.Println("Id field missing or empty")
		return &events.SQSBatchItemFailure{ItemIdentifier: message.MessageId}
	}

	resp, err := http.Get(fmt.Sprintf("https://jsonplaceholder.typicode.com/todos/%s", body.Id))
	if err != nil {
		fmt.Println("Request failed", err)
		return &events.SQSBatchItemFailure{ItemIdentifier: message.MessageId}
	}
	fmt.Printf("METRIC response_code 1 %d\n", resp.StatusCode)
	if resp.StatusCode != 200 {
		fmt.Println("Received non 200 status code", resp.StatusCode)
		return &events.SQSBatchItemFailure{ItemIdentifier: message.MessageId}
	}

	return nil
}

func main() {
	lambda.Start(handler)
}
