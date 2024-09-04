package services

import (
	"context"
	"fmt"
	"strings"
	"unicode"

	"wander-wallet-tools/logger"

	"cloud.google.com/go/firestore"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

type CostOfLivingCleanupService struct {
	firestoreClient *firestore.Client
}

func NewCostOfLivingCleanupService(firestoreClient *firestore.Client) *CostOfLivingCleanupService {
	return &CostOfLivingCleanupService{
		firestoreClient: firestoreClient,
	}
}

func (s *CostOfLivingCleanupService) CleanupCostOfLivingData(ctx context.Context) error {
	logger.LogInfoLn("Starting cleanup of cost-of-living-staging data")

	// Get all documents from the collection
	docSnaps, err := s.firestoreClient.Collection("cost-of-travel-staging").Documents(ctx).GetAll()
	if err != nil {
		logger.LogErrorLn("Error getting all documents", err)
		return err
	}

	logger.LogInfoLn(fmt.Sprintf("Retrieved %d documents for processing", len(docSnaps)))

	// Process documents in batches
	batchSize := 500 // Firestore's maximum batch size
	for i := 0; i < len(docSnaps); i += batchSize {
		end := i + batchSize
		if end > len(docSnaps) {
			end = len(docSnaps)
		}

		err := s.processBatch(ctx, docSnaps[i:end])
		if err != nil {
			logger.LogErrorLn(fmt.Sprintf("Error processing batch %d to %d", i, end), err)
			return err
		}

		logger.LogInfoLn(fmt.Sprintf("Processed batch %d to %d", i, end))
	}

	logger.LogInfoLn("Completed cleanup of cost-of-living-staging data")
	return nil
}

func (s *CostOfLivingCleanupService) processBatch(ctx context.Context, docSnaps []*firestore.DocumentSnapshot) error {
	batch := s.firestoreClient.Batch()
	changesMade := false

	for _, doc := range docSnaps {
		oldID := doc.Ref.ID
		newID := s.cleanupDocID(oldID)

		if oldID != newID {
			// Delete the old document
			batch.Delete(doc.Ref)

			// Create a new document with the cleaned-up ID
			newDocRef := s.firestoreClient.Collection("cost-of-travel-staging").Doc(newID)
			batch.Set(newDocRef, doc.Data())

			changesMade = true

			logger.LogInfoLn(fmt.Sprintf("Cleaned up document ID: %s -> %s", oldID, newID))
		}
	}

	// Only commit if changes were made
	if changesMade {
		_, err := batch.Commit(ctx)
		if err != nil {
			logger.LogErrorLn("Error committing batch", err)
			return err
		}
	}

	return nil
}

func (s *CostOfLivingCleanupService) cleanupDocID(id string) string {
	// Remove accents and other diacritical marks
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	result, _, _ := transform.String(t, id)

	// Convert to lowercase
	result = strings.ToLower(result)

	// Remove any character that's not a letter, number, or hyphen
	result = strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsNumber(r) || r == '-' {
			return r
		}
		return -1
	}, result)

	// Replace multiple hyphens with a single hyphen
	for strings.Contains(result, "--") {
		result = strings.ReplaceAll(result, "--", "-")
	}

	// Trim hyphens from the start and end
	result = strings.Trim(result, "-")

	return result
}
