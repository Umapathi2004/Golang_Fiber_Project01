package services

import (
	"Blr_server_update/DBConnection"
	"Blr_server_update/api_request"
	"Blr_server_update/config"
	"Blr_server_update/helpers"
	"Blr_server_update/logs"
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

func UpdateDRS() error {
	var (
		collectionName = "in_consignment"
		docIndex       int
		totalDocs      int
	)
	_, ErrorLog := logs.Logger()
	configration := config.Init()
	db, MongoClient, err := DBConnection.InitMongoDB()
	if err != nil {
		log.Printf("Error Connected to MongoDB: %v\n", err)
		return nil
	}
	
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
		ErrorLog.Printf("failed to query documents: %v\n", err)
		return nil
	}
	defer cursor.Close(context.Background())

	var data []bson.M
	if err := cursor.All(context.Background(), &data); err != nil {
		ErrorLog.Printf("failed to parse documents: %v\n", err)
		return nil
	}
	totalDocs = len(data)
	log.Printf("Found %d docs", totalDocs)
	if totalDocs == 0 {
		ErrorLog.Printf("Found %d docs\n", totalDocs)
		return nil
	}
	for idx, doc := range data {

		param := getDrsUrl(doc)
		if param != nil {
			docIndex = idx + 1
			log.Printf("%d/%d - CNo: %v try to update", docIndex, totalDocs, doc["cno"])

			drsURL, _ := configration["drsUrl"].(string)
			result := api_request.SendData("DRS", drsURL, param, doc)
			fmt.Println(result)
			// if result["success"] == true {
			// 	updateResult, err := db.Collection(collectionName).UpdateOne(context.Background(), bson.M{"_id": doc["_id"]}, bson.M{
			// 		"$set": bson.M{
			// 			"blr_server_update":     1,
			// 			"blr_server_updated_on": time.Now(),
			// 		},
			// 	})
			// 	if err != nil {
			// 		ErrorLog.Printf("Error Failed to update document CNo: %v\n", doc["cno"])
			// 		continue

			// 	}
			// 	if updateResult.ModifiedCount == 1 {
			// 		successCount++
			// 		log.Printf("%d/%d - CNo: %v updated successfully\n", docIndex, totalDocs, doc["cno"])
			// 		SuccessLog.Printf("%d/%d - CNo: %v updated successfully\n", docIndex, totalDocs, doc["cno"])
			// 	}
			// } else {
			// 	retry := 0
			// 	if docRetry, ok := doc["blr_server_retry"].(int); ok {
			// 		retry = docRetry
			// 	}
			// 	_, err = db.Collection(collectionName).UpdateOne(context.Background(), bson.M{"_id": doc["_id"]}, bson.M{
			// 		"$set": bson.M{
			// 			"blr_server_retry":            retry + 1,
			// 			"blr_server_retry_updated_on": time.Now(),
			// 		}},
			// 		options.Update().SetUpsert(true),
			// 	)
			// 	if err != nil {
			// 		ErrorLog.Printf("Failed to update document: %v\n", doc["cno"])
			// 		continue
			// 	}
			// }
		}
	}

	stat := map[string]interface{}{
		"date":        time.Now(),
		"txn":         "DRS",
		"server":      "BLR",
		"totalDocs":   totalDocs,
		"successDocs": successCount,
	}
	// if _, err := db.Collection("main_server_update").InsertOne(context.Background(), bson.M(stat)); err != nil {
	// 	ErrorLog.Printf("failed to insert stats: %v\n", err)
	// }
	fmt.Println(stat)
	return nil
}

func getDrsUrl(r bson.M) map[string]interface{} {
	configration := config.Init()
	if r["cno"] == nil {
		return nil
	}

	err, dt := helpers.StringToDateConverter(r["updated_on"])
	if err {
		fmt.Println(err)
		return nil
	}

	param := map[string]interface{}{
		"SYS_DT":    dt.Format("01-02-2006"),
		"ORIGIN":    configration["branchCode"].(string),
		"DESTN":     configration["branchCode"].(string),
		"SYS_TM":    dt.Format("15:04"),
		"REMARKS":   "DE",
		"DRSNO":     r["txn_id"],
		"TYPEOFDOC": "INBOUND",
		"POD_NO":    r["cno"],
		"CC": func() any {
			if r["pccode"] != nil {
				return r["pccode"]
			}
			return "PC00"
		}(),
		"CONSIGNEE": func() any {
			if r["customer_address"] != nil {
				return r["customer_address"]
			}
			return ""
		}(),
		"PHNO": func() any {
			if r["customer_mobile"] != nil {
				return r["customer_mobile"]
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
		"WEIGHT": func() any {
			if r["weight"] != nil {
				return r["weight"]
			}
			return 0.100
		}(),
		"CODAMOUNT": func() any {
			if r["cod_amount"] != nil {
				return r["cod_amount"]
			}
			return 0.00
		}(),
		"PIECES": func() any {
			if r["pieces"] != nil {
				return r["pieces"]
			}
			return 1
		}(),
		"userid":    configration["branchCode"].(string),
		"xmlorigin": configration["branchCode"].(string),
		"id":        configration["loginId"],
	}

	return param
}
