package models

import (
	"fmt"
)

var CollectionName = "internet-speed-cache"

type InternetSpeed struct {
	LocationName       string   `firestore:"locationName"`
	Latitude           float64  `firestore:"latitude"`
	Longitude          float64  `firestore:"longitude"`
	DownloadSpeed_Mbps float64  `firestore:"downloadSpeed_mbps"`
	UploadSpeed_Mbps   float64  `firestore:"uploadSpeed_mbps"`
	Types              []string `firestore:"types"`
}

func GetInternetSpeedPathFromLocationMapping(mapping LocationMapping) string {
	return fmt.Sprintf("%v/%s", CollectionName, mapping.StandardName)
}
