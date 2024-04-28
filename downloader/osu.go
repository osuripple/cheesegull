// Package downloader implements downloading from the osu! website, through,
// well, mostly scraping and dirty hacks.
package downloader

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
)

// OsuClient is a client that can download beatmaps from the osu! website.
type OsuClient struct {
	// httpClient is the http.Client used to do requests.
	httpClient *http.Client

	// username is the osu! username used to log in.
	username string

	// password is the osu! password used to log in.
	password string

	// requestPreparer is the LogInRequestPreparer used to prepare cookies and headers.
	// The zero-value is an EmptyLogInRequestPreparer.
	requestPreparer LogInRequestPreparer
}

// NewOsuClient returns a new OsuClient that is logged in.
// It returns an error if the login fails.
func NewOsuClient(username, password string, requestPreparer LogInRequestPreparer) (*OsuClient, error) {
	cl := &OsuClient{
		httpClient:      http.DefaultClient,
		username:        username,
		password:        password,
		requestPreparer: requestPreparer,
	}
	if err := cl.logIn(); err != nil {
		return nil, fmt.Errorf("login: %w", err)
	}
	return cl, nil
}

// logIn logs in into an osu! account and replaces the http.Client with the one containing the cookies.
func (c *OsuClient) logIn() error {
	// Prepare cookies
	cookieJar, err := c.requestPreparer.PrepareCookies()
	if err != nil {
		return fmt.Errorf("prepare cookies: %w", err)
	}
	httpClient := &http.Client{
		Jar: cookieJar,
	}

	// Prepare headers
	preparedHeaders, err := c.requestPreparer.PrepareHeaders()
	if err != nil {
		return fmt.Errorf("prepare headers: %w", err)
	}

	// POST form values
	vals := url.Values{}
	vals.Add("redirect", "/")
	vals.Add("sid", "")
	vals.Add("username", c.username)
	vals.Add("password", c.password)
	vals.Add("autologin", "on")
	vals.Add("login", "Login")
	req, err := http.NewRequest("POST", "https://osu.ppy.sh/forum/ucp.php?mode=login", strings.NewReader(vals.Encode()))
	if err != nil {
		return err
	}
	// Add headers from requestPreparer
	for k, v := range preparedHeaders {
		req.Header.Set(k, v)
	}

	// Other headers
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Content-Length", strconv.Itoa(len(vals.Encode())))
	req.Header.Set("accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3")
	req.Header.Set("accept-encoding", "gzip, deflate, br")
	req.Header.Set("accept-language", "it-IT,it;q=0.9,en-US;q=0.8,en;q=0.7")
	req.Header.Set("cache-control", "max-age=0")
	req.Header.Set("origin", "https://osu.ppy.sh")
	req.Header.Set("referer", "https://osu.ppy.sh/forum/ucp.php?mode=login")
	req.Header.Set("sec-fetch-mode", "navigate")
	req.Header.Set("sec-fetch-site", "same-origin")
	req.Header.Set("sec-fetch-user", "?1")
	req.Header.Set("upgrade-insecure-requests", "1")

	// Do the request
	loginResp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	if loginResp.Request.URL.Path != "/home" {
		return errors.New("cheesegull/downloader: could not log in (was not redirected to /home) but to " + loginResp.Request.URL.Path)
	}

	// Replace the client's http.Client with the one we just used
	// which contains the correct cookies
	c.httpClient = httpClient
	return nil
}

// HasVideo checks whether a beatmap has a video.
func (c *OsuClient) HasVideo(setID int) (bool, error) {
	page, err := c.httpClient.Get(fmt.Sprintf("https://old.ppy.sh/s/%d", setID))
	if err != nil {
		return false, err
	}
	defer page.Body.Close()
	body, err := io.ReadAll(page.Body)
	if err != nil {
		return false, err
	}
	return bytes.Contains(body, []byte(fmt.Sprintf(`href="/d/%dn"`, setID))), nil
}

// Download downloads a beatmap from the osu! website. noVideo specifies whether
// we should request the beatmap to not have the video.
func (c *OsuClient) Download(setID int, noVideo bool) (io.ReadCloser, error) {
	suffix := ""
	if noVideo {
		suffix = "n"
	}
	resp, err := c.httpClient.Get("https://old.ppy.sh/d/" + strconv.Itoa(setID) + suffix)
	if err != nil {
		return nil, err
	}
	if resp.Request.URL.Host == "old.ppy.sh" {
		resp.Body.Close()
		return nil, ErrNoRedirect
	}
	return resp.Body, nil
}

// LogInRequestPreparer prepares cookies to set before
// doing any osu! website http request (log-in)
type LogInRequestPreparer interface {
	PrepareCookies() (http.CookieJar, error)
	PrepareHeaders() (map[string]string, error)
}

// EmptyLogInRequestPreparer is a cookie preparer that returns
// an empty CookieJar
type EmptyLogInRequestPreparer struct{}

// PrepareCookies returns an empty cookie jar
func (*EmptyLogInRequestPreparer) PrepareCookies() (http.CookieJar, error) {
	j, err := cookiejar.New(&cookiejar.Options{})
	if err != nil {
		return nil, err
	}
	return j, nil
}

// PrepareHeaders returns an empty map (i.e.: no additional headers)
func (*EmptyLogInRequestPreparer) PrepareHeaders() (map[string]string, error) {
	return map[string]string{}, nil
}

// FckCf is a LogInRequestPreparer that prepares cookies using FckCf
type FckCf struct {
	Address string

	proxyResp map[string]interface{}
}

// InitializeProxyResponse initializes the
func (c *FckCf) InitializeProxyResponse() error {
	buf := bytes.NewBuffer([]byte(`{"url":"https://old.ppy.sh/forum/ucp.php?mode=login"}`))
	fckCfResp, err := http.Post(c.Address, "application/json", buf)
	if err != nil {
		return fmt.Errorf("proxy request: %w", err)
	}
	if err := json.NewDecoder(fckCfResp.Body).Decode(&c.proxyResp); err != nil {
		return fmt.Errorf("json decode: %w", err)
	}
	return nil
}

func (c *FckCf) ppyCookie(name, value string) *http.Cookie {
	return &http.Cookie{
		Name:   name,
		Value:  value,
		Path:   "/",
		Domain: ".ppy.sh",
	}
}

// PrepareCookies calls FckCf and returns a CookieJar
// with the required cookies
func (c *FckCf) PrepareCookies() (http.CookieJar, error) {
	if c.proxyResp == nil {
		// Initialize proxy response only once
		if err := c.InitializeProxyResponse(); err != nil {
			return nil, fmt.Errorf("init proxy response: %w", err)
		}
	}
	j, err := cookiejar.New(&cookiejar.Options{})
	if err != nil {
		return nil, err
	}
	u, err := url.Parse("https://osu.ppy.sh")
	if err != nil {
		return nil, err
	}
	j.SetCookies(u, []*http.Cookie{
		c.ppyCookie("__cfduid", c.proxyResp["__cfduid"].(string)),
		c.ppyCookie("cf_clearance", c.proxyResp["cf_clearance"].(string)),
	})
	return j, nil
}

func (c *FckCf) PrepareHeaders() (map[string]string, error) {
	if c.proxyResp == nil {
		return nil, errors.New("must first call PrepareCookies to set the fckcf resp")
	}
	return map[string]string{"User-Agent": c.proxyResp["user_agent"].(string)}, nil
}
