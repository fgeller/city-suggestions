package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/adrg/strutil"
	"github.com/adrg/strutil/metrics"
	geo "github.com/kellydunn/golang-geo"
)

const (
	countryCodeGB = "GB"

	csvIndexName           = 1
	csvIndexLatitude       = 4
	csvIndexLongitude      = 5
	csvIndexCountryCode    = 8
)

type Suggestion struct {
	Name      string  `json:"name"`
	Latitude  string  `json:"latitude"`
	Longitude string  `json:"longitude"`
	Score     float64 `json:"score"`

	city          *City
	distance      float64
	stringScore   float64
	distanceScore float64
}

type SuggestionsResponse struct {
	Suggestions []*Suggestion `json:"suggestions"`
}

type suggester struct {
	dataPath string
	cities   []*City
}

func newSuggester(fp string) (*suggester, error) {
	s := &suggester{
		dataPath: fp,
		cities:   []*City{},
	}
	err := s.load()
	if err != nil {
		return nil, err
	}
	return s, nil
}

func distance(latA, lonA, latB, lonB float64) float64 {
	pa := geo.NewPoint(latA, lonA)
	pb := geo.NewPoint(latB, lonB)
	return pa.GreatCircleDistance(pb)
}

func (s *suggester) scoreLocation(latitude, longitude string, sugs []*Suggestion) error {
	lat, err := strconv.ParseFloat(latitude, 64)
	if err != nil {
		return fmt.Errorf("failed to parse latitude err=%w", err)
	}
	lon, err := strconv.ParseFloat(longitude, 64)
	if err != nil {
		return fmt.Errorf("failed to parse longitude err=%w", err)
	}

	var maxDistance float64
	minDistance := math.MaxFloat64

	for _, sug := range sugs {
		sug.distance = distance(lat, lon, sug.city.Latitude, sug.city.Longitude)
		if sug.distance > maxDistance {
			maxDistance = sug.distance
		}
		if sug.distance < minDistance {
			minDistance = sug.distance
		}
	}

	den := maxDistance - minDistance
	for _, sug := range sugs {
		sug.distanceScore = 1 - (sug.distance-minDistance)/den
	}

	return nil
}

func (s *suggester) Match(query, latitude, longitude string) ([]*Suggestion, error) {
	sugs := []*Suggestion{}
	q := strings.ToLower(query)

	for _, c := range s.cities {
		if strings.Contains(c.Name, q) {
			stringScore := strutil.Similarity(q, c.Name, metrics.NewJaroWinkler())
			sug := &Suggestion{
				Name:        c.OriginalName,
				Latitude:    fmt.Sprintf("%.5f", c.Latitude),
				Longitude:   fmt.Sprintf("%.5f", c.Longitude),
				city:        c,
				stringScore: stringScore,
				Score:       stringScore,
			}
			sugs = append(sugs, sug)
		}
	}

	withLoc := latitude != "" && longitude != ""
	if withLoc {
		s.scoreLocation(latitude, longitude, sugs)
		for _, sug := range sugs {
			sug.Score = (0.5 * sug.distanceScore) + (0.5 * sug.stringScore)
		}
	}

	return sugs, nil
}

// City represents the Geoname data for a given city, limited to the info that
// the suggester requires to minimize overhead of the data that's held in
// memory.  String values are changed to lower case for comparison, except
// OriginalName, ie Name == strings.ToLower(OriginalName).
type City struct {
	OriginalName string
	Name         string
	Latitude     float64
	Longitude    float64
}

func newCity(csvRec []string) (*City, error) {
	cty := &City{
		OriginalName: csvRec[csvIndexName],
		Name:         strings.ToLower(csvRec[csvIndexName]),
	}

	var err error

	cty.Latitude, err = strconv.ParseFloat(csvRec[csvIndexLatitude], 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse latitude err=%v", err)
	}

	cty.Longitude, err = strconv.ParseFloat(csvRec[csvIndexLongitude], 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse longitude err=%v", err)
	}

	return cty, nil
}

func (s *suggester) load() error {
	f, err := os.Open(s.dataPath)
	if err != nil {
		return fmt.Errorf("failed to open data file err=%w", err)
	}

	cr := csv.NewReader(f)
	for {
		rc, err := cr.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read CSV data err=%w", err)
		}
		if strings.ToLower(strings.TrimSpace(rc[csvIndexCountryCode])) != strings.ToLower(countryCodeGB) {
			continue
		}

		cty, err := newCity(rc)
		if err != nil {
			return fmt.Errorf("failed to unmarshal city err=%w", err)
		}
		s.cities = append(s.cities, cty)
	}

	log.Printf("loaded %v cities.", len(s.cities))
	return nil
}
