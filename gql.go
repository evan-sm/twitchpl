package twitchpl

func NewPlaybackAccessTokenQuery(login string) GraphQLQuery {
	return GraphQLQuery{
		OperationName: "PlaybackAccessToken_Template",
		Query:         `query PlaybackAccessToken_Template($login: String!, $isLive: Boolean!, $vodID: ID!, $isVod: Boolean!, $playerType: String!) {  streamPlaybackAccessToken(channelName: $login, params: {platform: "web", playerBackend: "mediaplayer", playerType: $playerType}) @include(if: $isLive) {    value    signature    __typename  }  videoPlaybackAccessToken(id: $vodID, params: {platform: "web", playerBackend: "mediaplayer", playerType: $playerType}) @include(if: $isVod) {    value    signature    __typename  }}`,
		Variables: GraphQLVariables{
			IsLive:     true,
			Login:      login,
			PlayerType: "site",
		},
	}
}

type GraphQLQuery struct {
	OperationName string           `json:"operationName"`
	Query         string           `json:"query"`
	Variables     GraphQLVariables `json:"variables"`
}

type GraphQLVariables struct {
	Login      string `json:"login"`
	PlayerType string `json:"playerType"`
	VodID      string `json:"vodID"`
	IsLive     bool   `json:"isLive"`
	IsVod      bool   `json:"isVod"`
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
