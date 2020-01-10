package subiratok

import (
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/Ak-Army/subs/config"
	"github.com/Ak-Army/subs/internal"
	"github.com/Ak-Army/xlog"
)

const Url string = "http://subirat.net"

type Subiratok struct {
	*internal.SeriesParams
	*config.Config
	Logger xlog.Logger
}

func (f *Subiratok) Download(lang string) error {
	req, err := http.NewRequest("GET", Url, nil)
	if err != nil {
		return err
	}
	title := strings.ReplaceAll(
		strings.ReplaceAll(
			strings.ToLower(f.Name),
			" ", "-"),
		"'", "")
	req.URL.Path = fmt.Sprintf("/t/%s/rss", title)

	req.Header.Add("Referer", Url)
	req.Header.Add("User-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_9_3) AppleWebKit/537.75.14 (KHTML, like Gecko) Version/7.0.3 Safari/7046A194A")
	f.Logger.Info("Download: ", req.URL.String())
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 400 {
		return fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}

	b, _ := ioutil.ReadAll(res.Body)
	//f.Logger.Debug("response: ", string(b))

	// Load the HTML document
	rss := struct {
		Channel struct {
			Item []struct {
				Description string `xml:"description"`
				Title       string `xml:"title"`
				Link        string `xml:"link"`
			} `xml:"item"`
		} `xml:"channel"`
	}{}
	if err := xml.Unmarshal(b, &rss); err != nil {
		return err
	}
	f.Logger.Debug(rss)

	re, err := regexp.Compile(f.Config.SubiratPattern)
	if err != nil {
		return err
	}
	for _, item := range rss.Channel.Item {
		if match := f.matchWithGroup(re, item.Description); len(match) > 0 {
			seasonNumber, ok := match["seasonnumber"]
			if !ok {
				f.Logger.Debug("no season number")
				continue
			}
			episodeNumber, ok := match["episodenumber"]
			if !ok {
				f.Logger.Debug("no episode number")
				continue
			}

			if strings.HasSuffix(item.Title, ".srt") &&
				fmt.Sprintf(f.Config.EpisodeNumber, seasonNumber) == f.SeasonNumber &&
				fmt.Sprintf(f.Config.EpisodeNumber, episodeNumber) == f.SeriesParams.EpisodeNumber {
				f.downloadFile(item.Link)
				return nil
			}
		}
	}

	return nil
}

func (f *Subiratok) matchWithGroup(r *regexp.Regexp, s string) map[string]string {
	namedGroups := make(map[string]string)
	if match := r.FindStringSubmatch(s); len(match) > 0 {
		for i, name := range r.SubexpNames() {
			if i != 0 && name != "" {
				namedGroups[name] = match[i]
			}
		}
	}
	return namedGroups
}

func (f *Subiratok) downloadFile(href string) {
	req1, err := http.NewRequest("GET", Url, nil)
	req1.Header.Add("Referer", Url)
	req1.Header.Add("User-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_9_3) AppleWebKit/537.75.14 (KHTML, like Gecko) Version/7.0.3 Safari/7046A194A")
	resHeaders, err := http.DefaultClient.Do(req1)
	if err != nil {
		return
	}
	req, err := http.NewRequest("GET", href, nil)
	if err != nil {
		return
	}
	req.Header.Add("Referer", Url)
	req.Header.Add("User-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_9_3) AppleWebKit/537.75.14 (KHTML, like Gecko) Version/7.0.3 Safari/7046A194A")
	for _, cookie := range resHeaders.Cookies() {
		req.AddCookie(cookie)
	}
	f.Logger.Info("Download file: ", href, req.Header)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 400 {
		return
	}
	out, err := os.Create(f.Path + "/" + f.Name + ".srt")
	_, err = io.Copy(out, res.Body)
	if err != nil {
		return
	}
}
