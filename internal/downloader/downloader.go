package downloader

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/Ak-Army/subs/config"
	"github.com/Ak-Army/subs/internal"
	"github.com/Ak-Army/xlog"
	"github.com/mholt/archiver/v3"
)

type Downloader interface {
	Download(sp *internal.SeriesParams) error
}

type BaseDownloader struct {
	*config.Config
	Logger  xlog.Logger
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

func (b *BaseDownloader) setCookies(url string) error {
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

func (b BaseDownloader) DownloadFile(href string, path string) error {
	b.Logger.Info("Download file: ", href)
	req, err := b.NewRequest("GET", href, nil)
	if err != nil {
		return err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 400 {
		return fmt.Errorf("wrong response code:  %d", res.StatusCode)
	}
	ext := filepath.Ext(href)
	base := filepath.Base(path)
	newPath := strings.ReplaceAll(path, base, strings.TrimSuffix(base, filepath.Ext(base))+"."+b.Config.LanguageSub+ext)
	out, err := os.Create(newPath)
	_, err = io.Copy(out, res.Body)
	if err != nil {
		return err
	}
	if ext == ".rar" || ext == ".zip" {
		return b.deCompress(newPath, strings.ReplaceAll(path, base, strings.TrimSuffix(base, filepath.Ext(base))+"."+b.Config.LanguageSub))
	}
	return nil
}

func (b BaseDownloader) deCompress(source string, destination string) error {
	if err := archiver.Unarchive(source, filepath.Join(destination)); err != nil {
		return err
	}
	defer os.Remove(source)
	filepath.Walk(destination, func(p string, f os.FileInfo, err error) error {
		b.Logger.Debug(p, " - ", destination)
		if f.IsDir() && p != destination {
			return filepath.SkipDir
		}
		if filepath.Ext(f.Name()) == ".srt" {
			return os.Rename(p, destination+".srt")
		}

		return nil
	})
	return os.RemoveAll(destination)
}
