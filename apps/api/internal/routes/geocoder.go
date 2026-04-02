package routes

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Geocoder interface {
	GeocodeCity(ctx context.Context, city string) (latitude float64, longitude float64, err error)
	SearchCityCandidates(ctx context.Context, query string, limit int) ([]CityCandidate, error)
}

type NominatimGeocoder struct {
	client      *http.Client
	endpointURL string
	userAgent   string
	countryCode string
}

func NewNominatimGeocoderFromEnv() Geocoder {
	endpointURL := strings.TrimSpace(os.Getenv("ROUTES_GEOCODING_URL"))
	if endpointURL == "" {
		endpointURL = "https://nominatim.openstreetmap.org/search"
	}

	userAgent := strings.TrimSpace(os.Getenv("ROUTES_GEOCODING_USER_AGENT"))
	if userAgent == "" {
		userAgent = "schumacher-tur/1.0 (routes geocoder)"
	}

	countryCode := strings.TrimSpace(os.Getenv("ROUTES_GEOCODING_COUNTRY_CODE"))
	if countryCode == "" {
		countryCode = "br"
	}

	return &NominatimGeocoder{
		client: &http.Client{
			Timeout: 8 * time.Second,
		},
		endpointURL: endpointURL,
		userAgent:   userAgent,
		countryCode: strings.ToLower(countryCode),
	}
}

type nominatimSearchResult struct {
	PlaceID     int64            `json:"place_id"`
	Lat         string           `json:"lat"`
	Lon         string           `json:"lon"`
	DisplayName string           `json:"display_name"`
	Addresstype string           `json:"addresstype"`
	Address     nominatimAddress `json:"address"`
}

type nominatimAddress struct {
	City         string `json:"city"`
	Town         string `json:"town"`
	Village      string `json:"village"`
	Municipality string `json:"municipality"`
	County       string `json:"county"`
	State        string `json:"state"`
	ISO3166Lvl4  string `json:"ISO3166-2-lvl4"`
	CountryCode  string `json:"country_code"`
}

var brazilStateCodes = map[string]string{
	"ACRE":                "AC",
	"ALAGOAS":             "AL",
	"AMAPA":               "AP",
	"AMAZONAS":            "AM",
	"BAHIA":               "BA",
	"CEARA":               "CE",
	"DISTRITO FEDERAL":    "DF",
	"ESPIRITO SANTO":      "ES",
	"GOIAS":               "GO",
	"MARANHAO":            "MA",
	"MATO GROSSO":         "MT",
	"MATO GROSSO DO SUL":  "MS",
	"MINAS GERAIS":        "MG",
	"PARA":                "PA",
	"PARAIBA":             "PB",
	"PARANA":              "PR",
	"PERNAMBUCO":          "PE",
	"PIAUI":               "PI",
	"RIO DE JANEIRO":      "RJ",
	"RIO GRANDE DO NORTE": "RN",
	"RIO GRANDE DO SUL":   "RS",
	"RONDONIA":            "RO",
	"RORAIMA":             "RR",
	"SANTA CATARINA":      "SC",
	"SAO PAULO":           "SP",
	"SERGIPE":             "SE",
	"TOCANTINS":           "TO",
}

func (g *NominatimGeocoder) GeocodeCity(ctx context.Context, city string) (float64, float64, error) {
	queries := geocodingQueries(city)
	if len(queries) == 0 {
		return 0, 0, errors.New("city is required")
	}

	var lastErr error
	for _, query := range queries {
		candidates, err := g.SearchCityCandidates(ctx, query, 10)
		if err == nil && len(candidates) > 0 {
			return candidates[0].Latitude, candidates[0].Longitude, nil
		}
		if err != nil {
			lastErr = err
		}
	}

	if lastErr == nil {
		lastErr = errors.New("no geocoding result")
	}
	return 0, 0, lastErr
}

func (g *NominatimGeocoder) SearchCityCandidates(ctx context.Context, query string, limit int) ([]CityCandidate, error) {
	trimmed := strings.TrimSpace(query)
	if trimmed == "" {
		return nil, errors.New("query is required")
	}

	results, err := g.search(ctx, trimmed, limit)
	if err != nil {
		return nil, err
	}

	candidates := make([]CityCandidate, 0, len(results))
	seen := make(map[string]struct{}, len(results))

	for _, item := range results {
		candidate, ok := buildCityCandidate(item)
		if !ok {
			continue
		}
		key := strings.ToLower(fmt.Sprintf("%s|%.6f|%.6f", candidate.City, candidate.Latitude, candidate.Longitude))
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		candidates = append(candidates, candidate)
	}

	if len(candidates) == 0 {
		return []CityCandidate{}, nil
	}
	return candidates, nil
}

func (g *NominatimGeocoder) search(ctx context.Context, query string, limit int) ([]nominatimSearchResult, error) {
	u, err := url.Parse(g.endpointURL)
	if err != nil {
		return nil, err
	}

	if limit <= 0 || limit > 10 {
		limit = 10
	}

	params := u.Query()
	params.Set("format", "jsonv2")
	params.Set("limit", strconv.Itoa(limit))
	params.Set("addressdetails", "1")
	params.Set("q", query)
	if g.countryCode != "" {
		params.Set("countrycodes", g.countryCode)
	}
	u.RawQuery = params.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", g.userAgent)

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("geocoding provider returned status %d", resp.StatusCode)
	}

	results := make([]nominatimSearchResult, 0, limit)
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, err
	}

	sort.SliceStable(results, func(i, j int) bool {
		return candidatePriority(results[i].Addresstype) < candidatePriority(results[j].Addresstype)
	})
	return results, nil
}

func geocodingQueries(rawCity string) []string {
	trimmed := strings.TrimSpace(rawCity)
	if trimmed == "" {
		return nil
	}

	candidates := []string{}
	displayName, stopName, state := normalizeStopInput(trimmed)
	if displayName != "" && state != "" {
		candidates = append(candidates, fmt.Sprintf("%s, %s, Brasil", stopName, state))
	}
	candidates = append(candidates, trimmed)
	candidates = append(candidates, fmt.Sprintf("%s, Brasil", trimmed))

	seen := make(map[string]struct{}, len(candidates))
	result := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		value := strings.TrimSpace(candidate)
		if value == "" {
			continue
		}
		key := strings.ToLower(value)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, value)
	}
	return result
}

func buildCityCandidate(raw nominatimSearchResult) (CityCandidate, bool) {
	latitude, err := strconv.ParseFloat(raw.Lat, 64)
	if err != nil {
		return CityCandidate{}, false
	}
	longitude, err := strconv.ParseFloat(raw.Lon, 64)
	if err != nil {
		return CityCandidate{}, false
	}

	cityName := strings.TrimSpace(firstNonEmpty(
		raw.Address.City,
		raw.Address.Town,
		raw.Address.Municipality,
		raw.Address.Village,
		raw.Address.County,
	))
	if cityName == "" {
		parts := strings.Split(raw.DisplayName, ",")
		if len(parts) > 0 {
			cityName = strings.TrimSpace(parts[0])
		}
	}
	if cityName == "" {
		return CityCandidate{}, false
	}

	stateCode := stateCodeFromISO(raw.Address.ISO3166Lvl4)
	if stateCode == "" {
		stateCode = stateCodeFromName(raw.Address.State)
	}

	city := cityName
	if stateCode != "" {
		city = fmt.Sprintf("%s/%s", cityName, stateCode)
	}

	placeID := strings.TrimSpace(strconv.FormatInt(raw.PlaceID, 10))
	if placeID == "0" {
		placeID = strings.ToLower(strings.ReplaceAll(city, " ", "_"))
	}

	return CityCandidate{
		PlaceID:     placeID,
		City:        city,
		StateCode:   stateCode,
		DisplayName: strings.TrimSpace(raw.DisplayName),
		Addresstype: strings.TrimSpace(raw.Addresstype),
		Latitude:    latitude,
		Longitude:   longitude,
	}, true
}

func candidatePriority(addresstype string) int {
	switch strings.ToLower(strings.TrimSpace(addresstype)) {
	case "city":
		return 0
	case "town":
		return 1
	case "municipality":
		return 2
	case "village":
		return 3
	case "administrative":
		return 4
	default:
		return 10
	}
}

func stateCodeFromISO(iso string) string {
	trimmed := strings.TrimSpace(strings.ToUpper(iso))
	if strings.HasPrefix(trimmed, "BR-") && len(trimmed) >= 5 {
		return trimmed[len(trimmed)-2:]
	}
	return ""
}

func stateCodeFromName(state string) string {
	key := normalizeBrazilStateName(state)
	if key == "" {
		return ""
	}
	if code, ok := brazilStateCodes[key]; ok {
		return code
	}
	return ""
}

func normalizeBrazilStateName(state string) string {
	upper := strings.ToUpper(strings.TrimSpace(state))
	if upper == "" {
		return ""
	}
	replacer := strings.NewReplacer(
		"\u00C1", "A",
		"\u00C0", "A",
		"\u00C3", "A",
		"\u00C2", "A",
		"\u00C4", "A",
		"\u00C9", "E",
		"\u00CA", "E",
		"\u00C8", "E",
		"\u00CB", "E",
		"\u00CD", "I",
		"\u00CE", "I",
		"\u00CC", "I",
		"\u00CF", "I",
		"\u00D3", "O",
		"\u00D4", "O",
		"\u00D5", "O",
		"\u00D2", "O",
		"\u00D6", "O",
		"\u00DA", "U",
		"\u00DB", "U",
		"\u00D9", "U",
		"\u00DC", "U",
		"\u00C7", "C",
	)
	return replacer.Replace(upper)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}
