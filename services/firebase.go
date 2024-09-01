package services

import (
	"context"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
)

type FirebaseApp struct {
	App *firebase.App
}

func NewFirebaseApp(ctx context.Context) (*FirebaseApp, error) {
	conf := &firebase.Config{ProjectID: "travel-buddy-ionic-app"}
	app, err := firebase.NewApp(ctx, conf)
	if err != nil {
		return nil, err
	}
	return &FirebaseApp{App: app}, nil
}

func (fa *FirebaseApp) GetFirestore(ctx context.Context) (*firestore.Client, error) {
	client, err := fa.App.Firestore(ctx)
	if err != nil {
		return nil, err
	}
	return client, nil
}
