package main

import (
	"Blr_server_update/logs"
	"Blr_server_update/services"
	"fmt"
	"os"
)

func main() {
	_, ErrorLog := logs.Logger()
	Service := os.Args[1]

	switch Service {
	case "bookingupdate":
		services.UpdateBookingConsignment()
	case "ptp":
		services.UpdatePTP()
	case "ptpmaa":
		services.UpdatePTP_MAA()
	case "drstd":
		services.UpdateDrsTd()
	case "drs":
		services.UpdateDRS()
	case "drsbooking":
		services.UpdateBookingDRS()
	case "drstemp":
		services.UpdateDRSTemp()
	case "inmft":
		services.UpdateInCommingManifest()
	case "outmft":
		services.UpdateOutCommingManifest()
	case "outmftfix":
		services.Fixes()
	default:
		fmt.Printf("Invalid Service %v\n", Service)
		ErrorLog.Printf("Invalid Service %v\n", Service)
	}
}
