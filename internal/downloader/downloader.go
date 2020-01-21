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
		return fmt.Errorf("Status code error: %d %s", res.StatusCode, res.Status)
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
		return fmt.Errorf("Wrong response code:  %d", res.StatusCode)
	}
	ext := filepath.Ext(href)
	if ext == ".rar" || ext == ".zip" {
		zip := b.replaceExtension(path, ".zip")
		b.Logger.Info(zip, " -> ", path)
		out, err := os.Create(zip)
		_, err = io.Copy(out, res.Body)
		if err != nil {
			return err
		}
		return b.deCompress(zip, path)
	} else {
		out, err := os.Create(path)
		_, err = io.Copy(out, res.Body)
		if err != nil {
			return err
		}
	}
	return nil
}

func (b BaseDownloader) GetSrtPath(filename string, path string) string {
	ext := filepath.Ext(filename)
	return b.replaceExtension(path, "."+b.Config.LanguageSub+ext)
}

func (b BaseDownloader) replaceExtension(path string, ext string) string {
	return strings.TrimSuffix(path, filepath.Ext(path)) + ext
}

func (b BaseDownloader) deCompress(source string, destination string) error {
	destDir := destination + "_decomp"
	if err := archiver.Unarchive(source, destDir); err != nil {
		return err
	}
	defer os.Remove(source)

	filepath.Walk(destDir, func(p string, f os.FileInfo, err error) error {
		b.Logger.Debug(p, " - ", destDir)
		if f.IsDir() && p != destDir {
			return filepath.SkipDir
		}
		if filepath.Ext(f.Name()) == ".srt" {
			return os.Rename(p, destination)
		}

		return nil
	})
	return os.RemoveAll(destDir)
}
