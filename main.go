package main

import (
	"context"
	"wander-wallet-tools/logger"
	"wander-wallet-tools/services"
)

func main() {
	logger.Init()
	logger.LogInfoLn("Logger initialized")

	ctx := context.Background()

	fbApp, err := services.NewFirebaseApp(ctx)
	if err != nil {
		logger.LogFatalLn("Firebase failed to initialize", err)
	}

	fsClient, err := fbApp.GetFirestore(ctx)
	if err != nil {
		logger.LogFatalLn("Firestore DB failed to initialize", err)
	}
	defer fsClient.Close()

	// costOfLivingService := services.NewCostOfLivingService(fsClient)
	// costOfLivingService.PopulateCostOfTravelData(ctx)

	// cleanupService := services.NewCostOfLivingCleanupService(fsClient)
	// cleanupService.CleanupCostOfLivingData(ctx)

	migrationService := services.NewCostOfLivingMigrationService(fsClient)
	migrationService.MigrateCostOfLivingData(ctx)
}