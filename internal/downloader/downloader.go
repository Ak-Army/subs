package downloader

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Ak-Army/xlog"
	"github.com/mholt/archiver/v3"

	"github.com/Ak-Army/subs/config"
	"github.com/Ak-Army/subs/internal/fileparser"
)

var errSubFound = errors.New("subtitle found")

type Downloader interface {
	Download(sp *fileparser.SeriesParams) error
	CheckForDownloaded(sp *fileparser.SeriesParams) bool
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

func (b *BaseDownloader) DownloadFile(href string, sp *fileparser.SeriesParams) error {
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
		return fmt.Errorf("wrong response code: %d", res.StatusCode)
	}
	ext := strings.Split(filepath.Ext(href), "&")[0]
	if ext == ".rar" || ext == ".zip" {
		arch := b.replaceExtension(sp.Path, ext)
		out, err := os.Create(arch)
		if err != nil {
			return err
		}
		defer os.Remove(arch)
		_, err = io.Copy(out, res.Body)
		if err != nil {
			return err
		}
		return b.deCompress(arch, sp)
	} else {
		path := b.replaceExtension(sp.Path, "."+b.Config.LanguageSub+".srt")
		out, err := os.Create(path)
		if err != nil {
			return err
		}
		_, err = io.Copy(out, res.Body)
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *BaseDownloader) CheckForDownloaded(sp *fileparser.SeriesParams) bool {
	f, err := os.Stat(sp.BasePath + b.DecompSuffix)
	if err != nil {
		return false
	}
	if !f.IsDir() {
		return false
	}
	if err := filepath.Walk(sp.BasePath+b.DecompSuffix, func(p string, f os.FileInfo, err error) error {
		b.Logger.Debug(p, " - ", sp.Path)
		if filepath.Ext(f.Name()) == ".srt" {
			if b.Find("[E|x]"+sp.EpisodeNumber+".*"+sp.ExtraInfo+".*"+sp.ReleaseGroup, p) {
				if err := os.Rename(p, b.GetSrtPath(sp.Path)); err != nil {
					return err
				}
				return errSubFound
			}
		}
		return nil
	}); err == errSubFound {
		return true
	}
	return false
}

func (b *BaseDownloader) GetSrtPath(path string) string {
	return b.replaceExtension(path, "."+b.Config.LanguageSub+".srt")
}

func (b *BaseDownloader) Find(match string, s string) bool {
	re, err := regexp.MatchString(strings.ToLower(match), strings.ToLower(s))
	if err != nil {
		return false
	}
	if re {
		return true
	}
	return false
}

func (b *BaseDownloader) replaceExtension(path string, ext string) string {
	return strings.TrimSuffix(path, filepath.Ext(path)) + ext
}

func (b *BaseDownloader) deCompress(source string, sp *fileparser.SeriesParams) error {
	destDir := sp.Path + b.DecompSuffix
	if b.Config.Season {
		destDir = sp.BasePath + b.DecompSuffix
	}
	ext := filepath.Ext(source)
	if ext == ".zip" {
		z := archiver.Zip{
			OverwriteExisting: true,
		}
		defer z.Close()
		if err := z.Unarchive(source, destDir); err != nil {
			return err
		}
	} else if ext == ".rar" {
		r := archiver.Rar{
			OverwriteExisting: true,
		}
		defer r.Close()
		if err := r.Unarchive(source, destDir); err != nil {
			return err
		}
	} else {
		b.Logger.Error("Unsupported compressed file ", source)
		return nil
	}
	if !b.Config.Season {
		filepath.Walk(destDir, func(p string, f os.FileInfo, err error) error {
			b.Logger.Debug(p, " - ", sp.Path)
			if filepath.Ext(f.Name()) == ".srt" {
				return os.Rename(p, b.GetSrtPath(sp.Path))
			}
			return nil
		})
		return os.RemoveAll(destDir)
	}
	return nil
}
