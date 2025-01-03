package services

import (
	"GoFiber_Project01/DBConnection"
	"GoFiber_Project01/api_request"
	"GoFiber_Project01/config"
	"GoFiber_Project01/logs"
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func UpdateDRSTemp() error {
	var (
		collectionName = "in_consignment"
		docIndex       int
		successCount   int
		totalDocs      int
		db             *mongo.Database
		MongoClient    *mongo.Client
	)
	config.Init()
	DBConnection.DBConfig()
	logs.Logger()
	configration := config.Config
	db = DBConnection.DB
	MongoClient = DBConnection.MongoClient
	defer MongoClient.Disconnect(context.Background())

	query := bson.D{
		{Key: "txn", Value: "drs"},
		{Key: "blr_server_update", Value: 1},
	}

	cursor, err := db.Collection(collectionName).Find(context.TODO(), query)
	if err != nil {
		logs.ErrorLog.Printf("failed to find documents: %v\n", err)
		return nil
	}
	defer cursor.Close(context.TODO())

	var data []bson.M
	if err = cursor.All(context.TODO(), &data); err != nil {
		logs.ErrorLog.Printf("failed to read cursor: %v\n", err)
		return nil
	}
	totalDocs = len(data)
	log.Printf("Found %d docs\n", totalDocs)

	for _, doc := range data {

		txnID := doc["txn_id"].(string)
		txnIDInt, err := strconv.Atoi(txnID[4:])
		if err != nil {
			logs.ErrorLog.Printf("failed to convert txn_id to int: %v\n", err)
			continue
		}
		doc["txn_id"] = "DTRZ" + fmt.Sprintf("%07d", txnIDInt)

		param := getDrsUrlTemp(doc)
		if param != nil {
			docIndex++
			fmt.Printf("%d/%d - CNo:%s try to update\n", docIndex, totalDocs, doc["cno"])
			DrsUrl, _ := configration["DrsUrl"].(string)
			result := api_request.SendData("DRS", DrsUrl, param, doc)

			if result["success"] == true {
				updateResult, err := db.Collection(collectionName).UpdateOne(context.TODO(), bson.M{"_id": doc["_id"]}, bson.M{"$set": bson.M{
					"blr_server_update": 2,
					"txn_id":            doc["txn_id"],
				}})
				if err != nil || updateResult.ModifiedCount != 1 {
					log.Printf("Failed to update document: %v\n", doc["cno"])
					continue
				} else {
					successCount++
					logs.SuccessLog.Printf("%d/%d - %v updated to the local DB\n", docIndex, totalDocs, doc["cno"])
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
		logs.ErrorLog.Printf("failed to insert stats: %v\n", err)
	}
	return nil
}

func getDrsUrlTemp(r bson.M) map[string]interface{} {
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
		"id":        configration["loginID"].(int16),
	}
	return param
}
