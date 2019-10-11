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

// LogIn logs in into an osu! account and returns a Client.
func LogIn(username, password, fckcfAddr string) (*Client, error) {
	fckCfReqJson := []byte(`{"url":"https://old.ppy.sh/forum/ucp.php?mode=login"}`)
	fckCfResp, err := http.Post(fckcfAddr, "application/json", bytes.NewBuffer(fckCfReqJson))
	if err != nil {
		return nil, err
	}
	fckCfBody, err := ioutil.ReadAll(fckCfResp.Body)
	if err != nil {
		return nil, err
	}
	var fckCfData map[string]interface{}
	json.Unmarshal(fckCfBody, &fckCfData)
	if err != nil {
		return nil, err
	}

	j, err := cookiejar.New(&cookiejar.Options{})
	if err != nil {
		return nil, err
	}
	u, _ := url.Parse("https://osu.ppy.sh")
	var cookies []*http.Cookie
	cookie := &http.Cookie{
		Name: "__cfduid",
		Value: fckCfData["__cfduid"].(string),
		Path: "/",
		Domain: ".ppy.sh",
	}
	cookies = append(cookies, cookie)
	cookie = &http.Cookie{
		Name: "cf_clearance",
			Value: fckCfData["cf_clearance"].(string),
			Path: "/",
			Domain: ".ppy.sh",
	}
	cookies = append(cookies, cookie)
	j.SetCookies(u, cookies)
	c := &http.Client{
		Jar: j,
	}
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
	req.Header.Set("User-Agent", fckCfData["user_agent"].(string))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Content-Length", strconv.Itoa(len(vals.Encode())))
	// loginResp, err := client.PostForm("https://osu.ppy.sh/forum/ucp.php?mode=login", vals)
	loginResp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	/*defer loginResp.Body.Close()
	body, err := ioutil.ReadAll(loginResp.Body)
	if err != nil {
		return nil, errors.New("cheesegull/downloader: could not read login response body")
	}
	fmt.Println(string(body))
	fmt.Println(j.Cookies(u))*/
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

