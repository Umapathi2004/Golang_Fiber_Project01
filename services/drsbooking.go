package services

import (
	"GoFiber_Project01/DBConnection"
	"GoFiber_Project01/api_request"
	"GoFiber_Project01/config"
	"GoFiber_Project01/logs"
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func UpdateBookingDRS() error {
	var (
		collectionName = "booking_consignment"
		docIndex       int
		totalDocs      int
		successCount   int
	)
	logs.Logger()
	config.Init()
	DBConnection.DBConfig()
	cutOffDate := time.Date(2023, 4, 1, 0, 0, 0, 0, time.UTC)
	configration := config.Config
	db := DBConnection.DB
	MongoClient := DBConnection.MongoClient
	defer MongoClient.Disconnect(context.Background())

	query := mongo.Pipeline{
		{{
			Key: "$match", Value: bson.D{
				{Key: "$and", Value: bson.A{
					bson.D{{Key: "bdate", Value: bson.D{{Key: "$gte", Value: cutOffDate}}}},
					bson.D{{Key: "customer_code", Value: bson.D{{Key: "$in", Value: bson.A{"PAL101", "KRM209", "TAB742"}}}}},
				}},
			},
		}},
		{{
			Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "in_consignment"},
				{Key: "localField", Value: "_id"},
				{Key: "foreignField", Value: "cno"},
				{Key: "as", Value: "result"},
			},
		}},
		{{
			Key: "$unwind", Value: bson.D{
				{Key: "path", Value: "$result"},
				{Key: "preserveNullAndEmptyArrays", Value: false},
			},
		}},
		{{
			Key: "$match", Value: bson.D{
				{Key: "$and", Value: bson.A{
					bson.D{{Key: "result.txn", Value: "drs"}},
					bson.D{{Key: "result.status", Value: "DE"}},
					bson.D{{Key: "result.blr_server_update", Value: bson.D{{Key: "$exists", Value: false}}}},
				}},
			},
		}},
	}

	cursor, err := db.Collection(collectionName).Aggregate(context.Background(), query)
	if err != nil {
		logs.ErrorLog.Printf("failed to query documents: %v\n", err)
		return nil
	}
	defer cursor.Close(context.TODO())

	var data []bson.M
	if err = cursor.All(context.TODO(), &data); err != nil {
		logs.ErrorLog.Printf("failed to parse documents: %v\n", err)
		return nil
	}

	totalDocs = len(data)
	fmt.Printf("Founded %d docs\n", totalDocs)
	var bulkUpdateOps []mongo.WriteModel

	for _, doc := range data {
		param := getDrsBookingUrl(doc)
		if param != nil {
			docIndex++
			fmt.Printf("%d/%d - CNo:%s try to update\n", docIndex, totalDocs, doc["cno"])

			result := api_request.SendData("DRS-BK", configration["drsUrl"].(string), param, doc)
			if result["success"] == true {
				bulkUpdateOps = append(bulkUpdateOps, mongo.NewUpdateOneModel().SetFilter(bson.M{"_id": doc["result"].(bson.M)["_id"]}).SetUpdate(bson.M{"$set": bson.M{"blr_server_update": 1}}))
				successCount++
			}
		}
	}
	if len(bulkUpdateOps) > 0 {
		updateResult, err := db.Collection("in_consignment").BulkWrite(context.TODO(), bulkUpdateOps)
		if err != nil {
			logs.ErrorLog.Printf("Error updating documents: %v\n", err)
		} else {
			logs.SuccessLog.Printf("%d documents updated to the local DB\n", updateResult.ModifiedCount)
		}
	}

	stat := bson.M{
		"date":        time.Now(),
		"txn":         "DRS-BK",
		"server":      "BLR",
		"totalDocs":   totalDocs,
		"successDocs": successCount,
	}

	if _, err := db.Collection("main_server_update").InsertOne(context.Background(), bson.M(stat)); err != nil {
		logs.ErrorLog.Printf("failed to insert stats: %v", err)
	}
	return nil
}

func getDrsBookingUrl(r bson.M) map[string]interface{} {
	config.Init()
	configration := config.Config
	if r["cno"] == nil {
		return nil
	}

	dt, err := time.Parse("2006-01-02T15:04:05.000Z", r["updated_on"].(string))
	if err != nil {
		fmt.Println(err)
		return nil
	}

	param := map[string]interface{}{
		"SYS_DT":    time.Now().Format("01-02-2006"),
		"ORIGIN":    configration["branchCode"].(string),
		"DESTN":     configration["branchCode"].(string),
		"SYS_TM":    time.Now().Format("15:04"),
		"REMARKS":   "DE",
		"DRSNO":     r["txn_id"],
		"TYPEOFDOC": "INBOUND",
		"POD_NO":    r["cno"],
		"CC":        r["pccode"],
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
		"id":        configration["loginId"].(int16),
	}

	return param
}
