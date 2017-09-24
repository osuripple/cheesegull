// Package downloader implements downloading from the osu! website, through,
// well, mostly scraping and dirty hacks.
package downloader

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
)

// LogIn logs in into an osu! account and returns a Client.
func LogIn(username, password string) (*Client, error) {
	j, err := cookiejar.New(&cookiejar.Options{})
	if err != nil {
		return nil, err
	}
	c := &http.Client{
		Jar: j,
	}
	vals := url.Values{}
	vals.Add("redirect", "/")
	vals.Add("sid", "")
	vals.Add("username", username)
	vals.Add("password", password)
	vals.Add("autologin", "on")
	vals.Add("login", "login")
	loginResp, err := c.PostForm("https://osu.ppy.sh/forum/ucp.php?mode=login", vals)
	if err != nil {
		return nil, err
	}
	if loginResp.Request.URL.Path != "/" {
		return nil, errors.New("downloader: Login: could not log in (was not redirected to index)")
	}
	return (*Client)(c), nil
}

// Client is a wrapper around an http.Client which can fetch beatmaps from the
// osu! website.
type Client http.Client

// HasVideo checks whether a beatmap has a video.
func (c *Client) HasVideo(setID int) (bool, error) {
	h := (*http.Client)(c)

	page, err := h.Get(fmt.Sprintf("https://osu.ppy.sh/s/%d", setID))
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

// Download downloads a beatmap from the osu! website.
// First reader is beatmap with video.
// Second reader is beatmap without video.
// If video is not in the beatmap, second reader will be nil and first reader
// will be beatmap without video.
func (c *Client) Download(setID int) (io.ReadCloser, io.ReadCloser, error) {
	hasVideo, err := c.HasVideo(setID)
	if err != nil {
		return nil, nil, err
	}

	if hasVideo {
		r1, err := c.getReader(strconv.Itoa(setID))
		if err != nil {
			return nil, nil, err
		}
		r2, err := c.getReader(strconv.Itoa(setID) + "n")
		if err != nil {
			// If r2 fails that is OK, however r1 must be closed, otherwise connections leak
			r1.Close()
			return nil, nil, err
		}
		return r1, r2, nil
	}

	r, err := c.getReader(strconv.Itoa(setID))
	return r, nil, err
}

// ErrNoRedirect is returned from Download when we were not redirect, thus
// indicating that the beatmap is unavailable.
var ErrNoRedirect = errors.New("no redirect happened, beatmap could not be downloaded")

func (c *Client) getReader(str string) (io.ReadCloser, error) {
	h := (*http.Client)(c)

	resp, err := h.Get("https://osu.ppy.sh/d/" + str)
	if err != nil {
		return nil, err
	}
	if resp.Request.URL.Host == "osu.ppy.sh" {
		resp.Body.Close()
		return nil, ErrNoRedirect
	}

	return resp.Body, nil
}
