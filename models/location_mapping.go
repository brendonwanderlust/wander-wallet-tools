package models

import (
	"fmt"
	"strings"
	"time"
	"unicode"
	"wander-wallet-tools/logger"
	"wander-wallet-tools/utils"

	"cloud.google.com/go/firestore"
)

type AddressComponent struct {
	ShortText string   `firestore:"shortText"`
	LongText  string   `firestore:"longText"`
	Types     []string `firestore:"types"`
}

type LocationMapping struct {
	Id                string             `firestore:"id"`
	StandardName      string             `firestore:"standardName"`
	PlaceId           string             `firestore:"placeId"`
	DisplayName       string             `firestore:"displayName"`
	FormattedAddress  string             `firestore:"formattedAddress"`
	Sublocality       string             `firestore:"sublocality,omitempty"`
	City              string             `firestore:"city,omitempty"`
	Country           string             `firestore:"country"`
	StateOrProvince   string             `firestore:"stateOrProvince,omitempty"`
	Continent         string             `firestore:"continent,omitempty"`
	ContinentCode     string             `firestore:"continentCode,omitempty"`
	Latitude          float64            `firestore:"latitude,omitempty"`
	Longitude         float64            `firestore:"longitude,omitempty"`
	Rank              int64              `firestore:"rank"`
	PhotoUri1         string             `firestore:"photoUri1"`
	PhotoUri2         string             `firestore:"photoUri2"`
	Aliases           []string           `firestore:"aliases,omitempty"`
	AddressComponents []AddressComponent `firestore:"addressComponents,omitempty"`
	Types             []string           `firestore:"types,omitempty"`

	// Document references
	CitySafetyRef            *firestore.DocumentRef `firestore:"citySafetyRef,omitempty"`
	CountrySafetyRef         *firestore.DocumentRef `firestore:"countrySafetyRef,omitempty"`
	InternetSpeedRef         *firestore.DocumentRef `firestore:"internetSpeedRef,omitempty"`
	CostOfLivingRef          *firestore.DocumentRef `firestore:"costOfLivingRef,omitempty"`
	CostOfLivingAnalyticsRef *firestore.DocumentRef `firestore:"costOfLivingAnalyticsRef,omitempty"`

	// Monitoring
	Count         int64     `firestore:"count"`
	LastRequested time.Time `firestore:"lastRequested"`
}

func ConstructStandardName(sublocality, city, stateOrProvince, country string) string {
	parts := []string{}
	for _, part := range []string{sublocality, city, stateOrProvince, country} {
		if part == "" {
			continue
		}
		part = utils.RemoveAccentsAndSpecialChars(part)
		parts = append(parts, part)
	}

	combined := strings.Join(parts, "-")
	combined = strings.ToLower(combined)
	combined = strings.ReplaceAll(combined, " ", "")
	combined = strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || r == '-' {
			return r
		}
		return -1
	}, combined)

	for strings.Contains(combined, "--") {
		combined = strings.ReplaceAll(combined, "--", "-")
	}
	combined = strings.Trim(combined, "-")

	return combined
}

func CreateLocationMappingFromMap(data map[string]interface{}) (*LocationMapping, error) {
	mapping := &LocationMapping{}
	getString := func(key string) string {
		if value, ok := data[key].(string); ok {
			return value
		}
		return ""
	}

	getFloat64 := func(key string) float64 {
		if value, ok := data[key].(float64); ok {
			return value
		}
		return 0
	}

	getInt64 := func(key string) int64 {
		if value, ok := data[key].(int64); ok {
			return value
		}
		return 0
	}

	getStringSlice := func(key string) []string {
		if value, ok := data[key].([]interface{}); ok {
			result := make([]string, len(value))
			for i, v := range value {
				if s, ok := v.(string); ok {
					result[i] = s
				}
			}
			return result
		}
		return nil
	}

	mapping.StandardName = getString("standardName")
	mapping.PlaceId = getString("placeId")
	mapping.DisplayName = getString("displayName")
	mapping.FormattedAddress = getString("formattedAddress")
	mapping.Sublocality = getString("sublocality")
	mapping.City = getString("city")
	mapping.Country = getString("country")
	mapping.StateOrProvince = getString("stateOrProvince")
	mapping.Continent = getString("continent")
	mapping.ContinentCode = getString("continentCode")
	mapping.Latitude = getFloat64("latitude")
	mapping.Longitude = getFloat64("longitude")
	mapping.Aliases = getStringSlice("aliases")
	mapping.Types = getStringSlice("types")
	mapping.Count = getInt64("count")
	mapping.Rank = getInt64("rank")

	// Handle AddressComponents
	if addressComponentsData, ok := data["addressComponents"].([]interface{}); ok {
		mapping.AddressComponents = make([]AddressComponent, len(addressComponentsData))
		for i, compData := range addressComponentsData {
			if compMap, ok := compData.(map[string]interface{}); ok {
				mapping.AddressComponents[i] = AddressComponent{
					ShortText: utils.GetString(compMap, "shortText"),
					LongText:  utils.GetString(compMap, "longText"),
					Types:     utils.GetStringSlice(compMap, "types"),
				}
			}
		}
	}

	// Handle document references
	if ref, ok := data["citySafetyRef"].(*firestore.DocumentRef); ok {
		mapping.CitySafetyRef = ref
	}
	if ref, ok := data["countrySafetyRef"].(*firestore.DocumentRef); ok {
		mapping.CountrySafetyRef = ref
	}
	if ref, ok := data["internetSpeedRef"].(*firestore.DocumentRef); ok {
		mapping.InternetSpeedRef = ref
	}
	if ref, ok := data["costOfLivingRef"].(*firestore.DocumentRef); ok {
		mapping.CostOfLivingRef = ref
	}
	if ref, ok := data["costOfLivingAnalyticsRef"].(*firestore.DocumentRef); ok {
		mapping.CostOfLivingAnalyticsRef = ref
	}

	// Validate required fields
	if mapping.StandardName == "" {
		return nil, fmt.Errorf("standardName is required")
	}
	if mapping.Country == "" {
		return nil, fmt.Errorf("country is required")
	}

	return mapping, nil
}

func MapDocumentReferenceToLocationMapping(mapping *LocationMapping, docRef *firestore.DocumentRef, collection string) error {
	switch collection {
	case "city-safety":
		mapping.CitySafetyRef = docRef
	case "country-safety":
		mapping.CountrySafetyRef = docRef
	case "internet-speed-cache":
		mapping.InternetSpeedRef = docRef
	case "cost-of-living":
		mapping.CostOfLivingRef = docRef
	case "cost-of-living-analytics":
		mapping.CostOfLivingAnalyticsRef = docRef
	default:
		err := fmt.Errorf("unsupported collection: %s", collection)
		logger.LogErrorLn("Error", err)
		return err
	}
	return nil
}

func MapAddressComponent(modelComponent AddressComponent) AddressComponent {
	return AddressComponent{
		ShortText: modelComponent.ShortText,
		LongText:  modelComponent.LongText,
		Types:     modelComponent.Types,
	}
}

func MapAddressComponents(modelComponents []AddressComponent) []AddressComponent {
	firestoreComponents := make([]AddressComponent, len(modelComponents))
	for i, component := range modelComponents {
		firestoreComponents[i] = MapAddressComponent(component)
	}
	return firestoreComponents
}
func UniqueNonEmptyStrings(strs ...string) []string {
	unique := make(map[string]bool)
	var result []string
	for _, str := range strs {
		if str != "" && !unique[str] {
			unique[str] = true
			result = append(result, str)
		}
	}
	return result
}

func getContinentCode(continentName string) string {
	switch continentName {
	case "Africa":
		return "AF"
	case "Asia":
		return "AS"
	case "Europe":
		return "EU"
	case "North America":
		return "NA"
	case "South America":
		return "SA"
	case "Oceania":
		return "OC"
	case "Antarctica":
		return "AN"
	default:
		return ""
	}
}
