package services

import (
	"GoFiber_Project01/DBConnection"
	"GoFiber_Project01/api_request"
	"GoFiber_Project01/config"
	"GoFiber_Project01/logs"
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func UpdateInCommingManifest() error {
	var (
		collectionName = "in_consignment"
		docIndex       int
		totalDocs      int
	)
	logs.Logger()
	config.Init()
	DBConnection.DBConfig()
	configration := config.Config
	db := DBConnection.DB
	MongoClient := DBConnection.MongoClient
	defer MongoClient.Disconnect(context.Background())
	successCount := 0
	currentTime := time.Now()
	lastCutoff := currentTime.AddDate(0, 0, -10)

	query := bson.M{
		"created_on": bson.M{
			"$gte": lastCutoff,
			"$lte": currentTime,
		},
		"$or": []bson.M{
			{"blr_server_retry": bson.M{"$exists": false}},
			{
				"$and": []bson.M{
					{"blr_server_retry": bson.M{"$exists": true}},
					{"blr_server_retry": bson.M{"$lt": 3}},
				},
			},
		},
		"txn":               "in",
		"cno":               bson.M{"$not": bson.M{"$regex": configration["branchCode"], "$options": "i"}},
		"blr_server_update": bson.M{"$exists": false},
	}

	cursor, err := db.Collection(collectionName).Find(context.Background(), query)
	if err != nil {
		logs.ErrorLog.Printf("failed to query documents: %v\n", err)
		return nil
	}
	defer cursor.Close(context.Background())

	var data []bson.M
	if err := cursor.All(context.Background(), &data); err != nil {
		logs.ErrorLog.Printf("failed to parse documents: %v\n", err)
		return nil
	}
	totalDocs = len(data)
	log.Printf("Found %d docs", totalDocs)

	for idx, doc := range data {

		param := getInComeMftUrl(doc)
		if param != nil {
			docIndex = idx + 1
			log.Printf("%d/%d - CNo: %v try to update", docIndex, totalDocs, doc["cno"])

			manifestUrl, _ := configration["manifestUrl"].(string)
			result := api_request.SendData("IN-MFT", manifestUrl, param, doc)
			if result["success"] == true {
				updateResult, err := db.Collection(collectionName).UpdateOne(context.Background(), bson.M{"_id": doc["_id"]}, bson.M{
					"$set": bson.M{
						"blr_server_update":     1,
						"blr_server_updated_on": time.Now(),
					},
				})
				if err != nil || updateResult.ModifiedCount != 1 {
					log.Printf("Failed to update document: %v\n", doc["cno"])
					continue
				} else {
					successCount++
					logs.ErrorLog.Printf("%d/%d - %v updated to the local DB\n", docIndex, totalDocs, doc["cno"])
				}
			} else {
				retry := 0
				if docRetry, ok := doc["blr_server_retry"].(int); ok {
					retry = docRetry
				}
				_, err = db.Collection(collectionName).UpdateOne(context.Background(), bson.M{"_id": doc["_id"]}, bson.M{
					"$set": bson.M{
						"blr_server_retry":            retry + 1,
						"blr_server_retry_updated_on": time.Now(),
					}},
					options.Update().SetUpsert(true),
				)
				if err != nil {
					logs.ErrorLog.Printf("Failed to update document: %v\n", doc["cno"])
					continue
				}
			}
		}
	}

	stat := map[string]interface{}{
		"date":        time.Now(),
		"txn":         "IN-MFT",
		"server":      "BLR",
		"totalDocs":   totalDocs,
		"successDocs": successCount,
	}
	if _, err := db.Collection("main_server_update").InsertOne(context.Background(), bson.M(stat)); err != nil {
		logs.ErrorLog.Printf("failed to insert stats: %v", err)
	}
	return nil
}

func getInComeMftUrl(r bson.M) map[string]interface{} {
	config.Init()
	configration := config.Config
	orgin := r["cno"].(string)[:3]
	dt, err := time.Parse("2006-01-02T15:04:05.000Z", r["created_on"].(string))
	if err != nil {
		fmt.Println(err)
		return nil
	}

	param := map[string]interface{}{
		"SYS_DT":    dt.Format("01-02-2006"),
		"ORIGIN":    orgin,
		"DESTN":     configration["branchCode"].(string),
		"SYS_TM":    dt.Format("15:04"),
		"Remarks":   "DX",
		"Aremarks":  "DX",
		"MFSTNO":    r["txn_id"],
		"TYPEOFDOC": "INBOUND",
		"POD_NO":    r["cno"],
		"SLNO":      1,
		"WEIGHT": func() any {
			if r["weight"] != nil {
				return r["weight"]
			}
			return 0.100
		}(),
		"PIECES": func() any {
			if r["pieces"] != nil {
				return r["pieces"]
			}
			return 1
		}(),
		"VEHICELNO": "",
		"userid":    configration["branchCode"].(string),
		"XMLORIGIN": configration["branchCode"].(string),
		"id":        configration["loginID"].(int16),
	}

	return param
}
