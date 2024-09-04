package services

import (
	"context"
	"fmt"
	"time"

	"wander-wallet-tools/logger"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

type CostOfLivingMigrationService struct {
	firestoreClient *firestore.Client
}

func NewCostOfLivingMigrationService(firestoreClient *firestore.Client) *CostOfLivingMigrationService {
	return &CostOfLivingMigrationService{
		firestoreClient: firestoreClient,
	}
}

func (s *CostOfLivingMigrationService) MigrateCostOfLivingData(ctx context.Context) error {
	logger.LogInfoLn("Starting migration of cost-of-living data")

	sourceCollection := s.firestoreClient.Collection("cost-of-travel-staging")
	destinationCollection := s.firestoreClient.Collection("cost-of-living")

	// Get all documents from the source collection
	iter := sourceCollection.Documents(ctx)
	defer iter.Stop()

	batch := s.firestoreClient.Batch()
	batchSize := 0
	maxBatchSize := 500 // Firestore limit
	totalMigrated := 0

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			logger.LogErrorLn("Error iterating documents", err)
			return err
		}

		// Create a new document in the destination collection with the same ID and data
		newDocRef := destinationCollection.Doc(doc.Ref.ID)
		batch.Set(newDocRef, doc.Data())

		batchSize++
		totalMigrated++

		// If we've reached the batch limit, commit and start a new batch
		if batchSize == maxBatchSize {
			if err := s.commitBatchWithRetry(ctx, batch); err != nil {
				return err
			}
			batch = s.firestoreClient.Batch()
			batchSize = 0
			logger.LogInfoLn(fmt.Sprintf("Migrated batch of %d documents. Total migrated: %d", maxBatchSize, totalMigrated))
		}
	}

	// Commit any remaining documents
	if batchSize > 0 {
		if err := s.commitBatchWithRetry(ctx, batch); err != nil {
			return err
		}
		logger.LogInfoLn(fmt.Sprintf("Migrated final batch of %d documents. Total migrated: %d", batchSize, totalMigrated))
	}

	logger.LogInfoLn(fmt.Sprintf("Completed migration of cost-of-living data. Total documents migrated: %d", totalMigrated))
	return nil
}

func (s *CostOfLivingMigrationService) commitBatchWithRetry(ctx context.Context, batch *firestore.WriteBatch) error {
	maxRetries := 3
	var lastErr error

	for i := 0; i < maxRetries; i++ {
		_, err := batch.Commit(ctx)
		if err == nil {
			return nil
		}
		lastErr = err
		logger.LogErrorLn(fmt.Sprintf("Error committing batch (attempt %d/%d): %v", i+1, maxRetries, err), err)
		time.Sleep(time.Second * time.Duration(i+1)) // Simple exponential backoff
	}

	logger.LogErrorLn(fmt.Sprintf("Failed to commit batch after %d attempts", maxRetries), lastErr)
	return lastErr
}
