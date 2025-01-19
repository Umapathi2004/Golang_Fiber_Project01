package services

import (
	"Blr_server_update/DBConnection"
	"Blr_server_update/logs"
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/bson"
)

func Fixes() error {
	var (
		collectionName = "booking_consignment"
		docIndex       int
		totalDocs int
	)
	_, ErrorLog := logs.Logger()
	db, MongoClient, err := DBConnection.InitMongoDB()
	if err != nil {
		log.Printf("Error Connected to MongoDB: %v\n", err)
		return nil
	}
	// db := DBConnection.DB
	// MongoClient := DBConnection.MongoClient
	defer MongoClient.Disconnect(context.Background())
	// successCount := 0
	query := bson.M{
		"mdes":   bson.M{"$exists": false},
		"txn_id": bson.M{"$exists": true},
	}
	cursor, err := db.Collection(collectionName).Find(context.TODO(), query)
	if err != nil {
		ErrorLog.Printf("failed to query documents: %v\n", err)
		return nil
	}
	defer cursor.Close(context.TODO())

	var data []bson.M
	if err = cursor.All(context.TODO(), &data); err != nil {
		ErrorLog.Printf("failed to parse documents: %v\n", err)
		return nil
	}

	totalDocs = len(data)
	fmt.Printf("Found %d docs\n", totalDocs)
	if totalDocs == 0 {
		ErrorLog.Printf("Found %d docs\n", totalDocs)
		return nil
	}
	for _, doc := range data {
		docIndex++
		txnID := doc["txn_id"]
		var mft bson.M
		err := db.Collection("out_mft").FindOne(context.TODO(), bson.M{"_id": txnID}).Decode(&mft)
		if err != nil {
			ErrorLog.Printf("failed to find mft for txn_id: %v\n", txnID)
			continue
		}
		// if mft != nil {
		// 	updateResult, err := db.Collection(collectionName).UpdateOne(
		// 		context.TODO(),
		// 		bson.M{"_id": doc["_id"]},
		// 		bson.M{"$set": bson.M{"mdes": mft["to_des"]}},
		// 	)

		// 	if err != nil {
		// 		ErrorLog.Printf("Error Failed to update document CNo: %v\n", doc["cno"])
		// 		continue

		// 	}
		// 	if updateResult.ModifiedCount == 1 {
		// 		successCount++
		// 		log.Printf("%d/%d - CNo: %v updated successfully\n", docIndex, totalDocs, doc["cno"])
		// 		SuccessLog.Printf("%d/%d - CNo: %v updated successfully\n", docIndex, totalDocs, doc["cno"])
		// 	}
		// }
	}

	return nil
}
