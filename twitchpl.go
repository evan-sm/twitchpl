package twitchpl

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/grafov/m3u8"
)

const (
	ClientID = "kimne78kx3ncx6brgo4mv6wki5h1ko"
	UsherAPI = "https://usher.ttvnw.net/api/channel/hls/%s.m3u8"
	TTVAPI   = "https://api.ttv.lol/playlist/%s.m3u8"
	GraphURL = "https://gql.twitch.tv/gql"
)

var (
	Client = &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			MaxIdleConnsPerHost: 100,
			IdleConnTimeout:     90 * time.Second,
		},
	}
	UserAgent       = "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:86.0) Gecko/20100101 Firefox/86.0"
	backoffSchedule = []time.Duration{30 * time.Second, 90 * time.Second, 120 * time.Second}
)

type PlaylistManager struct {
	StreamPlaybackAccessToken *StreamPlaybackAccessToken
	Variant                   *[]QualityVariant
	Outputter                 *Outputter
	ChannelName               string
	Quality                   string
	Resolution                string
	DesiredVariant            string
	Errors                    []error
}

type Outputter struct {
	Channel    string  `json:"channel"`
	Quality    string  `json:"quality"`
	Resolution string  `json:"resolution"`
	URL        string  `json:"url"`
	FrameRate  float64 `json:"frame_rate"`
}

type QualityVariant struct {
	Name       string
	Resolution string
	URL        string
	FrameRate  float64
}

func Get(ctx context.Context, channel string, useAdProxy bool) (*PlaylistManager, error) {
	p := newPlaylistManager(channel)

	err := p.getToken(ctx)
	if err != nil {
		return &PlaylistManager{}, err
	}

	pl, err := p.getPlaylist(useAdProxy)
	if err != nil {
		return pl, err
	}

	return pl, nil
}

func GetMPL(ctx context.Context, channel string, useAdProxy bool) (string, error) {
	p := newPlaylistManager(channel)

	err := p.getToken(ctx)
	if err != nil {
		return "", err
	}

	mpl, err := p.getMasterPlaylist(useAdProxy)
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

func (p *PlaylistManager) AsJSON() string {
	err := p.updateOutputter()
	if err != nil {
		return ""
	}

	bs, err := json.Marshal(&p.Outputter)
	if err != nil {
		log.Printf("couldn't marshal JSON: '%v'", err)
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

// generateMasterPlaylistURL generates the master playlist URL
func (p *PlaylistManager) generateMasterPlaylistURL(useAdProxy bool) (mplURL *url.URL, err error) {
	var apiURL string

	switch useAdProxy {
	case true:
		apiURL = fmt.Sprintf(TTVAPI, p.ChannelName)
	default:
		apiURL = fmt.Sprintf(UsherAPI, p.ChannelName)
	}

	mplURL, err = url.Parse(apiURL)
	if err != nil {
		return mplURL, fmt.Errorf("generateMasterPlaylistURL: %w, apiURL: %v", err, apiURL)
	}

	return mplURL, nil
}

func (p *PlaylistManager) getMasterPlaylist(useAdProxy bool) (mplURL *url.URL, err error) {
	mplURL, err = p.generateMasterPlaylistURL(useAdProxy)
	if err != nil {
		return mplURL, fmt.Errorf("getMasterPlaylist: %w", err)
	}

	query := mplURL.Query()

	query.Set("allow_source", "true")
	query.Set("acmb", "e30=")
	query.Set("allow_audio_only", "true")
	query.Set("fast_bread", "true")
	query.Set("playlist_include_framerate", "true")
	query.Set("reassignments_supported", "true")
	query.Set("player_backend", "mediaplayer")
	query.Set("supported_codecs", "vp09,avc1")
	query.Set("p", "1234567890")
	query.Set("play_session_id", "1b0c77f72af01d4db1f993803dacd90f")
	query.Set("cdm", "wv")
	query.Set("player_version", "1.18.0")
	query.Set("player_type", "embed")
	query.Set("sig", p.StreamPlaybackAccessToken.Signature)
	query.Set("token", p.StreamPlaybackAccessToken.Value)

	mplURL.RawQuery = query.Encode()

	return mplURL, nil
}

func (p *PlaylistManager) getPlaylist(useAdProxy bool) (*PlaylistManager, error) {
	pURL, err := p.getMasterPlaylist(useAdProxy)
	if err != nil {
		return p, fmt.Errorf("failed to get master playlist: %w", err)
	}

	URL := pURL.String()

	if useAdProxy {
		path := pURL.Scheme + "://" + pURL.Host + pURL.Path
		query := "%3F" + pURL.RawQuery // puzzled over this for 2 hours "?" needs to be converted to "%3F"
		URL = path + query
	}

	req, err := http.NewRequest(http.MethodGet, URL, http.NoBody)
	if err != nil {
		p.Errors = append(p.Errors, err)
		return p, fmt.Errorf("failed to create GET request: %w", err)
	}

	if useAdProxy {
		req.Header.Set("x-donate-to", "https://ttv.lol/donate") // otherwise you get {"message":"sadge"}
	}

	res, err := p.doRequestWithRetries(req)
	if err != nil {
		p.Errors = append(p.Errors, err)
		return p, fmt.Errorf("failed to make GET request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		if res.StatusCode != http.StatusNotFound {
			p.Errors = append(p.Errors, fmt.Errorf("playlist got http status %s", res.Status))
			return p, fmt.Errorf("error: %w playlist got http status: %v", err, res.Status)
		}
		p.Errors = append(p.Errors, errors.New("stream is offline or channel not found"))
		return p, fmt.Errorf("stream is offline or channel not found")
	}

	if res != nil && res.StatusCode != http.StatusOK {
		p.Errors = append(p.Errors, fmt.Errorf("http.StatusOK failed: got a response from usher: %s", res.Status))
		return p, fmt.Errorf("http.StatusOK failed: got a response from usher: %s", res.Status)
	}

	playlist, _, err := m3u8.DecodeFrom(res.Body, true)
	if err != nil {
		p.Errors = append(p.Errors, err)
		return p, fmt.Errorf("failed to decode m3u8: %w", err)
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

func (p *PlaylistManager) getToken(ctx context.Context) error {
	defer recoverFromPanic()
	u, err := url.Parse(GraphURL)
	if err != nil {
		return err
	}

	query := NewPlaybackAccessTokenQuery(p.ChannelName)
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(query); err != nil {
		return err
	}

	res, err := p.doPostRequestWithRetries(ctx, u.String(), *buf)
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
func (p *PlaylistManager) doRequestWithRetries(req *http.Request) (res *http.Response, err error) {
	for _, backoff := range backoffSchedule {
		res, err = Client.Do(req)
		if err == nil {
			break
		}
		log.Printf("Request error: '%v' Retrying in %v", err, backoff)
		time.Sleep(backoff)

	}
	if err != nil {
		return res, err
	}
	return res, err
}

// doPostRequestWithRetries makes POST request, if failed it retries 3 more times with backoff timer
func (p *PlaylistManager) doPostRequestWithRetries(ctx context.Context, URL string, buf bytes.Buffer) (*http.Response, error) {
	var err error
	var res *http.Response

	for _, backoff := range backoffSchedule {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, URL, &buf)
		if err != nil {
			return res, err
		}
		req.Header.Add("Client-ID", ClientID)
		req.Header.Set("User-Agent", UserAgent)

		res, err = Client.Do(req)
		if err == nil {
			return res, err
		}
		log.Printf("Request error: '%v' Retrying in %v", err, backoff)
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
		log.Println("recovered from ", r)
	}
}
