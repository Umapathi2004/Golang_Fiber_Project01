package api_request

import (
	"GoFiber_Project01/logs"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func SendData(action string, url_ string, param map[string]interface{}, doc bson.M) map[string]interface{} {
	SuccessLog, ErrorLog := logs.Logger()
	// db, MongoClient, _ := DBConnection.InitMongoDB()
	// defer MongoClient.Disconnect(context.Background())

	u, err := url.Parse(url_)
	if err != nil {
		return handleError(action, doc, err, param)
	}
	q := u.Query()
	for key, value := range param {
		q.Set(key, fmt.Sprintf("%v", value))
	}
	u.RawQuery = q.Encode()

	resp, err := http.Get(u.String())
	fmt.Println(u.String(), "temp") //temp
	if err != nil {
		return handleError(action, doc, err, param)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		docs, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			return handleError(action, doc, err, param)
		}

		result := docs.Find("#lblResult").Text()
		if strings.Contains(result, "Successfully") {
			SuccessLog.Printf("%s - %s : updated to BLR server\n", action, doc["cno"])
			return map[string]interface{}{"success": true, "path": ""}
		}
	}

	retry := 0
	if val, ok := doc["blr_server_retry"].(int); ok {
		retry = val
	}
	retry += 1
	update := bson.M{
		"$set": bson.M{
			"txn":                         action,
			"error_detail":                "Failed to find success message",
			"param":                       param,
			"blr_server_retry":            retry,
			"blr_server_retry_updated_on": time.Now(),
		},
	}
	filter := bson.M{"_id": doc["cno"], "txn": action}
	opts := options.Update().SetUpsert(true)
	// collection := db.Collection("blr_server_error")
	// _, err = collection.UpdateOne(context.TODO(), filter, update, opts)
	// if err != nil {
	// 	ErrorLog.Printf("Error updating MongoDB: %v\n", err)
	// }
	fmt.Println(update, filter, opts)

	ErrorLog.Printf("%s - CNO: %s error: Failed to find success message\n", action, doc["cno"])
	return map[string]interface{}{"success": false, "path": ""}
}

func handleError(action string, doc bson.M, err error, param bson.M) map[string]interface{} {
	_, ErrorLog := logs.Logger()
	// DBConnection.InitMongoDB()
	// db := DBConnection.DB
	// MongoClient := DBConnection.MongoClient
	// defer func() {
	// 	if err := MongoClient.Disconnect(context.Background()); err != nil {
	// 		log.Printf("Error disconnecting MongoClient: %v", err)
	// 	}
	// }()

	retry := 0
	if val, ok := doc["blr_server_retry"].(int); ok {
		retry = val
	}
	retry += 1
	update := bson.M{
		"$set": bson.M{
			"txn":                         action,
			"error_detail":                err.Error(),
			"param":                       param,
			"blr_server_retry":            retry,
			"blr_server_retry_updated_on": time.Now(),
		},
	}
	filter := bson.M{"_id": doc["cno"], "txn": action}
	opts := options.Update().SetUpsert(true)
	// collection := db.Collection("blr_server_error")
	// _, dbErr := collection.UpdateOne(context.TODO(), filter, update, opts)
	// if dbErr != nil {
	// 	ErrorLog.Printf("Error updating MongoDB: %v", dbErr)
	// }
	fmt.Println(update, filter, opts)
	ErrorLog.Printf("%s - CNO: %s error: %v\n", action, doc["cno"], err)
	return map[string]interface{}{"success": false, "path": ""}
}
