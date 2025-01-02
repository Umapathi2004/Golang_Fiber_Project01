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
	"go.mongodb.org/mongo-driver/mongo/options"
)

func UpdatePTP() error {
	var (
		collectionName = "in_consignment"
		docIndex       int
		totalDocs      int
	)
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
		"txn": "ptp",
		"cno": bson.M{
			"$not": bson.M{"$regex": configration["branchCode"], "$options": "i"},
		},
		"blr_server_update": bson.M{"$exists": false},
		"$or": bson.A{
			bson.M{
				"blr_server_retry": bson.M{"$exists": false},
			},
			bson.M{
				"blr_server_retry": bson.M{"$exists": true, "$lt": 3},
			},
		},
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

		param := getPtpUrl(doc)
		if param != nil {
			docIndex = idx + 1
			log.Printf("%d/%d - CNo: %v try to update", docIndex, totalDocs, doc["cno"])

			ptpURL, _ := configration["ptpURL"].(string)
			result := api_request.SendData("PTP", ptpURL, param, doc)
			if result["success"] == true {
				updateResult, err := db.Collection(collectionName).UpdateOne(context.Background(), bson.M{"_id": doc["_id"]}, bson.M{
					"$set": bson.M{
						"blr_server_update":     1,
						"blr_server_updated_on": time.Now(),
					},
				})
				if err != nil || updateResult.ModifiedCount != 1 {
					log.Printf("Failed to update document: %v", doc["cno"])
				} else {
					successCount++
					log.Printf("%d/%d - %v updated to the local DB", docIndex, totalDocs, doc["cno"])
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
					log.Printf("Failed to update document: %v", doc["cno"])
				}
			}
		}
	}

	stat := map[string]interface{}{
		"date":        time.Now(),
		"txn":         "PTP",
		"server":      "BLR",
		"totalDocs":   totalDocs,
		"successDocs": successCount,
	}
	if _, err := db.Collection("main_server_update").InsertOne(context.Background(), bson.M(stat)); err != nil {
		return fmt.Errorf("failed to insert stats: %v", err)
	}
	return nil
}

func getPtpUrl(r bson.M) map[string]interface{} {
	configration := config.Config
	cno, ok := r["cno"].(string)
	if !ok {
		log.Printf("Invalid CNo in record: %v", r)
		return nil
	}

	createdOnStr, ok := r["created_on"].(string)
	if !ok {
		log.Printf("Invalid created_on field in record: %v", r)
		return nil
	}

	createdOn, err := time.Parse(time.RFC3339, createdOnStr)
	if err != nil {
		log.Printf("Error parsing created_on for CNo %s: %v", cno, err)
		return nil
	}
	sample := map[string]interface{}{
		"SYS_DT": createdOn.Format("2006-01-02"),
		"ORIGIN": configration["branchCode"],
		"DESTN": func() string {
			if r["pccode"] != nil {
				return r["pccode"].(string)
			}
			return "PC00"
		}(),
		"SYS_TM":    createdOn.Format("15:04"),
		"Remarks":   "ND",
		"Aremarks":  "ND",
		"IOMNO":     r["txn_id"].(string),
		"TYPEOFDOC": "INBOUND",
		"pod_no":    cno,
		"SLNO":      1,
		"Weight": func() float64 {
			if r["weight"] != nil {
				return r["weight"].(float64)
			}
			return 0.100
		}(),
		"CODAMOUNT": 0.00,
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
	return sample
}
