package twpl

import (
	"bytes"
	"encoding/json"
	"fmt"

	//	"log"
	"net/http"
	"net/url"

	//"github.com/k0kubun/pp"
	"github.com/grafov/m3u8"
)

const (
	USHER_API_MASK = "https://usher.ttvnw.net/api/channel/hls/%s.m3u8"
	GRAPHQL_URL    = "https://gql.twitch.tv/gql"
	CLIENT_ID      = "kimne78kx3ncx6brgo4mv6wki5h1ko"
)

var (
	Client *http.Client = http.DefaultClient
)

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

func Get(channelName string) (string, error) {

	p := &PlaylistManager{
		ChannelName: channelName,
	}

	p.getToken()
	variant, err := p.getVariant()
	//	log.Printf("variant: %v", variant)

	return variant, err
}

func GetMPL(channelName string) (string, error) {

	p := &PlaylistManager{
		ChannelName: channelName,
	}

	p.getToken()
	mpl, err := p.getMasterPlaylist()

	return fmt.Sprintf("%v", mpl), err
}

func (p *PlaylistManager) getMasterPlaylist() (*url.URL, error) {
	base_url, _ := url.Parse(fmt.Sprintf(USHER_API_MASK, p.ChannelName))

	v := url.Values{}
	v.Add("allow_source", "true")
	v.Add("fast_bread", "true")
	v.Add("p", "1234567890")
	v.Add("player_backend", "mediaplayer")
	v.Add("sig", p.StreamPlaybackAccessToken.Signature)
	v.Add("supported_codecs", "vp09,avc1")
	v.Add("token", p.StreamPlaybackAccessToken.Value)
	v.Add("cdm", "wv")
	v.Add("player_version", "1.2.0")

	base_url.RawQuery = v.Encode()

	return base_url, nil
}

func (p *PlaylistManager) getVariant() (string, error) {
	base_url, _ := url.Parse(fmt.Sprintf(USHER_API_MASK, p.ChannelName))

	v := url.Values{}
	v.Add("allow_source", "true")
	v.Add("fast_bread", "true")
	v.Add("p", "1234567890")
	v.Add("player_backend", "mediaplayer")
	v.Add("sig", p.StreamPlaybackAccessToken.Signature)
	v.Add("supported_codecs", "vp09,avc1")
	v.Add("token", p.StreamPlaybackAccessToken.Value)
	v.Add("cdm", "wv")
	v.Add("player_version", "1.2.0")

	base_url.RawQuery = v.Encode()
	//	log.Printf("base_url: %v\n", base_url.String())

	req, _ := http.NewRequest("GET", base_url.String(), nil)
	//	req.Header.Add("Client-ID", "kimne78kx3ncx6brgo4mv6wki5h1ko")
	resp, err := http.DefaultClient.Do(req)
	//resp, err := http.Get(base_url.String())
	if err != nil {
		return "", err
	}

	if resp != nil && resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("http.StatusOK failed: got a response from usher: %s", resp.Status)
	}

	playlist, _, _ := m3u8.DecodeFrom(resp.Body, true)
	resp.Body.Close()
	if err != nil {
		return "", err
	}

	//playlist, _, _ := m3u8.DecodeFrom(resp.Body, true)
	master_playlist := playlist.(*m3u8.MasterPlaylist)
	for _, variant := range master_playlist.Variants {
		//pp.Printf("%v\n%v\n", variant.Resolution, variant.URI)
		return variant.URI, nil
		//break
	}

	return "", fmt.Errorf("got a response from usher: %s", resp.Status)
	//pp.Println(master_playlist)
}

func (p *PlaylistManager) getToken() error {
	u, err := url.Parse(GRAPHQL_URL)

	if err != nil {
		return err
	}

	query := NewPlaybackAccessTokenQuery(p.ChannelName)

	buf := new(bytes.Buffer)
	err = json.NewEncoder(buf).Encode(query)
	if err != nil {
		return err
	}

	req, _ := http.NewRequest("POST", u.String(), buf)
	req.Header.Add("Client-ID", CLIENT_ID)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("non-200 code returned for graphql request for PlaybackAccessToken: %s", resp.Status)
	}

	var graphResponse PlaybackAccessTokenGraphQLResponse
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&graphResponse)
	resp.Body.Close()
	if err != nil {
		return err
	}

	p.StreamPlaybackAccessToken = &graphResponse.Data.StreamPlaybackAccessToken

	return nil
}
