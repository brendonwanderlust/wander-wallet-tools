package main

import (
	"context"
	"fmt"
	"wander-wallet-tools/config"
	"wander-wallet-tools/logger"
	"wander-wallet-tools/models"
	"wander-wallet-tools/services"

	"cloud.google.com/go/bigquery"
	"cloud.google.com/go/firestore"
	"googlemaps.github.io/maps"
)

func main() {
	logger.Init()
	logger.LogInfoLn("Logger initialized")

	ctx := context.Background()
	cfg := config.NewConfig(models.Dev)

	fbApp, err := services.NewFirebaseApp(ctx)
	if err != nil {
		logger.LogFatalLn("Firebase failed to initialize", err)
	}

	fsClient, err := fbApp.GetFirestore(ctx)
	if err != nil {
		logger.LogFatalLn("Firestore DB failed to initialize", err)
	}
	defer fsClient.Close()

	mapsClient, err := maps.NewClient(maps.WithAPIKey(cfg.GoogleMapsAPIKey))
	if err != nil {
		logger.LogFatalLn("error creating Google Maps client: %v", err)
	}

	bqClient, err := bigquery.NewClient(ctx, cfg.FirebaseProjectId)
	if err != nil {
		logger.LogFatalLn("Failed to create BigQuery client: %v", err)
	}
	defer bqClient.Close()

	// costOfLivingService := services.NewCostOfLivingService(fsClient)
	// costOfLivingService.PopulateCostOfTravelData(ctx)

	// cleanupService := services.NewCostOfLivingCleanupService(fsClient)
	// cleanupService.CleanupCostOfLivingData(ctx)

	// migrationService := services.NewCostOfLivingMigrationService(fsClient)
	// migrationService.MigrateCostOfLivingData(ctx)

	// analyzerService := services.NewCostOfLivingAnalyzerService(fsClient)
	// err = analyzerService.AnalyzeAndStoreData(ctx)
	// if err != nil {
	// 	logger.LogFatalLn("Failed to analyze and store data: %v", err)
	// }
	// topDestService := services.NewTopDestinationsService(fsClient)
	// err = topDestService.ProcessAndSaveTopDestinations(ctx)
	// if err != nil {
	// 	logger.LogFatalLn("Failed to process and save top destinations: %v", err)
	// }

	enrichService := services.NewTopDestinationEnrichmentService(bqClient, fsClient, cfg, mapsClient)
	err = enrichService.EnrichTopDestinations(ctx)
	if err != nil {
		logger.LogFatalLn("Failed to enrich top destinations: %v", err)
	}
	// errors := copyCitySafetyDocument(ctx, fsClient, "minneapolisstpaul-unitedstates", "minneapolis-unitedstates")
	// if errors != nil {
	// 	logger.LogErrorLn("Failed", errors)
	// }

}

func copyCitySafetyDocument(ctx context.Context, client *firestore.Client, oldDocID, newDocID string) error {
	// Step 1: Retrieve the original document
	collectionName := "top-destinations"
	docRef := client.Collection(collectionName).Doc(oldDocID)
	docSnap, err := docRef.Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to retrieve original document: %v", err)
	}

	// Step 2: Check if the new document already exists
	newDocRef := client.Collection(collectionName).Doc(newDocID)
	_, err = newDocRef.Get(ctx)
	if err == nil {
		return fmt.Errorf("document with ID %s already exists", newDocID)
	} else if err != nil {
		fmt.Errorf("error checking new document: %v", err)
	}

	// Step 3: Create a copy of the document data
	data := docSnap.Data()

	// Step 4: Save the new document
	_, err = newDocRef.Set(ctx, data)
	if err != nil {
		return fmt.Errorf("failed to create new document: %v", err)
	}

	fmt.Printf("Successfully copied document from %s to %s\n", oldDocID, newDocID)
	return nil
}
