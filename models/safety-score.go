package models

type SafetyScore struct {
	City    string `firestore:"city"`
	Country string `firestore:"country"`
	Score   int64  `firestore:"score"`
}
