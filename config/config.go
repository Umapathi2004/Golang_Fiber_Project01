package config

import (
	"GoFiber_Project01/logs"
	"os"

	"github.com/joho/godotenv"
)

var Config map[string]interface{}

func Init() {
	logs.Logger()
	err := godotenv.Load(".env")
	if err != nil {
		logs.ErrorLog.Printf("Error loading .env file\n")
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
}
