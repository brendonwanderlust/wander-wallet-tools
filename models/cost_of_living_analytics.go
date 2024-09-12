package models

import (
	"fmt"
	"wander-wallet-tools/utils"
)

type MetricStats struct {
	Min   float64 `firestore:"min"`
	Max   float64 `firestore:"max"`
	Avg   float64 `firestore:"avg"`
	Count int64   `firestore:"count"`
}

type CostOfLivingAnalytics struct {
	City    string `firestore:"city"`
	Country string `firestore:"country"`
	Scores  Scores `firestore:"scores"`
	Stats   Stats  `firestore:"stats"`
}

type Scores struct {
	Overall                   float64 `firestore:"overall"`
	Apt1BedCityCenter         float64 `firestore:"apt1BedCityCenter"`
	Apt1BedOutsideCenter      float64 `firestore:"apt1BedOutsideCenter"`
	Apt3BedCityCenter         float64 `firestore:"apt3BedCityCenter"`
	Apt3BedOutsideCenter      float64 `firestore:"apt3BedOutsideCenter"`
	AvgNetSalary              float64 `firestore:"avgNetSalary"`
	CappuccinoRestaurant      float64 `firestore:"cappuccinoRestaurant"`
	ComboMealMcdonalds        float64 `firestore:"comboMealMcdonalds"`
	DomesticBeerMarket        float64 `firestore:"domesticBeerMarket"`
	DomesticBeerRestaurant    float64 `firestore:"domesticBeerRestaurant"`
	FitnessClubMonthly        float64 `firestore:"fitnessClubMonthly"`
	Gasoline1L                float64 `firestore:"gasoline1L"`
	ImportedBeerMarket        float64 `firestore:"importedBeerMarket"`
	ImportedBeerRestaurant    float64 `firestore:"importedBeerRestaurant"`
	InternetUnlimited         float64 `firestore:"internetUnlimited"`
	MealInexpensiveRestaurant float64 `firestore:"mealInexpensiveRestaurant"`
	MealFor2MidRange          float64 `firestore:"mealFor2MidRange"`
	MobileTariff1Min          float64 `firestore:"mobileTariff1Min"`
	MonthlyPass               float64 `firestore:"monthlyPass"`
	PricePerSqmCityCenter     float64 `firestore:"pricePerSqmCityCenter"`
	PricePerSqmOutsideCenter  float64 `firestore:"pricePerSqmOutsideCenter"`
	SodaRestaurant            float64 `firestore:"sodaRestaurant"`
	TaxiStart                 float64 `firestore:"taxiStart"`
	Taxi1Km                   float64 `firestore:"taxi1Km"`
	TicketOneWay              float64 `firestore:"ticketOneWay"`
	Utilities85sqmApartment   float64 `firestore:"utilities85sqmApartment"`
	WaterRestaurant           float64 `firestore:"waterRestaurant"`

	// Unused fields
	Milk1L                  float64 `firestore:"milk1L"`
	BreadLoaf               float64 `firestore:"breadLoaf"`
	Rice1Kg                 float64 `firestore:"rice1Kg"`
	Eggs12Pack              float64 `firestore:"eggs12Pack"`
	LocalCheese1Kg          float64 `firestore:"localCheese1Kg"`
	ChickenFillet1Kg        float64 `firestore:"chickenFillet1Kg"`
	BeefRound1Kg            float64 `firestore:"beefRound1Kg"`
	Apples1Kg               float64 `firestore:"apples1Kg"`
	Banana1Kg               float64 `firestore:"banana1Kg"`
	Oranges1Kg              float64 `firestore:"oranges1Kg"`
	Tomato1Kg               float64 `firestore:"tomato1Kg"`
	Potato1Kg               float64 `firestore:"potato1Kg"`
	Onion1Kg                float64 `firestore:"onion1Kg"`
	LettuceHead             float64 `firestore:"lettuceHead"`
	Water1_5LMarket         float64 `firestore:"water1_5LMarket"`
	WineMidRange            float64 `firestore:"wineMidRange"`
	CigarettesPack          float64 `firestore:"cigarettesPack"`
	TaxiWaiting1Hour        float64 `firestore:"taxiWaiting1Hour"`
	VWGolfNew               float64 `firestore:"vwGolfNew"`
	ToyotaCorollaNew        float64 `firestore:"toyotaCorollaNew"`
	TennisCourtHourly       float64 `firestore:"tennisCourtHourly"`
	CinemaTicket            float64 `firestore:"cinemaTicket"`
	PreschoolMonthly        float64 `firestore:"preschoolMonthly"`
	IntlPrimarySchoolYearly float64 `firestore:"intlPrimarySchoolYearly"`
	Jeans                   float64 `firestore:"jeans"`
	SummerDress             float64 `firestore:"summerDress"`
	NikeShoes               float64 `firestore:"nikeShoes"`
	LeatherShoes            float64 `firestore:"leatherShoes"`
	MortgageRate            float64 `firestore:"mortgageRate"`
	DataQuality             float64 `firestore:"dataQuality"`
}

type Stats struct {
	Apt1BedCityCenter         MetricStats `firestore:"apt1BedCityCenter"`
	Apt1BedOutsideCenter      MetricStats `firestore:"apt1BedOutsideCenter"`
	Apt3BedCityCenter         MetricStats `firestore:"apt3BedCityCenter"`
	Apt3BedOutsideCenter      MetricStats `firestore:"apt3BedOutsideCenter"`
	AvgNetSalary              MetricStats `firestore:"avgNetSalary"`
	CappuccinoRestaurant      MetricStats `firestore:"cappuccinoRestaurant"`
	ComboMealMcdonalds        MetricStats `firestore:"comboMealMcdonalds"`
	DomesticBeerMarket        MetricStats `firestore:"domesticBeerMarket"`
	DomesticBeerRestaurant    MetricStats `firestore:"domesticBeerRestaurant"`
	FitnessClubMonthly        MetricStats `firestore:"fitnessClubMonthly"`
	Gasoline1L                MetricStats `firestore:"gasoline1L"`
	ImportedBeerMarket        MetricStats `firestore:"importedBeerMarket"`
	ImportedBeerRestaurant    MetricStats `firestore:"importedBeerRestaurant"`
	InternetUnlimited         MetricStats `firestore:"internetUnlimited"`
	MealInexpensiveRestaurant MetricStats `firestore:"mealInexpensiveRestaurant"`
	MealFor2MidRange          MetricStats `firestore:"mealFor2MidRange"`
	MobileTariff1Min          MetricStats `firestore:"mobileTariff1Min"`
	MonthlyPass               MetricStats `firestore:"monthlyPass"`
	PricePerSqmCityCenter     MetricStats `firestore:"pricePerSqmCityCenter"`
	PricePerSqmOutsideCenter  MetricStats `firestore:"pricePerSqmOutsideCenter"`
	SodaRestaurant            MetricStats `firestore:"sodaRestaurant"`
	TaxiStart                 MetricStats `firestore:"taxiStart"`
	Taxi1Km                   MetricStats `firestore:"taxi1Km"`
	TicketOneWay              MetricStats `firestore:"ticketOneWay"`
	Utilities85sqmApartment   MetricStats `firestore:"utilities85sqmApartment"`
	WaterRestaurant           MetricStats `firestore:"waterRestaurant"`

	// Unused fields
	Milk1L                  MetricStats `firestore:"milk1L"`
	BreadLoaf               MetricStats `firestore:"breadLoaf"`
	Rice1Kg                 MetricStats `firestore:"rice1Kg"`
	Eggs12Pack              MetricStats `firestore:"eggs12Pack"`
	LocalCheese1Kg          MetricStats `firestore:"localCheese1Kg"`
	ChickenFillet1Kg        MetricStats `firestore:"chickenFillet1Kg"`
	BeefRound1Kg            MetricStats `firestore:"beefRound1Kg"`
	Apples1Kg               MetricStats `firestore:"apples1Kg"`
	Banana1Kg               MetricStats `firestore:"banana1Kg"`
	Oranges1Kg              MetricStats `firestore:"oranges1Kg"`
	Tomato1Kg               MetricStats `firestore:"tomato1Kg"`
	Potato1Kg               MetricStats `firestore:"potato1Kg"`
	Onion1Kg                MetricStats `firestore:"onion1Kg"`
	LettuceHead             MetricStats `firestore:"lettuceHead"`
	Water1_5LMarket         MetricStats `firestore:"water1_5LMarket"`
	WineMidRange            MetricStats `firestore:"wineMidRange"`
	CigarettesPack          MetricStats `firestore:"cigarettesPack"`
	TaxiWaiting1Hour        MetricStats `firestore:"taxiWaiting1Hour"`
	VWGolfNew               MetricStats `firestore:"vwGolfNew"`
	ToyotaCorollaNew        MetricStats `firestore:"toyotaCorollaNew"`
	TennisCourtHourly       MetricStats `firestore:"tennisCourtHourly"`
	CinemaTicket            MetricStats `firestore:"cinemaTicket"`
	PreschoolMonthly        MetricStats `firestore:"preschoolMonthly"`
	IntlPrimarySchoolYearly MetricStats `firestore:"intlPrimarySchoolYearly"`
	Jeans                   MetricStats `firestore:"jeans"`
	SummerDress             MetricStats `firestore:"summerDress"`
	NikeShoes               MetricStats `firestore:"nikeShoes"`
	LeatherShoes            MetricStats `firestore:"leatherShoes"`
	MortgageRate            MetricStats `firestore:"mortgageRate"`
	DataQuality             MetricStats `firestore:"dataQuality"`
}

func GetCostOfLivingAnalyticsPath(city, country string) string {
	collectionName := "cost-of-living-analytics"
	formattedCity := utils.NormalizeAndFormat(city)
	formattedCountry := utils.NormalizeAndFormat(country)
	id := fmt.Sprintf("%s-%s", formattedCity, formattedCountry)
	return fmt.Sprintf("%s/%s", collectionName, id)
}
