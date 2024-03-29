package subiratok

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	"github.com/Ak-Army/subs/internal/downloader"
	"github.com/Ak-Army/subs/internal/fileparser"
)

const Url string = "http://subirat.net"

type Subiratok struct {
	*downloader.BaseDownloader
}

func (s *Subiratok) Download(sp *fileparser.SeriesParams) error {
	s.Logger.Info("Searching for subtitle: ", sp.Name, " ", sp.SeasonNumber, "x", sp.EpisodeNumber, " ", sp.ExtraInfo, "-", sp.ReleaseGroup)
	req, err := s.NewRequest("GET", Url, nil)
	if err != nil {
		return err
	}
	title := strings.ReplaceAll(
                strings.ReplaceAll(
		strings.ReplaceAll(
			strings.ToLower(sp.Name),
			" ", "-"),
		"'", ""),
                ":", "")
	req.URL.Path = fmt.Sprintf("/t/%s/rss", title)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 400 {
		return fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}

	b, _ := ioutil.ReadAll(res.Body)
	// s.Logger.Debug("response: ", string(b))

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

	re, err := regexp.Compile(s.Config.SubiratPattern)
	if err != nil {
		return err
	}
	for _, item := range rss.Channel.Item {
		if match := s.matchWithGroup(re, item.Description); len(match) > 0 {
			seasonNumber, ok := match["seasonnumber"]
			if !ok {
				s.Logger.Debug("No season number")
				continue
			}
			episodeNumber, ok := match["episodenumber"]
			if !ok {
				s.Logger.Debug("No episode number")
				continue
			}

			if strings.HasSuffix(item.Title, ".srt") &&
				fmt.Sprintf(s.Config.EpisodeNumber, seasonNumber) == sp.SeasonNumber &&
				fmt.Sprintf(s.Config.EpisodeNumber, episodeNumber) == sp.EpisodeNumber &&
				strings.Contains(strings.ToLower(item.Description), strings.ToLower(sp.ReleaseGroup)) {
				return s.DownloadFile(item.Link, sp)
			}
		}
	}

	return errors.New("not found subirat.net")
}

func (s *Subiratok) matchWithGroup(r *regexp.Regexp, st string) map[string]string {
	namedGroups := make(map[string]string)
	if match := r.FindStringSubmatch(st); len(match) > 0 {
		for i, name := range r.SubexpNames() {
			if i != 0 && name != "" {
				namedGroups[name] = match[i]
			}
		}
	}
	return namedGroups
}
