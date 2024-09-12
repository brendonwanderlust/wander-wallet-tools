package models

import (
	"fmt"
	"wander-wallet-tools/utils"
)

type SafetyScore struct {
	City    string `firestore:"city"`
	Country string `firestore:"country"`
	Score   int64  `firestore:"score"`
}

func GetCitySafetyPath(city, country string) string {
	collectionName := "city-safety"
	formattedCity := utils.NormalizeAndFormat(city)
	formattedCountry := utils.NormalizeAndFormat(country)
	id := fmt.Sprintf("%s-%s", formattedCity, formattedCountry)
	return fmt.Sprintf("%s/%s", collectionName, id)
}

func GetCountrySafetyPath(country string) string {
	collectionName := "country-safety"
	id := utils.NormalizeAndFormat(country)
	return fmt.Sprintf("%s/%s", collectionName, id)
}

func FindAddressComponent(components []AddressComponent, componentType string) string {
	for _, component := range components {
		if utils.Contains(component.Types, componentType) {
			return component.LongText
		}
	}
	return ""
}
