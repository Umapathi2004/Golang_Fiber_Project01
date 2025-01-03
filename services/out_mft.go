package services

import (
	"GoFiber_Project01/DBConnection"
	"GoFiber_Project01/api_request"
	"GoFiber_Project01/config"
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

func UpdateOutCommingManifest() error {
	var (
		collectionName = "booking_consignment"
		docIndex       int
		totalDocs      int
	)
	configration := config.Config
	db := DBConnection.DB
	MongoClient := DBConnection.MongoClient
	defer MongoClient.Disconnect(context.Background())
	successCount := 0

	query := bson.M{
		"mdes":              bson.M{"$exists": true},
		"txn_id":            bson.M{"$exists": true},
		"blr_server_update": bson.M{"$exists": false},
	}

	cursor, err := db.Collection(collectionName).Find(context.Background(), query)
	if err != nil {
		return fmt.Errorf("failed to query documents: %v", err)
	}
	defer cursor.Close(context.Background())

	var data []bson.M
	if err := cursor.All(context.Background(), &data); err != nil {
		return fmt.Errorf("failed to parse documents: %v", err)
	}
	totalDocs = len(data)
	log.Printf("Found %d docs", totalDocs)

	for idx, doc := range data {
		if doc["txn_id"] == "" {
			continue
		}
		param := getOutComeMftUrl(doc)
		if param != nil {
			docIndex = idx + 1
			log.Printf("%d/%d - CNo: %v try to update", docIndex, totalDocs, doc["cno"])

			manifestUrl, _ := configration["manifestUrl"].(string)
			result := api_request.SendData("IN-MFT", manifestUrl, param, doc)
			if result["success"] == true {
				updateResult, err := db.Collection(collectionName).UpdateOne(context.Background(), bson.M{"_id": doc["_id"]}, bson.M{
					"$set": bson.M{
						"blr_server_update": 1,
					},
				})
				if err != nil || updateResult.ModifiedCount != 1 {
					log.Printf("Failed to update document: %v", doc["cno"])
				} else {
					successCount++
					log.Printf("%d/%d - %v updated to the local DB", docIndex, totalDocs, doc["cno"])
				}
			}
		}
	}
	stat := map[string]interface{}{
		"date":        time.Now(),
		"txn":         "OUT-MFT",
		"server":      "BLR",
		"totalDocs":   totalDocs,
		"successDocs": successCount,
	}
	if _, err := db.Collection("main_server_update").InsertOne(context.Background(), bson.M(stat)); err != nil {
		return fmt.Errorf("failed to insert stats: %v", err)
	}
	return nil
}

func getOutComeMftUrl(r bson.M) map[string]interface{} {
	configration := config.Config
	dt, err := time.Parse("2006-01-02T15:04:05.000Z", r["created_on"].(string))
	if err != nil {
		fmt.Println(err)
		return nil
	}

	param := map[string]interface{}{
		"SYS_DT":    dt.Format("01-02-2006"),
		"ORIGIN":    configration["branchCode"],
		"DESTN":     r["mdes"],
		"SYS_TM":    dt.Format("15:04"),
		"Remarks":   "DX",
		"Aremarks":  "DX",
		"MFSTNO":    r["txn_id"],
		"TYPEOFDOC": "OUTBOUND",
		"POD_NO": func() any {
			if r["cno"] != nil {
				return r["cno"]
			}
			return r["_id"]
		}(),
		"SLNO": 1,
		"WEIGHT": func() float64 {
			if r["weight"] != nil {
				return r["weight"].(float64)
			}
			return 0.100
		}(),
		"PIECES": func() int {
			if r["pieces"] != nil {
				return r["pieces"].(int)
			}
			return 1
		}(),
		"VEHICELNO": "",
		"userid":    configration["branchCode"],
		"XMLORIGIN": configration["branchCode"],
		"id":        configration["loginID"],
	}

	return param
}
