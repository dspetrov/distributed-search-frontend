package model

type FrontendSearchRequest struct {
	SearchQuery        string
	MaxNumberOfResults int
	MinScore           float64
}

func NewFrontendSearchRequest() *FrontendSearchRequest {
	maxInt := int(^uint(0) >> 1)

	fsr := FrontendSearchRequest{
		MaxNumberOfResults: maxInt,
		MinScore:           0,
	}

	return &fsr
}
