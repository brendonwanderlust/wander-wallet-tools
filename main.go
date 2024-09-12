package main

import (
	"context"
	"wander-wallet-tools/config"
	"wander-wallet-tools/logger"
	"wander-wallet-tools/models"
	"wander-wallet-tools/services"

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

	enrichService := services.NewTopDestinationEnrichmentService(fsClient, cfg, mapsClient)
	err = enrichService.EnrichTopDestinations(ctx)
	if err != nil {
		logger.LogFatalLn("Failed to enrich top destinations: %v", err)
	}

}
