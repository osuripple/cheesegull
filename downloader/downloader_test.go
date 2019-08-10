package downloader

import (
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"testing"
)

var c *Client

var (
	username = os.Getenv("OSU_USERNAME")
	password = os.Getenv("OSU_PASSWORD")
)

func TestLogIn(t *testing.T) {
	var err error
	c, err = LogIn(username, password)
	if err != nil {
		t.Fatal(err)
	}
}

func TestLogInWrongDetails(t *testing.T) {
	_, err := LogIn("a", "i")
	if err == nil {
		t.Fatal("Unexpected non-error when trying to log in with user 'a' and password 'i'")
	}
}

func TestDownload(t *testing.T) {
	if c == nil {
		t.Skip("c is nil")
	}
	{
		vid, err := c.Download(1, false)
		if err != nil {
			t.Fatal(err)
		}
		md5Test(t, vid, "f40fae62893087e72672b3e6d1468a70")
	}
	{
		vid, err := c.Download(100517, false)
		if err != nil {
			t.Fatal(err)
		}
		md5Test(t, vid, "500b361f47ff99551dbb9931cdf39ace")
	}
	{
		novid, err := c.Download(100517, true)
		if err != nil {
			t.Fatal(err)
		}
		md5Test(t, novid, "3de1e07850e2fe1f21333e4d5b01a350")
	}
}

func cleanUp(files ...string) {
	for _, f := range files {
		os.Remove(f)	
	}
}

func md5Test(t *testing.T, f io.Reader, expect string) {
	data, err := ioutil.ReadAll(f)
	if err != nil {
		t.Fatal(err)
	}
	sum := fmt.Sprintf("%x", md5.Sum(data))
	if sum != expect {
		t.Fatal("expecting md5 sum to be", expect, "got", sum)
	}
}
