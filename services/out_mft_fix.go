package services

import (
	"GoFiber_Project01/DBConnection"
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/bson"
)

func Fixes() error {
	var (
		collectionName = "booking_consignment"
		docIndex       int
		totalDocs      int
	)
	db := DBConnection.DB
	MongoClient := DBConnection.MongoClient
	defer MongoClient.Disconnect(context.Background())
	successCount := 0
	query := bson.M{
		"mdes":   bson.M{"$exists": false},
		"txn_id": bson.M{"$exists": true},
	}
	cursor, err := db.Collection(collectionName).Find(context.TODO(), query)
	if err != nil {
		log.Fatal(err)
	}
	defer cursor.Close(context.TODO())

	var data []bson.M
	if err = cursor.All(context.TODO(), &data); err != nil {
		log.Fatal(err)
	}

	totalDocs = len(data)
	fmt.Printf("Found %d docs\n", totalDocs)

	for _, doc := range data {
		txnID := doc["txn_id"]
		var mft bson.M
		err := db.Collection("out_mft").FindOne(context.TODO(), bson.M{"_id": txnID}).Decode(&mft)
		if err != nil {
			log.Println(err)
			continue
		}

		updateResult, err := db.Collection(collectionName).UpdateOne(
			context.TODO(),
			bson.M{"_id": doc["_id"]},
			bson.M{"$set": bson.M{"mdes": mft["to_des"]}},
		)
		if err != nil || updateResult.ModifiedCount != 1 {
			log.Printf("Failed to update document: %v", doc["cno"])
		} else {
			successCount++
			log.Printf("%d/%d - %v updated to the local DB", docIndex, totalDocs, doc["cno"])
		}
	}

	return nil
}
