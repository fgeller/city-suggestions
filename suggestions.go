package main

type Suggestion struct {
	Name      string  `json:"name"`
	Latitude  string  `json:"latitude"`
	Longitude string  `json:"longitude"`
	Score     float64 `json:"score"`
}

type SuggestionsResponse struct {
	Suggestions []*Suggestion `json:"suggestions"`
}

func newSuggestionsResponse() *SuggestionsResponse {
	return &SuggestionsResponse{
		Suggestions: []*Suggestion{},
	}
}
