package downloader

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var c *Client

var (
	username = os.Getenv("OSU_USERNAME")
	password = os.Getenv("OSU_PASSWORD")
)

func TestLogIn(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	var err error
	c, err = LogIn(username, password, &EmptyLogInRequestPreparer{})
	assert.NoError(t, err)
}

func TestLogInWrongDetails(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	_, err := LogIn("a", "i", &EmptyLogInRequestPreparer{})
	assert.NoError(t, err)
}

func TestDownload(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

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

func md5Test(t *testing.T, f io.Reader, expect string) {
	data, err := io.ReadAll(f)
	require.NoError(t, err)
	sum := fmt.Sprintf("%x", md5.Sum(data))
	require.Equal(t, expect, sum)
}
