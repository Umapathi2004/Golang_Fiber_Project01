package config

import (
	"Blr_server_update/logs"
	"os"

	"github.com/joho/godotenv"
)

var Config map[string]interface{}

func Init() map[string]interface{} {
	_, ErrorLog := logs.Logger()
	err := godotenv.Load(".env")
	if err != nil {
		ErrorLog.Printf("Error loading .env file: %v\n", err)
	}

	Base_URL := os.Getenv("BASE_URL")
	Config = map[string]interface{}{
		"baseUrl":          Base_URL,
		"drsTdUrl":         Base_URL + "DRSTD.aspx?",
		"drsUrl":           Base_URL + "DRS.aspx?",
		"ptpUrl":           Base_URL + "iom.aspx?",
		"bookingUpdateUrl": Base_URL + "refdata.aspx",
		"manifestUrl":      Base_URL + "mft.aspx?",
		"branchCode":       "TRZ",
		"backLogDays":      3,
		"loginId":          5024,
	}
	return Config
}
