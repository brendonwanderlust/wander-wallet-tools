package services

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"sort"
	"strings"
	"wander-wallet-tools/logger"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

type SafetyRecord struct {
	DocID string
	Score int64
}

type CostOfLivingRecord struct {
	DocID string
	Score float64
}

type CombinedRecord struct {
	DocID       string
	SafetyScore int64
	COLScore    float64
	AvgScore    float64
	Country     string
}

var excludedCountries = map[string]bool{
	"ukraine":  true,
	"russia":   true,
	"israel":   true,
	"belarus":  true,
	"pakistan": true,
	"iraq":     true,
	"syria":    true,
	"zimbabwe": true,
}

var countryLimits = map[string]int{
	"india":        10,
	"poland":       5,
	"unitedstates": 50,
}
var defaultCountryLimit int = 150

type TopDestinationsService struct {
	firestoreClient *firestore.Client
}

func NewTopDestinationsService(client *firestore.Client) *TopDestinationsService {
	return &TopDestinationsService{
		firestoreClient: client,
	}
}

func (s *TopDestinationsService) fetchSafetyData(ctx context.Context) ([]SafetyRecord, error) {
	var safetyRecords []SafetyRecord
	iter := s.firestoreClient.Collection("city-safety").OrderBy("score", firestore.Desc).Limit(2000).Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error fetching safety data: %v", err)
		}
		var record SafetyRecord
		record.DocID = doc.Ref.ID
		record.Score = doc.Data()["score"].(int64)
		safetyRecords = append(safetyRecords, record)
	}
	return safetyRecords, nil
}

func (s *TopDestinationsService) fetchCostOfLivingData(ctx context.Context) ([]CostOfLivingRecord, error) {
	var colRecords []CostOfLivingRecord
	iter := s.firestoreClient.Collection("cost-of-living-analytics").OrderBy("scores.overall", firestore.Asc).Limit(5000).Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error fetching cost of living data: %v", err)
		}
		var record CostOfLivingRecord
		record.DocID = doc.Ref.ID
		record.Score = doc.Data()["scores"].(map[string]interface{})["overall"].(float64)
		colRecords = append(colRecords, record)
	}
	return colRecords, nil
}

func (s *TopDestinationsService) processCombinedData(safetyRecords []SafetyRecord, colRecords []CostOfLivingRecord) []CombinedRecord {
	safetyMap := make(map[string]int64)
	for _, record := range safetyRecords {
		safetyMap[record.DocID] = record.Score
	}

	var combinedRecords []CombinedRecord
	countryCounts := make(map[string]int)

	for _, colRecord := range colRecords {
		country := extractCountry(colRecord.DocID)

		if excludedCountries[country] {
			continue
		}

		limit := defaultCountryLimit
		if l, exists := countryLimits[country]; exists {
			limit = l
		}

		if countryCounts[country] >= limit {
			logger.LogInfoLn(fmt.Sprintf("Skipping record for %s: limit reached (%d)", country, limit))
			continue
		}

		if safetyScore, exists := safetyMap[colRecord.DocID]; exists {
			avgScore := (float64(safetyScore) + (100 - colRecord.Score)) / 2 // Invert COL score
			combinedRecords = append(combinedRecords, CombinedRecord{
				DocID:       colRecord.DocID,
				SafetyScore: safetyScore,
				COLScore:    100 - colRecord.Score, // Invert COL score
				AvgScore:    avgScore,
				Country:     country,
			})
			countryCounts[country]++
		}
	}

	sort.Slice(combinedRecords, func(i, j int) bool {
		return combinedRecords[i].AvgScore > combinedRecords[j].AvgScore
	})

	return combinedRecords
}

func (s *TopDestinationsService) GenerateTopDestinationsCSV(ctx context.Context) error {
	safetyRecords, err := s.fetchSafetyData(ctx)
	if err != nil {
		return err
	}

	colRecords, err := s.fetchCostOfLivingData(ctx)
	if err != nil {
		return err
	}

	combinedRecords := s.processCombinedData(safetyRecords, colRecords)

	file, err := os.Create(os.ExpandEnv("C:/Users/B Carrasquillo/Documents/top_destinations.csv"))
	if err != nil {
		return fmt.Errorf("error creating CSV file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	headers := []string{"DocID", "Safety Score", "Cost of Living Score", "Average Score"}
	if err := writer.Write(headers); err != nil {
		return fmt.Errorf("error writing CSV headers: %v", err)
	}

	for _, record := range combinedRecords {
		row := []string{
			record.DocID,
			fmt.Sprintf("%.2f", record.SafetyScore),
			fmt.Sprintf("%.2f", record.COLScore),
			fmt.Sprintf("%.2f", record.AvgScore),
		}
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("error writing CSV row: %v", err)
		}
	}

	return nil
}

func extractCountry(docID string) string {
	parts := strings.Split(docID, "-")
	if len(parts) > 1 {
		return strings.ToLower(parts[len(parts)-1])
	}
	return ""
}
