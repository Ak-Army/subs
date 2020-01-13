package downloader

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/Ak-Army/subs/config"
	"github.com/Ak-Army/subs/internal"
	"github.com/Ak-Army/xlog"
)

type Downloader interface {
	Download(sp *internal.SeriesParams) error
}

type BaseDownloader struct {
	*config.Config
	Logger  xlog.Logger
	Lang    string
	cookies []*http.Cookie
}

func (b *BaseDownloader) NewRequest(method, url string, body io.Reader) (*http.Request, error) {
	if len(b.cookies) == 0 {
		if err := b.setCookies(url); err != nil {
			return nil, err
		}
	}
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return req, err
	}
	req.Header.Add("Referer", url)
	req.Header.Add("User-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_9_3) AppleWebKit/537.75.14 (KHTML, like Gecko) Version/7.0.3 Safari/7046A194A")
	for _, cookie := range b.cookies {
		req.AddCookie(cookie)
	}
	return req, nil
}

func (b BaseDownloader) setCookies(url string) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 400 {
		return fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}
	b.cookies = res.Cookies()
	return nil
}

func (b BaseDownloader) DownloadFile(href string, path string) {
	b.Logger.Info("Download file: ", href)
	req, err := b.NewRequest("GET", href, nil)
	if err != nil {
		return
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 400 {
		return
	}
	out, err := os.Create(path + ".srt")
	_, err = io.Copy(out, res.Body)
	if err != nil {
		return
	}
}
