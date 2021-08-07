// Package downloader implements downloading from the osu! website, through,
// well, mostly scraping and dirty hacks.
package downloader

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
)

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

func ppyCookie(name, value string) *http.Cookie {
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
		ppyCookie("__cfduid", c.proxyResp["__cfduid"].(string)),
		ppyCookie("cf_clearance", c.proxyResp["cf_clearance"].(string)),
	})
	return j, nil
}

func (c *FckCf) PrepareHeaders() (map[string]string, error) {
	if c.proxyResp == nil {
		return nil, errors.New("must first call PrepareCookies to set the fckcf resp")
	}
	return map[string]string{"User-Agent": c.proxyResp["user_agent"].(string)}, nil
}

// LogIn logs in into an osu! account and returns a Client.
func LogIn(username string, password string, requestPreparer LogInRequestPreparer) (*Client, error) {
	// Prepare cookies
	cookieJar, err := requestPreparer.PrepareCookies()
	if err != nil {
		return nil, fmt.Errorf("prepare cookies: %w", err)
	}
	c := &http.Client{
		Jar: cookieJar,
	}

	// Prepare headers
	preparedHeaders, err := requestPreparer.PrepareHeaders()
	if err != nil {
		return nil, fmt.Errorf("prepare headers: %w", err)
	}

	// POST form values
	vals := url.Values{}
	vals.Add("redirect", "/")
	vals.Add("sid", "")
	vals.Add("username", username)
	vals.Add("password", password)
	vals.Add("autologin", "on")
	vals.Add("login", "Login")
	req, err := http.NewRequest("POST", "https://osu.ppy.sh/forum/ucp.php?mode=login", strings.NewReader(vals.Encode()))
	if err != nil {
		return nil, err
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
	loginResp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	if loginResp.Request.URL.Path != "/home" {
		return nil, errors.New("cheesegull/downloader: could not log in (was not redirected to /home) but to " + loginResp.Request.URL.Path)
	}
	return (*Client)(c), nil
}

// Client is a wrapper around an http.Client which can fetch beatmaps from the
// osu! website.
type Client http.Client

// HasVideo checks whether a beatmap has a video.
func (c *Client) HasVideo(setID int) (bool, error) {
	h := (*http.Client)(c)

	page, err := h.Get(fmt.Sprintf("https://old.ppy.sh/s/%d", setID))
	if err != nil {
		return false, err
	}
	defer page.Body.Close()
	body, err := ioutil.ReadAll(page.Body)
	if err != nil {
		return false, err
	}
	return bytes.Contains(body, []byte(fmt.Sprintf(`href="/d/%dn"`, setID))), nil
}

// Download downloads a beatmap from the osu! website. noVideo specifies whether
// we should request the beatmap to not have the video.
func (c *Client) Download(setID int, noVideo bool) (io.ReadCloser, error) {
	suffix := ""
	if noVideo {
		suffix = "n"
	}

	return c.getReader(strconv.Itoa(setID) + suffix)
}

// ErrNoRedirect is returned from Download when we were not redirect, thus
// indicating that the beatmap is unavailable.
var ErrNoRedirect = errors.New("cheesegull/downloader: no redirect happened, beatmap could not be downloaded")

var errNoZip = errors.New("cheesegull/downloader: file is not a zip archive")

const zipMagic = "PK\x03\x04"

func (c *Client) getReader(str string) (io.ReadCloser, error) {
	h := (*http.Client)(c)

	resp, err := h.Get("https://old.ppy.sh/d/" + str)
	if err != nil {
		return nil, err
	}
	if resp.Request.URL.Host == "old.ppy.sh" {
		resp.Body.Close()
		return nil, ErrNoRedirect
	}

	// check that it is a zip file
	first4 := make([]byte, 4)
	_, err = resp.Body.Read(first4)
	if err != nil {
		return nil, err
	}
	if string(first4) != zipMagic {
		return nil, errNoZip
	}

	return struct {
		io.Reader
		io.Closer
	}{
		io.MultiReader(strings.NewReader(zipMagic), resp.Body),
		resp.Body,
	}, nil
}
