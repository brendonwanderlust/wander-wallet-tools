package services

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"

	"wander-wallet-tools/logger"

	"cloud.google.com/go/firestore"
	"github.com/sirupsen/logrus"
)

type CostOfLivingService struct {
	firestoreClient *firestore.Client
}

type ColumnMapping struct {
	OriginalColumnName string
	NewColumnName      string
	Description        string
	DataType           string
}

func NewCostOfLivingService(firestoreClient *firestore.Client) *CostOfLivingService {
	return &CostOfLivingService{
		firestoreClient: firestoreClient,
	}
}

func (s *CostOfLivingService) PopulateCostOfTravelData(ctx context.Context) {
	// Step 1: Read and parse the column mappings
	columnMappings, err := s.readColumnMappings("C:\\Source\\wander-wallet-tools\\data\\cost_of_living\\column_mapping.csv")
	if err != nil {
		logger.LogErrorLn("failed to read column mappings: %v", err)
		return
	}

	// Step 2: Read and parse the CSV data
	records, err := s.readCSVData("C:\\Source\\wander-wallet-tools\\data\\cost_of_living\\col_data.csv")
	if err != nil {
		logger.LogErrorLn("failed to read CSV data: %v", err)
		return
	}

	// Step 3: Process data and rename columns
	data := s.processData(records, columnMappings)
	logger.LogInfoLn("Data renaming completed")

	// Step 4: Upload data to Firestore
	s.uploadToFirestore(ctx, data)
}

func (s *CostOfLivingService) readColumnMappings(filename string) (map[string]ColumnMapping, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	mappings := make(map[string]ColumnMapping)
	for _, record := range records[1:] { // Skip header
		mappings[record[0]] = ColumnMapping{
			OriginalColumnName: record[0],
			NewColumnName:      record[1],
			Description:        record[2],
			DataType:           record[3],
		}
	}

	return mappings, nil
}

func (s *CostOfLivingService) readCSVData(filename string) ([][]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	return reader.ReadAll()
}

func (s *CostOfLivingService) processData(records [][]string, columnMappings map[string]ColumnMapping) []map[string]interface{} {
	headers := records[0]
	var data []map[string]interface{}

	for _, record := range records[1:] {
		item := make(map[string]interface{})
		for i, value := range record {
			originalColumnName := headers[i]
			mapping, exists := columnMappings[originalColumnName]
			if exists {
				if value == "nan" {
					continue
				} else {
					convertedValue := s.convertValue(value, mapping.DataType)
					item[mapping.NewColumnName] = convertedValue
				}
			} else {
				item[originalColumnName] = value
			}
		}
		data = append(data, item)
	}

	return data
}

func (s *CostOfLivingService) convertValue(value string, dataType string) interface{} {
	switch dataType {
	case "float64":
		if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
			return floatVal
		}
	case "int":
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return value // Default to string if conversion fails
}

func (s *CostOfLivingService) uploadToFirestore(ctx context.Context, data []map[string]interface{}) error {
	// collection := s.firestoreClient.Collection("cost-of-travel-staging")
	const batchSize = 100
	var successCount int

	for i := 0; i < len(data); i += batchSize {
		batch := s.firestoreClient.Batch()
		end := i + batchSize
		if end > len(data) {
			end = len(data)
		}

		for _, item := range data[i:end] {
			city, cityOk := item["city"].(string)
			country, countryOk := item["country"].(string)

			if !cityOk || !countryOk {
				logger.LogErrorLn("Missing city or country for item", nil)
				continue // Skip this item if city or country is missing
			}

			// Create the document ID
			docID := fmt.Sprintf("%s-%s", strings.ToLower(strings.ReplaceAll(city, " ", "")), strings.ToLower(strings.ReplaceAll(country, " ", "")))
			// Replace spaces with hyphens and remove any special characters
			docID = strings.ReplaceAll(docID, " ", "")
			docID = strings.Map(func(r rune) rune {
				if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '`' || r == '\'' {
					return r
				}
				logger.LogInfoWithFields("Removed rune: ", logrus.Fields{"Rune": r})
				return -1
			}, docID)

			// docRef := collection.Doc(docID)
			// batch.Set(docRef, item)
		}

		_, err := batch.Commit(ctx)
		if err != nil {
			return fmt.Errorf("failed to commit batch starting at index %d: %v", i, err)
		}
		successCount += end - i
		logger.LogInfoLn(fmt.Sprintf("Successfully uploaded batch of %d items to Firestore", end-i))
	}

	logger.LogInfoLn(fmt.Sprintf("Successfully uploaded a total of %d items to Firestore", successCount))
	return nil
}
