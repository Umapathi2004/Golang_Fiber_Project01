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

func UpdateDRS() error {
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
		"txn":    "drs",
		"status": "DE",
		"cno": bson.M{
			"$not": bson.M{
				"$regex":   configration["branchCode"],
				"$options": "i",
			},
		},
		"$or": bson.A{
			bson.M{
				"blr_server_retry": bson.M{
					"$exists": false,
				},
			},
			bson.M{
				"$and": bson.A{
					bson.M{
						"blr_server_retry": bson.M{
							"$exists": true,
						},
					},
					bson.M{
						"blr_server_retry": bson.M{
							"$lt": 3,
						},
					},
				},
			},
		},
		"blr_server_update": bson.M{
			"$exists": false,
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

		param := getDrsUrl(doc)
		if param != nil {
			docIndex = idx + 1
			log.Printf("%d/%d - CNo: %v try to update", docIndex, totalDocs, doc["cno"])

			drsURL, _ := configration["drsUrl"].(string)
			result := api_request.SendData("DRS", drsURL, param, doc)
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
		"txn":         "DRS",
		"server":      "BLR",
		"totalDocs":   totalDocs,
		"successDocs": successCount,
	}
	if _, err := db.Collection("main_server_update").InsertOne(context.Background(), bson.M(stat)); err != nil {
		return fmt.Errorf("failed to insert stats: %v", err)
	}
	return nil
}

func getDrsUrl(r bson.M) map[string]interface{} {
	configration := config.Config
	if r["cno"] == nil {
		return nil
	}

	dt, err := time.Parse("2006-01-02T15:04:05.000Z", r["updated_on"].(string))
	if err != nil {
		log.Println(err)
		return nil
	}

	param := map[string]interface{}{
		"SYS_DT":    dt.Format("01-02-2006"),
		"ORIGIN":    configration["branchCode"],
		"DESTN":     configration["branchCode"],
		"SYS_TM":    dt.Format("15:04"),
		"REMARKS":   "DE",
		"DRSNO":     r["txn_id"],
		"TYPEOFDOC": "INBOUND",
		"POD_NO":    r["cno"],
		"CC": func() string {
			if r["pccode"] != nil {
				return r["pccode"].(string)
			}
			return "PC00"
		}(),
		"CONSIGNEE": func() string {
			if r["customer_address"] != nil {
				return r["customer_address"].(string)
			}
			return ""
		}(),
		"PHNO": func() string {
			if r["customer_mobile"] != nil {
				return r["customer_mobile"].(string)
			}
			return ""
		}(),
		"DEL_TM":    dt.Format("15:04"),
		"Id_PROOF":  "",
		"RELATION":  "",
		"longitude": 1,
		"latitude":  1,
		"RECPS":     "",
		"STAMP":     "",
		"SLNO":      1,
		"WEIGHT": func() float64 {
			if r["weight"] != nil {
				return r["weight"].(float64)
			}
			return 0.100
		}(),
		"CODAMOUNT": func() float64 {
			if r["cod_amount"] != nil {
				return r["cod_amount"].(float64)
			}
			return 0.00
		}(),
		"PIECES": func() int {
			if r["pieces"] != nil {
				return r["pieces"].(int)
			}
			return 1
		}(),
		"userid":    configration["branchCode"],
		"xmlorigin": configration["branchCode"],
		"id":        configration["loginId"],
	}

	return param
}
