package services

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
	"wander-wallet-tools/config"
	"wander-wallet-tools/logger"
	"wander-wallet-tools/models"
	"wander-wallet-tools/utils"

	"cloud.google.com/go/bigquery"
	"cloud.google.com/go/firestore"
	"github.com/sirupsen/logrus"
	"google.golang.org/api/iterator"
	"googlemaps.github.io/maps"
)

type MissingValueReport struct {
	City           string
	Country        string
	MissingPlaceID bool
	MissingSpeed   bool
	MissingSafety  bool
	MissingCOL     bool
	MissingPhotos  bool
}

type TopDestinationEnrichmentService struct {
	firestoreClient *firestore.Client
	cfg             *config.Config
	mapsClient      *maps.Client
	bigqueryClient  *bigquery.Client
}

func NewTopDestinationEnrichmentService(bigqueryClient *bigquery.Client, client *firestore.Client, cfg *config.Config, mapsClient *maps.Client) *TopDestinationEnrichmentService {
	return &TopDestinationEnrichmentService{
		bigqueryClient:  bigqueryClient,
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

	missingValueReports := []MissingValueReport{}

	for _, dest := range destinations {
		report, err := s.enrichDestination(ctx, &dest)
		if err != nil {
			logger.LogErrorWithFields("Failed to enrich destination", logrus.Fields{
				"Error":   err.Error(),
				"City":    dest.City,
				"Country": dest.Country,
			})
			continue
		}

		missingValueReports = append(missingValueReports, report)

		err = s.saveDestination(ctx, dest)
		if err != nil {
			logger.LogErrorWithFields("Failed to save enriched destination", logrus.Fields{
				"Error":   err.Error(),
				"City":    dest.City,
				"Country": dest.Country,
			})
		}
	}

	err = s.generateMissingValuesCSV(missingValueReports)
	if err != nil {
		logger.LogErrorWithFields("Failed to generate CSV report", logrus.Fields{"Error": err.Error()})
		return err
	}

	return nil
}

func (s *TopDestinationEnrichmentService) getTopDestinations(ctx context.Context) ([]models.TopDestination, error) {
	var destinations []models.TopDestination
	docs, err := s.firestoreClient.Collection("top-destinations").OrderBy("rank", firestore.Asc).Offset(60).Limit(100).Documents(ctx).GetAll()
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

func (s *TopDestinationEnrichmentService) enrichDestination(ctx context.Context, dest *models.TopDestination) (MissingValueReport, error) {
	report := MissingValueReport{
		City:    dest.City,
		Country: dest.Country,
	}

	mapping, err := s.getLocationMapping(ctx, *dest)
	if err != nil {
		return report, fmt.Errorf("failed to get location mapping: %v", err)
	}

	if mapping.PlaceId == "" {
		report.MissingPlaceID = true
	}
	dest.PlaceId = mapping.PlaceId

	if mapping.InternetSpeedRef != nil {
		internetSpeed, err := s.getInternetSpeed(ctx, mapping)
		if err != nil {
			logger.LogErrorWithFields("Failed to get internet speed", logrus.Fields{
				"Error":   err.Error(),
				"City":    dest.City,
				"Country": dest.Country,
			})
			report.MissingSpeed = true
		} else {
			dest.DownloadAvg = internetSpeed.DownloadSpeed_Mbps
		}
	} else {
		report.MissingSpeed = true
	}

	if mapping.CitySafetyRef != nil {
		safetyScore, err := s.getSafetyScore(ctx, mapping.CitySafetyRef)
		if err != nil {
			logger.LogErrorWithFields("Failed to get safety score", logrus.Fields{
				"Error":   err.Error(),
				"City":    dest.City,
				"Country": dest.Country,
			})
			report.MissingSafety = true
		} else {
			dest.SafetyScore = int64(safetyScore.Score)
		}
	} else {
		report.MissingSafety = true
	}

	if mapping.CostOfLivingAnalyticsRef != nil {
		colScore, err := s.getCostOfLivingScore(ctx, mapping.CostOfLivingAnalyticsRef)
		if err != nil {
			logger.LogErrorWithFields("Failed to get cost of living score", logrus.Fields{
				"Error":   err.Error(),
				"City":    dest.City,
				"Country": dest.Country,
			})
			report.MissingCOL = true
		} else {
			dest.CostOfLivingScore = colScore
		}
	} else {
		report.MissingCOL = true
	}

	if len(dest.Photos) == 0 {
		photos, err := s.fetchPhotosFromPexels(dest.City + " " + dest.Country)
		if err != nil {
			logger.LogErrorWithFields("Failed to fetch photos", logrus.Fields{
				"Error":   err.Error(),
				"City":    dest.City,
				"Country": dest.Country,
			})
			report.MissingPhotos = true
		} else {
			dest.Photos = photos
			if len(dest.PhotoUri1) == 0 {
				dest.PhotoUri1 = photos[0]
			}
			if len(dest.PhotoUri2) == 0 {
				dest.PhotoUri2 = photos[1]
			}
		}
	}

	return report, nil
}

func (s *TopDestinationEnrichmentService) getLocationMapping(ctx context.Context, dest models.TopDestination) (*models.LocationMapping, error) {
	docId := fmt.Sprintf("%v-%s", utils.NormalizeAndFormat(dest.City), utils.NormalizeAndFormat(dest.Country))
	docById, err := s.firestoreClient.Collection("location-mappings").Doc(docId).Get(ctx)
	if err != nil {
		logger.LogErrorLn("failed to get location mapping by DocId: %v", err)
	}

	if docById.Exists() {
		var mapping *models.LocationMapping
		if err := docById.DataTo(&mapping); err != nil {
			return nil, fmt.Errorf("failed to parse location mapping", err)
		}
		mapping = s.enrichMapping(ctx, *mapping)
		return mapping, nil
	}

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

		if bestMapping == nil || mapping.StateOrProvince != "" {
			bestMapping = &mapping
		}
	}

	bestMapping = s.enrichMapping(ctx, *bestMapping)
	return bestMapping, nil
}

func (s *TopDestinationEnrichmentService) enrichMapping(ctx context.Context, mapping models.LocationMapping) *models.LocationMapping {
	req := &maps.FindPlaceFromTextRequest{
		Input:     fmt.Sprintf("%s, %s", mapping.City, mapping.Country),
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
		return &mapping
	}

	if len(resp.Candidates) == 0 {
		return &mapping
	}

	candidate := resp.Candidates[0]

	var stateOrProvince = ""
	if containsAdministrativeArea(candidate.Types) {
		stateOrProvince = extractStateOrProvince(candidate.FormattedAddress)
	}
	standardName := models.ConstructStandardName("", mapping.City, stateOrProvince, mapping.Country)
	mapping.Id = standardName
	mapping.Latitude = candidate.Geometry.Location.Lat
	mapping.Longitude = candidate.Geometry.Location.Lng
	mapping.PlaceId = candidate.PlaceID
	mapping.Types = candidate.Types
	mapping.Aliases = models.UniqueNonEmptyStrings(candidate.Name, candidate.FormattedAddress, fmt.Sprintf("%s, %s", mapping.City, mapping.Country))
	mapping.LastRequested = time.Now()

	mapping = *s.createDocRefs(&mapping)
	s.firestoreClient.Collection("location-mappings").Doc(mapping.Id).Set(ctx, mapping)
	if err != nil {
		return &mapping
	}
	return &mapping
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

func (s *TopDestinationEnrichmentService) getInternetSpeed(ctx context.Context, mapping *models.LocationMapping) (*models.InternetSpeed, error) {
	doc, err := mapping.InternetSpeedRef.Get(ctx)
	if err != nil {
		logger.LogErrorLn("failed to get internet speed document: %v", err)
	}

	if doc.Exists() {
		var internetSpeed models.InternetSpeed
		if err := doc.DataTo(&internetSpeed); err != nil {
			return nil, fmt.Errorf("failed to parse internet speed data: %v", err)
		}
		return &internetSpeed, nil
	}

	newSpeed, err := s.createAndSaveInternetSpeeds(ctx, mapping)
	if err != nil {
		return nil, fmt.Errorf("failed to create internet speed data: %v", err)
	}

	return newSpeed, nil
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
	params.Add("per_page", "15")

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
		result.Photos[10].Src.Medium,
		result.Photos[11].Src.Medium,
		result.Photos[12].Src.Medium,
		result.Photos[13].Src.Medium,
		result.Photos[14].Src.Medium,
	}, nil
}

func (s *TopDestinationEnrichmentService) saveDestination(ctx context.Context, dest models.TopDestination) error {
	_, err := s.firestoreClient.Collection("top-destinations").Doc(dest.Id).Set(ctx, dest)
	if err != nil {
		return fmt.Errorf("failed to save destination: %v", err)
	}
	return nil
}

func (s *TopDestinationEnrichmentService) generateMissingValuesCSV(reports []MissingValueReport) error {
	file, err := os.Create("missing_values_report.csv")
	if err != nil {
		return fmt.Errorf("failed to create CSV file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	headers := []string{"City", "Country", "Missing PlaceID", "Missing Internet Speed", "Missing Safety Score", "Missing Cost of Living", "Missing Photos"}
	if err := writer.Write(headers); err != nil {
		return fmt.Errorf("error writing CSV headers: %v", err)
	}

	for _, report := range reports {
		row := []string{
			report.City,
			report.Country,
			fmt.Sprintf("%t", report.MissingPlaceID),
			fmt.Sprintf("%t", report.MissingSpeed),
			fmt.Sprintf("%t", report.MissingSafety),
			fmt.Sprintf("%t", report.MissingCOL),
			fmt.Sprintf("%t", report.MissingPhotos),
		}
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("error writing CSV row: %v", err)
		}
	}

	return nil
}

func (h *TopDestinationEnrichmentService) createAndSaveInternetSpeeds(ctx context.Context, mapping *models.LocationMapping) (*models.InternetSpeed, error) {
	type result struct {
		speed float64
		err   error
	}

	downloadCh := make(chan result)
	uploadCh := make(chan result)

	go func() {
		speed, err := h.getSpeedAvg(ctx, mapping.FormattedAddress, mapping.Latitude, mapping.Longitude, mapping.Types, true)
		downloadCh <- result{speed, err}
	}()

	go func() {
		speed, err := h.getSpeedAvg(ctx, mapping.FormattedAddress, mapping.Latitude, mapping.Longitude, mapping.Types, false)
		uploadCh <- result{speed, err}
	}()

	downloadResult := <-downloadCh
	uploadResult := <-uploadCh

	if downloadResult.err != nil {
		return nil, fmt.Errorf("failed to get download speed: %v", downloadResult.err)
	}

	if uploadResult.err != nil {
		return nil, fmt.Errorf("failed to get upload speed: %v", uploadResult.err)
	}

	var internetSpeed models.InternetSpeed
	internetSpeed.DownloadSpeed_Mbps = downloadResult.speed
	internetSpeed.UploadSpeed_Mbps = uploadResult.speed
	internetSpeed.Latitude = mapping.Latitude
	internetSpeed.Longitude = mapping.Longitude
	internetSpeed.LocationName = mapping.FormattedAddress
	internetSpeed.Types = mapping.Types

	var stateOrProvince = ""
	if containsAdministrativeArea(mapping.Types) {
		stateOrProvince = extractStateOrProvince(mapping.FormattedAddress)
	}
	standardName := models.ConstructStandardName("", mapping.City, stateOrProvince, mapping.Country)
	_, err := h.firestoreClient.Collection("internet-speed-cache").Doc(standardName).Set(ctx, mapping)
	if err != nil {
		return nil, fmt.Errorf("failed to save internet score: %v", err)
	}

	return &internetSpeed, nil
}

func (h *TopDestinationEnrichmentService) getSpeedAvg(ctx context.Context, formattedAddress string, lat, lng float64, types []string, isDownload bool) (float64, error) {
	boundaries := getBoundingLatLng(lat, lng)
	query := buildQuery(formattedAddress, boundaries, types, isDownload)

	q := h.bigqueryClient.Query(query)
	job, err := q.Run(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to execute BigQuery: %v", err)
	}
	status, err := job.Wait(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to execute BigQuery: %v", err)
	}
	if err := status.Err(); err != nil {
		return 0, fmt.Errorf("failed to execute BigQuery: %v", err)
	}
	it, err := job.Read(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to execute BigQuery: %v", err)
	}

	var row struct{ Avg float64 }
	err = it.Next(&row)
	if err == iterator.Done {
		return 0, fmt.Errorf("no results found")
	}
	if err != nil {
		return 0, fmt.Errorf("failed to get BigQuery result: %v", err)
	}

	return row.Avg, nil
}

func getBoundingLatLng(latitude, longitude float64) map[string]float64 {
	offset := 0.045
	latMax := latitude + offset
	latMin := latitude - offset

	lngOffset := offset * math.Cos(latitude*math.Pi/180.0)
	lngMax := longitude + lngOffset
	lngMin := longitude - lngOffset

	return map[string]float64{
		"maxLat": latMax,
		"maxLng": lngMax,
		"minLat": latMin,
		"minLng": lngMin,
	}
}

func buildQuery(formattedAddress string, boundaries map[string]float64, types []string, isDownload bool) string {
	table := "unified_downloads"
	if !isDownload {
		table = "unified_uploads"
	}

	if utils.Contains(types, "country") {
		return fmt.Sprintf(`
			SELECT AVG(a.MeanThroughputMbps) as avg
			FROM `+"`measurement-lab.ndt.%s`"+`
			WHERE LOWER(client.Geo.CountryName) = '%s'
			AND date > (CURRENT_DATE - 2) LIMIT 3000`,
			table, strings.ToLower(formattedAddress))
	} else {
		return fmt.Sprintf(`
			SELECT AVG(a.MeanThroughputMbps) as avg
			FROM `+"`measurement-lab.ndt.%s`"+`
			WHERE (client.Geo.Latitude >= %f AND client.Geo.Latitude <= %f)
			AND (client.Geo.Longitude >= %f AND client.Geo.Longitude <= %f)
			AND date > (CURRENT_DATE - 7) LIMIT 3000`,
			table, boundaries["minLat"], boundaries["maxLat"], boundaries["minLng"], boundaries["maxLng"])
	}
}
