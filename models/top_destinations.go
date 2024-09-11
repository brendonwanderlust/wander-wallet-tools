package models

type TopDestination struct {
	Rank    int64  `firestore:"rank"`
	City    string `firestore:"city"`
	Country string `firestore:"country"`
}
