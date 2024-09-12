package models

type TopDestination struct {
	Id                string   `firestore:"id"`
	City              string   `firestore:"city"`
	Country           string   `firestore:"country"`
	Rank              int64    `firestore:"rank"`
	PlaceId           string   `firestore:"placeId"`
	PhotoUri1         string   `firestore:"photoUri1"`
	PhotoUri2         string   `firestore:"photoUri2"`
	Photos            []string `firestore:"photos"`
	DownloadAvg       float64  `firestore:"downloadAvg"`
	SafetyScore       int64    `firestore:"safetyScore"`
	CostOfLivingScore float64  `firestore:"costOfLivingScore"`
}
