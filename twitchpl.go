package twitchpl

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/grafov/m3u8"
	log "github.com/sirupsen/logrus"
)

const (
	CLIENT_ID      = "kimne78kx3ncx6brgo4mv6wki5h1ko"
	USHER_API_MASK = "https://usher.ttvnw.net/api/channel/hls/%s.m3u8"
	GRAPHQL_URL    = "https://gql.twitch.tv/gql"
)

var (
	Client = &http.Client{Timeout: 2 * time.Second,
		Transport: &http.Transport{
			MaxIdleConnsPerHost: 100,
			IdleConnTimeout:     90 * time.Second},
	}
	USER_AGENT      = "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:86.0) Gecko/20100101 Firefox/86.0"
	backoffSchedule = []time.Duration{1 * time.Second, 2 * time.Second, 3 * time.Second}
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

func Get(channelName string) (*PlaylistManager, error) {

	p := &PlaylistManager{
		ChannelName: channelName,
	}

	p.getToken()
	pl, err := p.getPlaylist()
	if err != nil {
		return pl, err
	}

	return pl, nil
}

func (p *PlaylistManager) Quality(q string) (string, error) {
	if len(*p.Variant) == 0 {
		return "", errors.New("there's no stream quality to choose")
	}
	switch q {
	case "best":
		return (*p.Variant)[0].URL, nil
	case "1080p":
		for _, variant := range *p.Variant {
			if variant.Resolution == "1920x1080" {
				return variant.URL, nil
			}
		}
		return (*p.Variant)[0].URL, nil
	case "720p":
		for _, variant := range *p.Variant {
			if variant.Resolution == "1280x720" {
				return variant.URL, nil
			}
		}
		return (*p.Variant)[0].URL, nil
	case "480p":
		for _, variant := range *p.Variant {
			if variant.Resolution == "852x480" {
				return variant.URL, nil
			}
		}
		return (*p.Variant)[0].URL, nil
	case "360p":
		for _, variant := range *p.Variant {
			if variant.Resolution == "640x360" {
				return variant.URL, nil
			}
		}
		return (*p.Variant)[0].URL, nil
	case "160p":
		for _, variant := range *p.Variant {
			if variant.Resolution == "284x160" {
				return variant.URL, nil
			}
		}
		return (*p.Variant)[0].URL, nil
	case "audio":
		for _, variant := range *p.Variant {
			if variant.Name == "audio_only" {
				return variant.URL, nil
			}
		}
	case "worst":
		return (*p.Variant)[len(*p.Variant)-2].URL, nil
	default:
		return (*p.Variant)[0].URL, nil
	}
	return "", errors.New("failed to choose quality")
}

func (p *PlaylistManager) Best() (string, error) {
	if len(*p.Variant) == 0 {
		return "", errors.New("there's no stream quality to choose")
	}
	return (*p.Variant)[0].URL, nil
}

func (p *PlaylistManager) Worst() (string, error) {
	if len(*p.Variant) == 0 {
		return "", errors.New("there's no stream quality to choose")
	}
	return (*p.Variant)[len(*p.Variant)-2].URL, nil
}

func (p *PlaylistManager) Audio() (string, error) {
	if len(*p.Variant) == 0 {
		return "", errors.New("there's no stream quality to choose")
	}
	for _, variant := range *p.Variant {
		if variant.Name == "audio_only" {
			return variant.URL, nil
		}
	}
	return (*p.Variant)[0].URL, errors.New("failed to find audio_only track")
}

func GetMPL(channelName string) (string, error) {

	p := &PlaylistManager{
		ChannelName: channelName,
	}

	p.getToken()
	mpl, err := p.getMasterPlaylist()
	if err != nil {
		return "", errors.New("failed to generate master playlist")
	}

	return fmt.Sprintf("%v", mpl), err
}

func (p *PlaylistManager) getMasterPlaylist() (*url.URL, error) {
	mplURL, err := url.Parse(fmt.Sprintf(USHER_API_MASK, p.ChannelName))
	if err != nil {
		return mplURL, errors.New("failed to generate master playlist")
	}

	query := mplURL.Query()

	query.Set("allow_source", "true")
	query.Set("fast_bread", "true")
	query.Set("p", "1234567890")
	query.Set("player_backend", "mediaplayer")
	query.Set("sig", p.StreamPlaybackAccessToken.Signature)
	query.Set("supported_codecs", "vp09,avc1")
	query.Set("token", p.StreamPlaybackAccessToken.Value)
	query.Set("cdm", "wv")
	query.Set("player_version", "1.2.0")
	query.Set("player_type", "embed")

	mplURL.RawQuery = query.Encode()

	return mplURL, nil
}

func (p *PlaylistManager) getPlaylist() (*PlaylistManager, error) {
	pURL, err := url.Parse(fmt.Sprintf(USHER_API_MASK, p.ChannelName))
	if err != nil {
		p.Errors = append(p.Errors, err)
		return p, fmt.Errorf("%v\n", err)
	}

	query := pURL.Query()

	query.Set("allow_source", "true")
	query.Set("fast_bread", "true")
	query.Set("allow_audio_only", "true")
	query.Set("sig", p.StreamPlaybackAccessToken.Signature)
	query.Set("token", p.StreamPlaybackAccessToken.Value)
	query.Set("player_type", "embed")

	pURL.RawQuery = query.Encode()

	req, err := http.NewRequest("GET", pURL.String(), nil)
	if err != nil {
		p.Errors = append(p.Errors, err)
		return p, fmt.Errorf("failed to create GET request: %v\n", err)
	}

	res, err := p.doRequestWithRetries(req)
	if err != nil {
		p.Errors = append(p.Errors, err)
		return p, fmt.Errorf("failed to make GET request: %v\n", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		if res.StatusCode != http.StatusNotFound {
			p.Errors = append(p.Errors, fmt.Errorf("playlist got http status %s", res.Status))
			return p, fmt.Errorf("error: %v playlist got http status: %v\n", err, res.Status)
		}
		p.Errors = append(p.Errors, errors.New("stream is offline or channel not found"))
		return p, fmt.Errorf("stream is offline or channel not found\n")
	}

	if res != nil && res.StatusCode != http.StatusOK {
		p.Errors = append(p.Errors, fmt.Errorf("http.StatusOK failed: got a response from usher: %s", res.Status))
		return p, fmt.Errorf("http.StatusOK failed: got a response from usher: %s", res.Status)
	}

	playlist, _, err := m3u8.DecodeFrom(res.Body, true)
	if err != nil {
		p.Errors = append(p.Errors, err)
		return p, fmt.Errorf("failed to decode m3u8: %v", err)
	}

	masterpl := playlist.(*m3u8.MasterPlaylist)

	var quality []QualityVariant
	for _, variant := range masterpl.Variants {
		quality = append(quality, QualityVariant{
			Name:       variant.Alternatives[0].Name,
			Resolution: variant.Resolution,
			FrameRate:  variant.FrameRate,
			URL:        variant.URI,
		})

	}

	p.Variant = &quality
	return p, nil
}

func (p *PlaylistManager) getToken() error {
	u, err := url.Parse(GRAPHQL_URL)
	if err != nil {
		return err
	}

	query := NewPlaybackAccessTokenQuery(p.ChannelName)
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(query); err != nil {
		return err
	}

	res, err := p.doPostRequestWithRetries(u.String(), *buf)
	if err != nil {
		return fmt.Errorf("non-200 code returned for graphql request for PlaybackAccessToken: %s", res.Status)
	}
	defer res.Body.Close()

	var graphResponse PlaybackAccessTokenGraphQLResponse
	if err := json.NewDecoder(res.Body).Decode(&graphResponse); err != nil {
		return err
	}
	p.StreamPlaybackAccessToken = &graphResponse.Data.StreamPlaybackAccessToken

	return nil
}

// doRequestWithRetries makes request, if failed it retries 3 more times with backoff timer
func (p *PlaylistManager) doRequestWithRetries(req *http.Request) (*http.Response, error) {
	var err error
	var res *http.Response

	for _, backoff := range backoffSchedule {
		res, err = Client.Do(req)
		if err == nil {
			break
		}
		log.Errorf("Request error: '%v' Retrying in %v", err, backoff)
		time.Sleep(backoff)

	}
	if err != nil {
		return res, err
	}
	return res, err
}

// doPostRequestWithRetries makes POST request, if failed it retries 3 more times with backoff timer
func (p *PlaylistManager) doPostRequestWithRetries(url string, buf bytes.Buffer) (*http.Response, error) {
	var err error
	var res *http.Response

	for _, backoff := range backoffSchedule {
		req, err := http.NewRequest("POST", url, &buf)
		if err != nil {
			return res, err
		}
		req.Header.Add("Client-ID", CLIENT_ID)
		req.Header.Set("User-Agent", USER_AGENT)

		res, err = Client.Do(req)
		if err == nil {
			break
		}
		log.Errorf("Request error: '%v' Retrying in %v", err, backoff)
		time.Sleep(backoff)

	}
	if err != nil {
		return res, err
	}
	return res, err
}
