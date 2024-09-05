package services

import (
	"context"
	"fmt"
	"math"
	"reflect"
	"sort"
	"wander-wallet-tools/utils"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

type CostOfLiving struct {
	City                      string  `firestore:"city"`
	Country                   string  `firestore:"country"`
	MealInexpensiveRestaurant float64 `firestore:"mealInexpensiveRestaurant"`
	MealFor2MidRange          float64 `firestore:"mealFor2MidRange"`
	ComboMealMcdonalds        float64 `firestore:"comboMealMcdonalds"`
	DomesticBeerRestaurant    float64 `firestore:"domesticBeerRestaurant"`
	ImportedBeerRestaurant    float64 `firestore:"importedBeerRestaurant"`
	CappuccinoRestaurant      float64 `firestore:"cappuccinoRestaurant"`
	SodaRestaurant            float64 `firestore:"sodaRestaurant"`
	WaterRestaurant           float64 `firestore:"waterRestaurant"`
	WineMidRange              float64 `firestore:"wineMidRange"`
	DomesticBeerMarket        float64 `firestore:"domesticBeerMarket"`
	ImportedBeerMarket        float64 `firestore:"importedBeerMarket"`
	CigarettesPack            float64 `firestore:"cigarettesPack"`
	TicketOneWay              float64 `firestore:"ticketOneWay"`
	MonthlyPass               float64 `firestore:"monthlyPass"`
	TaxiStart                 float64 `firestore:"taxiStart"`
	Taxi1Km                   float64 `firestore:"taxi1Km"`
	Gasoline1L                float64 `firestore:"gasoline1L"`
	Utilities85sqmApartment   float64 `firestore:"utilities85sqmApartment"`
	MobileTariff1Min          float64 `firestore:"mobileTariff1Min"`
	InternetUnlimited         float64 `firestore:"internetUnlimited"`
	FitnessClubMonthly        float64 `firestore:"fitnessClubMonthly"`
	Apt1BedCityCenter         float64 `firestore:"apt1BedCityCenter"`
	Apt1BedOutsideCenter      float64 `firestore:"apt1BedOutsideCenter"`
	Apt3BedCityCenter         float64 `firestore:"apt3BedCityCenter"`
	Apt3BedOutsideCenter      float64 `firestore:"apt3BedOutsideCenter"`
	PricePerSqmCityCenter     float64 `firestore:"pricePerSqmCityCenter"`
	PricePerSqmOutsideCenter  float64 `firestore:"pricePerSqmOutsideCenter"`
	AvgNetSalary              float64 `firestore:"avgNetSalary"`
}

type MetricStats struct {
	Mean              float64 `firestore:"mean"`
	Median            float64 `firestore:"median"`
	Mode              float64 `firestore:"mode"`
	StandardDeviation float64 `firestore:"standardDeviation"`
}

type CostOfLivingAnalytics struct {
	City    string                 `firestore:"city"`
	Country string                 `firestore:"country"`
	Scores  map[string]float64     `firestore:"scores"`
	Stats   map[string]MetricStats `firestore:"stats"`
}

type CostOfLivingAnalyzerService struct {
	firestoreClient *firestore.Client
}

func NewCostOfLivingAnalyzerService(client *firestore.Client) *CostOfLivingAnalyzerService {
	return &CostOfLivingAnalyzerService{
		firestoreClient: client,
	}
}

func (s *CostOfLivingAnalyzerService) AnalyzeAndStoreData(ctx context.Context) error {
	colData, err := s.retrieveAllCostOfLivingData(ctx, s.firestoreClient)
	if err != nil {
		return fmt.Errorf("failed to retrieve data: %v", err)
	}

	relativeScores := s.analyzeData(colData)

	err = s.storeRelativeScores(ctx, s.firestoreClient, relativeScores)
	if err != nil {
		return fmt.Errorf("failed to store relative scores: %v", err)
	}

	return nil
}

func (s *CostOfLivingAnalyzerService) retrieveAllCostOfLivingData(ctx context.Context, client *firestore.Client) ([]CostOfLiving, error) {
	var colData []CostOfLiving
	iter := client.Collection("cost-of-living").Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		var col CostOfLiving
		err = doc.DataTo(&col)
		if err != nil {
			return nil, err
		}
		colData = append(colData, col)
	}
	return colData, nil
}

func (s *CostOfLivingAnalyzerService) analyzeData(colData []CostOfLiving) []CostOfLivingAnalytics {
	metricsToAnalyze := s.getPropertyNames(CostOfLiving{})
	var relativeScores []CostOfLivingAnalytics

	for _, location := range colData {
		scores := make(map[string]float64)
		stats := make(map[string]MetricStats)
		for _, metric := range metricsToAnalyze {
			value := s.getMetricValue(location, metric)
			if value > 0 {
				allValues := s.getAllMetricValues(colData, metric)
				percentile := s.calculatePercentile(allValues, value)
				scores[utils.FirstLetterToLower(metric)] = percentile
				stats[utils.FirstLetterToLower(metric)] = s.calculateStats(allValues)
			}
		}
		relativeScores = append(relativeScores, CostOfLivingAnalytics{
			City:    location.City,
			Country: location.Country,
			Scores:  scores,
			Stats:   stats,
		})
	}

	return relativeScores
}

func (s *CostOfLivingAnalyzerService) getMetricValue(col CostOfLiving, metric string) float64 {
	switch metric {
	case "MealInexpensiveRestaurant":
		return col.MealInexpensiveRestaurant
	case "MealFor2MidRange":
		return col.MealFor2MidRange
	case "ComboMealMcdonalds":
		return col.ComboMealMcdonalds
	case "DomesticBeerRestaurant":
		return col.DomesticBeerRestaurant
	case "ImportedBeerRestaurant":
		return col.ImportedBeerRestaurant
	case "CappuccinoRestaurant":
		return col.CappuccinoRestaurant
	case "SodaRestaurant":
		return col.SodaRestaurant
	case "WaterRestaurant":
		return col.WaterRestaurant
	case "WineMidRange":
		return col.WineMidRange
	case "DomesticBeerMarket":
		return col.DomesticBeerMarket
	case "ImportedBeerMarket":
		return col.ImportedBeerMarket
	case "CigarettesPack":
		return col.CigarettesPack
	case "TicketOneWay":
		return col.TicketOneWay
	case "MonthlyPass":
		return col.MonthlyPass
	case "TaxiStart":
		return col.TaxiStart
	case "Taxi1Km":
		return col.Taxi1Km
	case "Gasoline1L":
		return col.Gasoline1L
	case "Utilities85sqmApartment":
		return col.Utilities85sqmApartment
	case "MobileTariff1Min":
		return col.MobileTariff1Min
	case "InternetUnlimited":
		return col.InternetUnlimited
	case "FitnessClubMonthly":
		return col.FitnessClubMonthly
	case "Apt1BedCityCenter":
		return col.Apt1BedCityCenter
	case "Apt1BedOutsideCenter":
		return col.Apt1BedOutsideCenter
	case "Apt3BedCityCenter":
		return col.Apt3BedCityCenter
	case "Apt3BedOutsideCenter":
		return col.Apt3BedOutsideCenter
	case "PricePerSqmCityCenter":
		return col.PricePerSqmCityCenter
	case "PricePerSqmOutsideCenter":
		return col.PricePerSqmOutsideCenter
	case "AvgNetSalary":
		return col.AvgNetSalary
	default:
		return 0
	}
}

func (s *CostOfLivingAnalyzerService) getAllMetricValues(colData []CostOfLiving, metric string) []float64 {
	var values []float64
	for _, col := range colData {
		value := s.getMetricValue(col, metric)
		if value > 0 {
			values = append(values, value)
		}
	}
	return values
}

func (s *CostOfLivingAnalyzerService) calculatePercentile(values []float64, value float64) float64 {
	sort.Float64s(values)
	index := sort.SearchFloat64s(values, value)
	return float64(index) / float64(len(values)) * 100
}

func (s *CostOfLivingAnalyzerService) calculateStats(values []float64) MetricStats {
	mean := s.calculateMean(values)
	return MetricStats{
		Mean:              mean,
		Median:            s.calculateMedian(values),
		Mode:              s.calculateMode(values),
		StandardDeviation: s.calculateStandardDeviation(values, mean),
	}
}

func (s *CostOfLivingAnalyzerService) calculateMean(values []float64) float64 {
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func (s *CostOfLivingAnalyzerService) calculateMedian(values []float64) float64 {
	sort.Float64s(values)
	mid := len(values) / 2
	if len(values)%2 == 0 {
		return (values[mid-1] + values[mid]) / 2
	}
	return values[mid]
}

func (s *CostOfLivingAnalyzerService) calculateMode(values []float64) float64 {
	frequencyMap := make(map[float64]int)
	for _, v := range values {
		frequencyMap[v]++
	}
	mode := 0.0
	maxFrequency := 0
	for value, frequency := range frequencyMap {
		if frequency > maxFrequency {
			mode = value
			maxFrequency = frequency
		}
	}
	return mode
}

func (s *CostOfLivingAnalyzerService) calculateStandardDeviation(values []float64, mean float64) float64 {
	variance := 0.0
	for _, v := range values {
		variance += math.Pow(v-mean, 2)
	}
	variance /= float64(len(values))
	return math.Sqrt(variance)
}

func (s *CostOfLivingAnalyzerService) storeRelativeScores(ctx context.Context, client *firestore.Client, relativeScores []CostOfLivingAnalytics) error {
	for _, rs := range relativeScores {
		_, err := client.Collection("cost-of-living-analytics").Doc(fmt.Sprintf("%s-%s", utils.NormalizeAndFormat(rs.City), utils.NormalizeAndFormat(rs.Country))).Set(ctx, rs)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *CostOfLivingAnalyzerService) getPropertyNames(structType interface{}) []string {
	v := reflect.TypeOf(structType)
	var propertyNames []string

	for i := 0; i < v.NumField(); i++ {
		propertyNames = append(propertyNames, v.Field(i).Name)
	}

	return propertyNames
}
