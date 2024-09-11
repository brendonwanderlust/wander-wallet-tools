package services

// import (
// 	"context"
// 	"encoding/csv"
// 	"fmt"
// 	"os"
// 	"strconv"
// 	"strings"

// 	"wander-wallet-tools/logger"
// 	"wander-wallet-tools/utils"

// 	"cloud.google.com/go/firestore"
// )

// type TopDestinationsService struct {
// 	firestoreClient *firestore.Client
// }

// func NewTopDestinationsService(client *firestore.Client) *TopDestinationsService {
// 	return &TopDestinationsService{
// 		firestoreClient: client,
// 	}
// }

// func (s *TopDestinationsService) ProcessAndSaveTopDestinations(ctx context.Context) error {
// 	// Read CSV file
// 	filePath := "C:/Users/B Carrasquillo/Downloads/top_cities.csv"
// 	destinations, err := s.readCSV(filePath)
// 	if err != nil {
// 		return fmt.Errorf("error reading CSV: %v", err)
// 	}

// 	// Save to Firestore
// 	err = s.saveToFirestore(ctx, destinations)
// 	if err != nil {
// 		return fmt.Errorf("error saving to Firestore: %v", err)
// 	}

// 	return nil
// }

// func (s *TopDestinationsService) readCSV(filePath string) ([]TopDestination, error) {
// 	file, err := os.Open(filePath)
// 	if err != nil {
// 		return nil, fmt.Errorf("error opening file: %v", err)
// 	}
// 	defer file.Close()

// 	reader := csv.NewReader(file)
// 	records, err := reader.ReadAll()
// 	if err != nil {
// 		return nil, fmt.Errorf("error reading CSV: %v", err)
// 	}

// 	var destinations []TopDestination
// 	for i, record := range records {
// 		if i == 0 { // Skip header row
// 			continue
// 		}
// 		if len(record) != 3 {
// 			logger.LogInfoLn(fmt.Sprintf("Skipping invalid record: %v", record))
// 			continue
// 		}
// 		rank, err := strconv.ParseInt(record[0], 10, 64)
// 		if err != nil {
// 			logger.LogInfoLn(fmt.Sprintf("Error parsing rank for record: %v. Error: %v", record, err))
// 			continue
// 		}
// 		destinations = append(destinations, TopDestination{
// 			Rank:    rank,
// 			City:    strings.TrimSpace(record[1]),
// 			Country: strings.TrimSpace(record[2]),
// 		})
// 	}

// 	return destinations, nil
// }

// func (s *TopDestinationsService) saveToFirestore(ctx context.Context, destinations []TopDestination) error {
// 	batch := s.firestoreClient.Batch()
// 	collection := s.firestoreClient.Collection("top-destinations")

// 	for _, dest := range destinations {
// 		docID := fmt.Sprintf("%s-%s", utils.NormalizeAndFormat(dest.City), utils.NormalizeAndFormat(dest.Country))
// 		docRef := collection.Doc(docID)
// 		batch.Set(docRef, dest)
// 	}

// 	_, err := batch.Commit(ctx)
// 	if err != nil {
// 		return fmt.Errorf("error committing batch: %v", err)
// 	}

// 	logger.LogInfoLn(fmt.Sprintf("Successfully saved %d destinations to Firestore", len(destinations)))
// 	return nil
// }
