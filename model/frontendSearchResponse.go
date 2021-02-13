package model

type FrontendSearchResponse struct {
	SearchResults     []*SearchResultInfo
	DocumentsLocation string
}

func NewFrontendSearchResponse(searchResults []*SearchResultInfo, documentsLocation string) *FrontendSearchResponse {
	fsr := FrontendSearchResponse{
		SearchResults:     searchResults,
		DocumentsLocation: documentsLocation,
	}

	return &fsr
}

type SearchResultInfo struct {
	Title     string
	Extension string
	Score     int
}

func NewSearchResultInfo(title string, extension string, score int) *SearchResultInfo {
	sri := SearchResultInfo{
		Title:     title,
		Extension: extension,
		Score:     score,
	}

	return &sri
}
