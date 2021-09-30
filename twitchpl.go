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

type PlaylistManager struct {
	ChannelName               string
	Quality                   string
	Resolution                string
	DesiredVariant            string
	StreamPlaybackAccessToken *StreamPlaybackAccessToken
	Variant                   *[]QualityVariant
	Errors                    []error
	Outputter                 *Outputter
}

type Outputter struct {
	Channel    string  `json:"channel"`
	Quality    string  `json:"quality"`
	Resolution string  `json:"resolution"`
	FrameRate  float64 `json:"frame_rate"`
	URL        string  `json:"url"`
}

type QualityVariant struct {
	Name       string
	Resolution string
	FrameRate  float64
	URL        string
}

func Get(channel string) (*PlaylistManager, error) {
	p := newPlaylistManager(channel)

	err := p.getToken()
	if err != nil {
		return &PlaylistManager{}, err
	}

	pl, err := p.getPlaylist()
	if err != nil {
		return pl, err
	}

	return pl, nil
}

func GetMPL(channel string) (string, error) {
	p := newPlaylistManager(channel)

	err := p.getToken()
	if err != nil {
		return "", err
	}

	mpl, err := p.getMasterPlaylist()
	if err != nil {
		return "", errors.New("failed to generate master playlist")
	}

	return fmt.Sprintf("%v", mpl), err
}

func (p *PlaylistManager) updateOutputter() error {
	p.Outputter = &Outputter{}
	p.Outputter.Channel = p.ChannelName
	p.Outputter.Quality = p.Quality

	switch p.Quality {
	case "best":
		p.Outputter.Resolution = (*p.Variant)[0].Resolution
		p.Outputter.FrameRate = (*p.Variant)[0].FrameRate
		p.Outputter.URL = (*p.Variant)[0].URL
	case "worst":
		p.Outputter.Resolution = (*p.Variant)[len(*p.Variant)-2].Resolution
		p.Outputter.FrameRate = (*p.Variant)[len(*p.Variant)-2].FrameRate
		p.Outputter.URL = (*p.Variant)[len(*p.Variant)-2].URL
	case "audio":
		for _, variant := range *p.Variant {
			if variant.Name == "audio_only" {
				p.Outputter.Resolution = variant.Resolution
				p.Outputter.FrameRate = variant.FrameRate
				p.Outputter.URL = variant.URL
			}
		}
	default:
		p.Outputter.Resolution = (*p.Variant)[0].Resolution
		p.Outputter.FrameRate = (*p.Variant)[0].FrameRate
		p.Outputter.URL = (*p.Variant)[0].URL
	}
	return nil
}

func (p *PlaylistManager) AsURL() string {
	err := p.updateOutputter()
	if err != nil {
		return ""
	}
	return p.Outputter.URL
}

func (p *PlaylistManager) AsText() string {
	err := p.updateOutputter()
	if err != nil {
		return ""
	}
	return fmt.Sprintf("Channel: %v \nQuality: %v \nResolution: %v \nFrame Rate: %v \nURL: %v",
		p.Outputter.Channel,
		p.Outputter.Quality,
		p.Outputter.Resolution,
		p.Outputter.FrameRate,
		p.Outputter.URL,
	)
}

func (p *PlaylistManager) AsJSON() string {
	err := p.updateOutputter()
	if err != nil {
		return ""
	}
	bs, err := json.MarshalIndent(&p.Outputter, "", "	")
	if err != nil {
		log.Errorf("couldn't marshal JSON: '%v'", err)
		return ""
	}
	return string(bs)
}

func (p *PlaylistManager) Best() *PlaylistManager {
	p.Quality = "best"
	return p
}

func (p *PlaylistManager) Worst() *PlaylistManager {
	p.Quality = "worst"
	return p
}

func (p *PlaylistManager) Audio() *PlaylistManager {
	p.Quality = "audio"
	return p
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
	pURL, err := p.getMasterPlaylist()
	if err != nil {
		return p, fmt.Errorf("failed to get master playlist: %v\n", err)
	}

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
	defer recoverFromPanic()
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
			return res, err
		}
		log.Errorf("Request error: '%v' Retrying in %v", err, backoff)
		time.Sleep(backoff)

	}
	if err != nil {
		return res, err
	}
	return res, err
}

func newPlaylistManager(channel string) *PlaylistManager {
	return &PlaylistManager{
		ChannelName: channel,
		Quality:     "best", // Best by default
		Outputter:   &Outputter{},
	}
}

func recoverFromPanic() {
	if r := recover(); r != nil {
		fmt.Println("recovered from ", r)
	}
}
