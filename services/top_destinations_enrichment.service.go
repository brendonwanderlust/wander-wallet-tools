package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"wander-wallet-tools/config"
	"wander-wallet-tools/logger"
	"wander-wallet-tools/models"

	"cloud.google.com/go/firestore"
	"github.com/sirupsen/logrus"
)

type TopDestination struct {
	Id                string   `firestore:"id"`
	City              string   `firestore:"city"`
	Country           string   `firestore:"country"`
	Rank              int64    `firestore:"rank"`
	PlaceId           string   `firestore:"placeId"`
	PhotoUri1         string   `firestore:"photoUri1"`
	PhotoUri2         string   `firestore:"photoUri2"`
	Photos            []string `firestore:"photos"`
	DownloadAvg       float64  `firestore:"downloadAvg"`
	SafetyScore       int64    `firestore:"safetyScore"`
	CostOfLivingScore float64  `firestore:"costOfLivingScore"`
}

type TopDestinationEnrichmentService struct {
	firestoreClient *firestore.Client
	cfg             *config.Config
}

func NewTopDestinationEnrichmentService(client *firestore.Client, cfg *config.Config) *TopDestinationEnrichmentService {
	return &TopDestinationEnrichmentService{
		firestoreClient: client,
		cfg:             cfg,
	}
}

func (s *TopDestinationEnrichmentService) EnrichTopDestinations(ctx context.Context) error {
	destinations, err := s.getTopDestinations(ctx)
	if err != nil {
		logger.LogErrorWithFields("Failed to fetch top destinations", logrus.Fields{"Error": err.Error()})
		return err
	}

	for _, dest := range destinations {
		err := s.enrichDestination(ctx, &dest)
		if err != nil {
			logger.LogErrorWithFields("Failed to enrich destination", logrus.Fields{
				"Error":   err.Error(),
				"City":    dest.City,
				"Country": dest.Country,
			})
			continue
		}

		err = s.saveDestination(ctx, dest)
		if err != nil {
			logger.LogErrorWithFields("Failed to save enriched destination", logrus.Fields{
				"Error":   err.Error(),
				"City":    dest.City,
				"Country": dest.Country,
			})
		}
	}

	return nil
}

func (s *TopDestinationEnrichmentService) getTopDestinations(ctx context.Context) ([]TopDestination, error) {
	var destinations []TopDestination
	docs, err := s.firestoreClient.Collection("top-destinations").OrderBy("rank", firestore.Asc).Limit(20).Documents(ctx).GetAll()
	if err != nil {
		return nil, err
	}

	for _, doc := range docs {
		var dest TopDestination
		if err := doc.DataTo(&dest); err != nil {
			logger.LogErrorWithFields("Failed to parse top destination", logrus.Fields{
				"Error": err.Error(),
				"DocID": doc.Ref.ID,
			})
			continue
		}
		dest.Id = doc.Ref.ID
		destinations = append(destinations, dest)
	}

	return destinations, nil
}

func (s *TopDestinationEnrichmentService) enrichDestination(ctx context.Context, dest *TopDestination) error {
	mapping, err := s.getLocationMapping(ctx, dest.City, dest.Country)
	if err != nil {
		return fmt.Errorf("failed to get location mapping: %v", err)
	}

	if mapping.InternetSpeedRef != nil {
		internetSpeed, err := s.getInternetSpeed(ctx, mapping.InternetSpeedRef)
		if err != nil {
			logger.LogErrorWithFields("Failed to get internet speed", logrus.Fields{
				"Error":   err.Error(),
				"City":    dest.City,
				"Country": dest.Country,
			})
		} else {
			dest.DownloadAvg = internetSpeed.DownloadSpeed_Mbps
		}
	}

	if mapping.CitySafetyRef != nil {
		safetyScore, err := s.getSafetyScore(ctx, mapping.CitySafetyRef)
		if err != nil {
			logger.LogErrorWithFields("Failed to get safety score", logrus.Fields{
				"Error":   err.Error(),
				"City":    dest.City,
				"Country": dest.Country,
			})
		} else {
			dest.SafetyScore = int64(safetyScore.Score)
		}
	}

	if mapping.CostOfLivingAnalyticsRef != nil {
		colScore, err := s.getCostOfLivingScore(ctx, mapping.CostOfLivingAnalyticsRef)
		if err != nil {
			logger.LogErrorWithFields("Failed to get cost of living score", logrus.Fields{
				"Error":   err.Error(),
				"City":    dest.City,
				"Country": dest.Country,
			})
		} else {
			dest.CostOfLivingScore = colScore
		}
	}

	photos, err := s.fetchPhotosFromPexels(dest.City + " " + dest.Country)
	if err != nil {
		logger.LogErrorWithFields("Failed to fetch photos", logrus.Fields{
			"Error":   err.Error(),
			"City":    dest.City,
			"Country": dest.Country,
		})
	} else {
		// dest.PhotoUri1 = photos[0]
		// dest.PhotoUri2 = photos[1]
		dest.Photos = photos
		if len(dest.PhotoUri1) == 0 {
			dest.PhotoUri1 = photos[0]
		}
		if len(dest.PhotoUri2) == 0 {
			dest.PhotoUri2 = photos[1]
		}
	}

	return nil
}

func (s *TopDestinationEnrichmentService) getLocationMapping(ctx context.Context, city, country string) (*models.LocationMapping, error) {
	query := s.firestoreClient.Collection("location-mappings").Where("city", "==", city).Where("country", "==", country)
	docs, err := query.Documents(ctx).GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to query location mappings: %v", err)
	}

	if len(docs) == 0 {
		return nil, fmt.Errorf("no location mapping found for city %s and country %s", city, country)
	}

	var bestMapping *models.LocationMapping
	for _, doc := range docs {
		var mapping models.LocationMapping
		if err := doc.DataTo(&mapping); err != nil {
			logger.LogErrorWithFields("Failed to parse location mapping", logrus.Fields{"Error": err.Error(), "DocID": doc.Ref.ID})
			continue
		}

		if bestMapping == nil || (bestMapping.StateOrProvince == "" && mapping.StateOrProvince != "") {
			bestMapping = &mapping
		}
	}

	return bestMapping, nil
}

func (s *TopDestinationEnrichmentService) getInternetSpeed(ctx context.Context, ref *firestore.DocumentRef) (*models.InternetSpeed, error) {
	doc, err := ref.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get internet speed document: %v", err)
	}

	var internetSpeed models.InternetSpeed
	if err := doc.DataTo(&internetSpeed); err != nil {
		return nil, fmt.Errorf("failed to parse internet speed data: %v", err)
	}

	return &internetSpeed, nil
}

func (s *TopDestinationEnrichmentService) getSafetyScore(ctx context.Context, ref *firestore.DocumentRef) (*models.SafetyScore, error) {
	doc, err := ref.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get safety score document: %v", err)
	}

	var safetyScore models.SafetyScore
	if err := doc.DataTo(&safetyScore); err != nil {
		return nil, fmt.Errorf("failed to parse safety score data: %v", err)
	}

	return &safetyScore, nil
}

func (s *TopDestinationEnrichmentService) getCostOfLivingScore(ctx context.Context, ref *firestore.DocumentRef) (float64, error) {
	doc, err := ref.Get(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get cost of living analytics document: %v", err)
	}

	data := doc.Data()
	scores, ok := data["scores"].(map[string]interface{})
	if !ok {
		return 0, fmt.Errorf("invalid cost of living analytics data structure")
	}

	overallScore, ok := scores["overall"].(float64)
	if !ok {
		return 0, fmt.Errorf("invalid overall score in cost of living analytics")
	}

	return overallScore, nil
}

func (s *TopDestinationEnrichmentService) fetchPhotosFromPexels(query string) ([]string, error) {
	apiKey := s.cfg.PexelsAPIKey
	if apiKey == "" {
		return nil, fmt.Errorf("PEXELS_API_KEY not found in environment variables")
	}

	baseURL := "https://api.pexels.com/v1/search"
	params := url.Values{}
	params.Add("query", query)
	params.Add("per_page", "10")

	req, err := http.NewRequest("GET", baseURL+"?"+params.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Authorization", apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result struct {
		Photos []struct {
			Src struct {
				Medium string `json:"medium"`
			} `json:"src"`
		} `json:"photos"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	if len(result.Photos) < 2 {
		return nil, fmt.Errorf("not enough photos returned")
	}

	return []string{
		result.Photos[0].Src.Medium,
		result.Photos[1].Src.Medium,
		result.Photos[2].Src.Medium,
		result.Photos[3].Src.Medium,
		result.Photos[4].Src.Medium,
		result.Photos[5].Src.Medium,
		result.Photos[6].Src.Medium,
		result.Photos[7].Src.Medium,
	}, nil
}

func (s *TopDestinationEnrichmentService) saveDestination(ctx context.Context, dest TopDestination) error {
	_, err := s.firestoreClient.Collection("top-destinations").Doc(dest.Id).Set(ctx, dest)
	if err != nil {
		return fmt.Errorf("failed to save destination: %v", err)
	}
	return nil
}
