package models

import (
	"fmt"
	"wander-wallet-tools/utils"
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
	Milk1L                    float64 `firestore:"milk1L"`
	BreadLoaf                 float64 `firestore:"breadLoaf"`
	Rice1Kg                   float64 `firestore:"rice1Kg"`
	Eggs12Pack                float64 `firestore:"eggs12Pack"`
	LocalCheese1Kg            float64 `firestore:"localCheese1Kg"`
	ChickenFillet1Kg          float64 `firestore:"chickenFillet1Kg"`
	BeefRound1Kg              float64 `firestore:"beefRound1Kg"`
	Apples1Kg                 float64 `firestore:"apples1Kg"`
	Banana1Kg                 float64 `firestore:"banana1Kg"`
	Oranges1Kg                float64 `firestore:"oranges1Kg"`
	Tomato1Kg                 float64 `firestore:"tomato1Kg"`
	Potato1Kg                 float64 `firestore:"potato1Kg"`
	Onion1Kg                  float64 `firestore:"onion1Kg"`
	LettuceHead               float64 `firestore:"lettuceHead"`
	Water1_5LMarket           float64 `firestore:"water1_5LMarket"`
	WineMidRange              float64 `firestore:"wineMidRange"`
	DomesticBeerMarket        float64 `firestore:"domesticBeerMarket"`
	ImportedBeerMarket        float64 `firestore:"importedBeerMarket"`
	CigarettesPack            float64 `firestore:"cigarettesPack"`
	TicketOneWay              float64 `firestore:"ticketOneWay"`
	MonthlyPass               float64 `firestore:"monthlyPass"`
	TaxiStart                 float64 `firestore:"taxiStart"`
	Taxi1Km                   float64 `firestore:"taxi1Km"`
	TaxiWaiting1Hour          float64 `firestore:"taxiWaiting1Hour"`
	Gasoline1L                float64 `firestore:"gasoline1L"`
	VwGolfNew                 float64 `firestore:"vwGolfNew"`
	ToyotaCorollaNew          float64 `firestore:"toyotaCorollaNew"`
	Utilities85sqmApartment   float64 `firestore:"utilities85sqmApartment"`
	MobileTariff1Min          float64 `firestore:"mobileTariff1Min"`
	InternetUnlimited         float64 `firestore:"internetUnlimited"`
	FitnessClubMonthly        float64 `firestore:"fitnessClubMonthly"`
	TennisCourtHourly         float64 `firestore:"tennisCourtHourly"`
	CinemaTicket              float64 `firestore:"cinemaTicket"`
	PreschoolMonthly          float64 `firestore:"preschoolMonthly"`
	IntlPrimarySchoolYearly   float64 `firestore:"intlPrimarySchoolYearly"`
	Jeans                     float64 `firestore:"jeans"`
	SummerDress               float64 `firestore:"summerDress"`
	NikeShoes                 float64 `firestore:"nikeShoes"`
	LeatherShoes              float64 `firestore:"leatherShoes"`
	Apt1BedCityCenter         float64 `firestore:"apt1BedCityCenter"`
	Apt1BedOutsideCenter      float64 `firestore:"apt1BedOutsideCenter"`
	Apt3BedCityCenter         float64 `firestore:"apt3BedCityCenter"`
	Apt3BedOutsideCenter      float64 `firestore:"apt3BedOutsideCenter"`
	PricePerSqmCityCenter     float64 `firestore:"pricePerSqmCityCenter"`
	PricePerSqmOutsideCenter  float64 `firestore:"pricePerSqmOutsideCenter"`
	AvgNetSalary              float64 `firestore:"avgNetSalary"`
	MortgageRate              float64 `firestore:"mortgageRate"`
	DataQuality               int     `firestore:"dataQuality"`
}

func GetCostOfLivingPath(city, country string) string {
	collectionName := "cost-of-living"
	formattedCity := utils.NormalizeAndFormat(city)
	formattedCountry := utils.NormalizeAndFormat(country)
	id := fmt.Sprintf("%s-%s", formattedCity, formattedCountry)
	return fmt.Sprintf("%s/%s", collectionName, id)
}
