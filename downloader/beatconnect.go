package downloader

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

const BeatConnectAPIBase = "https://beatconnect.io"

// BeatConnectClient is a client for the BeatConnect API.
type BeatConnectClient struct {
	httpClient *http.Client
	apiToken   string
}

// NewBeatConnectClient returns a new BeatConnectClient.
func NewBeatConnectClient(apiToken string) *BeatConnectClient {
	cl := http.DefaultClient
	// I really wish we had context.Context back in the day.
	cl.Timeout = time.Minute * 2
	return &BeatConnectClient{
		httpClient: cl,
		apiToken:   apiToken,
	}
}

// newRequest creates a new http.Request with the given relativeURL.
// If withToken is true, it adds the 'token' query parameter to the request.
func (c *BeatConnectClient) newRequest(relativeURL string, withToken bool) (*http.Request, error) {
	req, err := http.NewRequest(
		http.MethodGet,
		BeatConnectAPIBase+"/"+relativeURL,
		nil,
	)
	if err != nil {
		return nil, err
	}
	if withToken {
		query := req.URL.Query()
		query.Set("token", c.apiToken)
		req.URL.RawQuery = query.Encode()
	}
	return req, nil
}

// HasVideo returns whether the beatmap set with the given ID has a video
// according to the BeatConnect API.
/* func (c *BeatConnectClient) HasVideo(setID int) (bool, error) {
	req, err := c.newRequest(fmt.Sprintf("api/beatmaps/%d/", setID))
	if err != nil {
		return false, fmt.Errorf("new request: %w", err)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("bad status code for %q: %d", req.URL, resp.StatusCode)
	}
	var body struct {
		Video bool `json:"video"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return false, fmt.Errorf("decode preliminary response: %w", err)
	}
	return body.Video, nil
} */

// Download downloads the osz file from BeatConnect.
// The noVideo flag is ignored, the video is always included.
func (c *BeatConnectClient) Download(setID int, noVideo bool) (io.ReadCloser, error) {
	oszReq, err := c.newRequest(fmt.Sprintf("b/%d", setID), false)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}
	resp, err := c.httpClient.Do(oszReq)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("bad status code: %d", resp.StatusCode)
	}
	return resp.Body, err
}
