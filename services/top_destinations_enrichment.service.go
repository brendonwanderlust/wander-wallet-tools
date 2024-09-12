package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
	"wander-wallet-tools/config"
	"wander-wallet-tools/logger"
	"wander-wallet-tools/models"

	"cloud.google.com/go/firestore"
	"github.com/sirupsen/logrus"
	"googlemaps.github.io/maps"
)

type TopDestinationEnrichmentService struct {
	firestoreClient *firestore.Client
	cfg             *config.Config
	mapsClient      *maps.Client
}

func NewTopDestinationEnrichmentService(client *firestore.Client, cfg *config.Config, mapsClient *maps.Client) *TopDestinationEnrichmentService {
	return &TopDestinationEnrichmentService{
		firestoreClient: client,
		cfg:             cfg,
		mapsClient:      mapsClient,
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

func (s *TopDestinationEnrichmentService) getTopDestinations(ctx context.Context) ([]models.TopDestination, error) {
	var destinations []models.TopDestination
	docs, err := s.firestoreClient.Collection("top-destinations").OrderBy("rank", firestore.Asc).Offset(60).Limit(90).Documents(ctx).GetAll()
	if err != nil {
		return nil, err
	}

	for _, doc := range docs {
		var dest models.TopDestination
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

func (s *TopDestinationEnrichmentService) enrichDestination(ctx context.Context, dest *models.TopDestination) error {
	mapping, err := s.getLocationMapping(ctx, *dest)
	if err != nil {
		return fmt.Errorf("failed to get location mapping: %v", err)
	}

	dest.PlaceId = mapping.PlaceId

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

func (s *TopDestinationEnrichmentService) getLocationMapping(ctx context.Context, dest models.TopDestination) (*models.LocationMapping, error) {
	query := s.firestoreClient.Collection("location-mappings").Where("city", "==", dest.City).Where("country", "==", dest.Country)
	docs, err := query.Documents(ctx).GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to query location mappings: %v", err)
	}

	if len(docs) == 0 {
		mapping, err := s.createLocationMapping(ctx, dest)
		if err != nil {
			return nil, fmt.Errorf("no location mapping locatoin mapping was able to be created", err)
		}

		return mapping, nil
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

func (s *TopDestinationEnrichmentService) createLocationMapping(ctx context.Context, topDest models.TopDestination) (*models.LocationMapping, error) {
	req := &maps.FindPlaceFromTextRequest{
		Input:     fmt.Sprintf("%s, %s", topDest.City, topDest.Country),
		InputType: maps.FindPlaceFromTextInputTypeTextQuery,
		Fields: []maps.PlaceSearchFieldMask{
			maps.PlaceSearchFieldMaskPlaceID,
			maps.PlaceSearchFieldMaskName,
			maps.PlaceSearchFieldMaskFormattedAddress,
			maps.PlaceSearchFieldMaskGeometryLocationLat,
			maps.PlaceSearchFieldMaskGeometryLocationLng,
			maps.PlaceSearchFieldMaskTypes,
			maps.PlaceSearchFieldMaskGeometry,
		},
	}

	resp, err := s.mapsClient.FindPlaceFromText(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("error finding place: %v", err)
	}

	if len(resp.Candidates) == 0 {
		return nil, fmt.Errorf("no places found for %s, %s", topDest.City, topDest.Country)
	}

	candidate := resp.Candidates[0]

	var stateOrProvince = ""
	if containsAdministrativeArea(candidate.Types) {
		stateOrProvince = extractStateOrProvince(candidate.FormattedAddress)
	}
	standardName := models.ConstructStandardName("", topDest.City, stateOrProvince, topDest.Country)
	mapping := &models.LocationMapping{
		Id:               standardName,
		StandardName:     standardName,
		DisplayName:      candidate.Name,
		FormattedAddress: candidate.FormattedAddress,
		City:             topDest.City,
		Country:          topDest.Country,
		Latitude:         candidate.Geometry.Location.Lat,
		Longitude:        candidate.Geometry.Location.Lng,
		PlaceId:          candidate.PlaceID,
		Types:            candidate.Types,
		Aliases:          models.UniqueNonEmptyStrings(candidate.Name, candidate.FormattedAddress),
		LastRequested:    time.Now(),
	}
	mapping = s.createDocRefs(mapping)
	result, err := s.firestoreClient.Collection("location-mappings").Doc(mapping.Id).Set(ctx, mapping)
	if err != nil {
		return nil, fmt.Errorf("failed to get internet speed document: %v", err)
	}
	logger.LogInfoLn(result.UpdateTime.GoString())
	return mapping, nil
}

func (s *TopDestinationEnrichmentService) createDocRefs(mapping *models.LocationMapping) *models.LocationMapping {
	if mapping.City != "" {
		citySafetyPath := models.GetCitySafetyPath(mapping.City, mapping.Country)
		citySafetyRef := s.firestoreClient.Doc(citySafetyPath)
		mapping.CitySafetyRef = citySafetyRef
	}

	countrySafetyPath := models.GetCountrySafetyPath(mapping.Country)
	countrySafetyRef := s.firestoreClient.Doc(countrySafetyPath)
	mapping.CountrySafetyRef = countrySafetyRef

	internetSpeedPath := models.GetInternetSpeedPathFromLocationMapping(*mapping)
	internetSpeedRef := s.firestoreClient.Doc(internetSpeedPath)
	mapping.InternetSpeedRef = internetSpeedRef

	costOfLivingPath := models.GetCostOfLivingPath(mapping.City, mapping.Country)
	costOfLivingRef := s.firestoreClient.Doc(costOfLivingPath)
	mapping.CostOfLivingRef = costOfLivingRef

	costOfLivingAnalyticsPath := models.GetCostOfLivingAnalyticsPath(mapping.City, mapping.Country)
	costOfLivingAnalyticsRef := s.firestoreClient.Doc(costOfLivingAnalyticsPath)
	mapping.CostOfLivingAnalyticsRef = costOfLivingAnalyticsRef
	return mapping
}

func containsAdministrativeArea(types []string) bool {
	for _, t := range types {
		if t == "administrative_area_level_1" {
			return true
		}
	}
	return false
}

func extractStateOrProvince(formattedAddress string) string {
	parts := strings.Split(formattedAddress, ",")
	if len(parts) >= 3 {
		return strings.TrimSpace(parts[len(parts)-2])
	}
	return ""
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
		result.Photos[8].Src.Medium,
		result.Photos[9].Src.Medium,
	}, nil
}

func (s *TopDestinationEnrichmentService) saveDestination(ctx context.Context, dest models.TopDestination) error {
	_, err := s.firestoreClient.Collection("top-destinations").Doc(dest.Id).Set(ctx, dest)
	if err != nil {
		return fmt.Errorf("failed to save destination: %v", err)
	}
	return nil
}
