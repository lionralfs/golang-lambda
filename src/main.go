package main

import (
	"context"
	"fmt"
	"time"

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

func handler(ctx context.Context, event events.SQSEvent) (events.SQSEventResponse, error) {
	for _, record := range event.Records {
		fmt.Println(record.Body)
		fmt.Println(record.MessageId)
	}
	response := events.SQSEventResponse{
		BatchItemFailures: []events.SQSBatchItemFailure{},
	}
	return response, nil
}

func main() {
	// lambda.Start(HandleRequest)
	lambda.Start(handler)
}
