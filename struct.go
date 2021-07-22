package twitchpl

type PlaylistManager struct {
	ChannelName               string
	DesiredVariant            string
	StreamPlaybackAccessToken *StreamPlaybackAccessToken
	Variant                   *[]QualityVariant
	Errors                    []error
}

type GraphQLQuery struct {
	OperationName string           `json:"operationName"`
	Query         string           `json:"query"`
	Variables     GraphQLVariables `json:"variables"`
}

type GraphQLVariables struct {
	IsLive     bool   `json:"isLive"`
	IsVod      bool   `json:"isVod"`
	Login      string `json:"login"`
	PlayerType string `json:"playerType"`
	VodID      string `json:"vodID"`
}

type PlaybackAccessTokenGraphQLResponse struct {
	Data PlaybackAccessTokenGraphQLData `json:"data"`
}

type PlaybackAccessTokenGraphQLData struct {
	StreamPlaybackAccessToken StreamPlaybackAccessToken `json:"streamPlaybackAccessToken"`
}

type StreamPlaybackAccessToken struct {
	Signature string `json:"signature"`
	Value     string `json:"value"`
}

type QualityVariant struct {
	Name       string
	Resolution string
	FrameRate  float64
	URL        string
}
