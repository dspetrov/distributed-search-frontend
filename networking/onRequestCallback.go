package networking

type OnRequestCallback interface {
	HandleRequest(requestPayload []byte) []byte
	GetEndpoint() string
}
