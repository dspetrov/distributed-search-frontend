package search

import (
	"dspetrov/distributed-search-frontend/clusterManagement"
	"dspetrov/distributed-search-frontend/model"
	"dspetrov/distributed-search-frontend/networking"
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"github.com/gogo/protobuf/proto"
)

const (
	ENDPOINT           = "/documents_search"
	DOCUMENTS_LOCATION = "books"
)

type UserSearchHandler struct {
	client                    *networking.WebClient
	searchCoordinatorRegistry *clusterManagement.ServiceRegistry
}

func NewUserSearchHandler(searchCoordinatorRegistry *clusterManagement.ServiceRegistry) *UserSearchHandler {
	ush := UserSearchHandler{
		searchCoordinatorRegistry: searchCoordinatorRegistry,
		client:                    networking.NewWebClient(),
	}

	return &ush
}

func (ush UserSearchHandler) HandleRequest(requestPayload []byte) []byte {
	var frontendSearchRequest model.FrontendSearchRequest
	if err := json.Unmarshal(requestPayload, &frontendSearchRequest); err != nil {
		fmt.Println(err)
		return []byte{}
	}

	frontendSearchResponse := ush.createFrontendResponse(frontendSearchRequest)

	frontendSearchResponseBytes, err := json.Marshal(frontendSearchResponse)
	if err != nil {
		fmt.Println(err)
		return []byte{}
	}

	return frontendSearchResponseBytes
}

func (ush UserSearchHandler) createFrontendResponse(frontendSearchRequest model.FrontendSearchRequest) *model.FrontendSearchResponse {
	searchClusterResponse := ush.sendRequestToSearchCluster(frontendSearchRequest.SearchQuery)

	filteredResults := filterResults(searchClusterResponse, int(frontendSearchRequest.MaxNumberOfResults), frontendSearchRequest.MinScore)

	return &model.FrontendSearchResponse{
		SearchResults:     filteredResults,
		DocumentsLocation: DOCUMENTS_LOCATION,
	}
}

func filterResults(searchClusterResponse *model.Response, maxResults int, minScore float64) []*model.SearchResultInfo {
	maxScore := getMaxScore(searchClusterResponse)

	searchResultInfoList := []*model.SearchResultInfo{}

	for i := 0; i < len(searchClusterResponse.RelevantDocuments) && i < maxResults; i++ {
		normalizedDocumentScore := normalizeScore(searchClusterResponse.RelevantDocuments[i].Score, maxScore)
		if float64(normalizedDocumentScore) < minScore {
			continue // break in the lecture is an error
		}

		documentName := searchClusterResponse.RelevantDocuments[i].DocumentName

		title := getDocumentTitle(documentName)
		extension := getDocumentExtension(documentName)

		resultInfo := model.SearchResultInfo{
			Title:     title,
			Extension: extension,
			Score:     int(normalizedDocumentScore),
		}

		searchResultInfoList = append(searchResultInfoList, &resultInfo)
	}

	return searchResultInfoList
}

func (ush UserSearchHandler) GetEndpoint() string {
	return ENDPOINT
}

func getDocumentExtension(document string) string {
	parts := strings.Split(document, ".")
	if len(parts) == 2 {
		return parts[1]
	}

	return ""
}

func getDocumentTitle(document string) string {
	return strings.Split(document, ".")[0]
}

func normalizeScore(inputScore float64, maxScore float64) int {
	return int(math.Ceil(inputScore * 100.0 / maxScore))
}

func getMaxScore(searchClusterResponse *model.Response) float64 {
	if len(searchClusterResponse.RelevantDocuments) == 0 {
		return 0
	}

	maxScore := 0.0
	for _, docStat := range searchClusterResponse.RelevantDocuments {
		if docStat.Score > maxScore {
			maxScore = docStat.Score
		}
	}

	return maxScore
}

func (ush UserSearchHandler) sendRequestToSearchCluster(searchQuery string) *model.Response {
	searchRequest := model.Request{
		SearchQuery: searchQuery,
	}

	coordinatorAddress := ush.searchCoordinatorRegistry.GetRandomServiceAddress()
	if coordinatorAddress == "" {
		fmt.Println("Search Cluster Coordinator is unavailable")
		return &model.Response{}
	}

	searchRequestBytes, err := proto.Marshal(&searchRequest)
	if err != nil {
		fmt.Println(err)
		return &model.Response{}
	}

	ch := make(chan []byte)
	go ush.client.SendTask(coordinatorAddress, searchRequestBytes, ch)
	payloadBody := <-ch

	var response model.Response
	if err := proto.Unmarshal(payloadBody, &response); err != nil {
		fmt.Println(err)
		return &model.Response{}
	}

	return &response
}
