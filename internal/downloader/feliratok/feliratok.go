package feliratok

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"

	"github.com/Ak-Army/subs/internal/downloader"
	"github.com/Ak-Army/subs/internal/fileparser"
)

const Url string = "https://www.feliratok.info"

type Feliratok struct {
	*downloader.BaseDownloader
}

func (f *Feliratok) Download(sp *fileparser.SeriesParams) error {
	if f.Config.Season {
		f.Logger.Info("Searching for subtitle: ", sp.Name, " Season ", strings.TrimLeft(sp.SeasonNumber, "0"))
	} else {
		f.Logger.Info("Searching for subtitle: ", sp.Name, " ", strings.TrimLeft(sp.SeasonNumber, "0"), "x", sp.EpisodeNumber, " ", sp.ExtraInfo, "-", sp.ReleaseGroup)
	}
	req, err := f.NewRequest("GET", Url, nil)
	if err != nil {
		return err
	}
	q := req.URL.Query()
	if f.Config.Season {
		q.Add("search", fmt.Sprintf("%s %s (Season %s)", sp.Name, sp.Year, strings.TrimLeft(sp.SeasonNumber, "0")))
	} else {
		q.Add("search", fmt.Sprintf("%s %s %sx%s %s %s", sp.Name, sp.Year, strings.TrimLeft(sp.SeasonNumber, "0"), sp.EpisodeNumber, sp.ExtraInfo, sp.ReleaseGroup))
	}
	q.Add("nyelv", f.Config.Language)
	req.URL.RawQuery = q.Encode()

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 400 {
		return fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}

	/*b,_ := ioutil.ReadAll(res.Body)
	f.Logger.Debug("response: ", string(b))*/

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return err
	}
	found := false
	// Find the review items
	doc.Find("table.result tr").Each(func(i int, tr *goquery.Selection) {
		if found && !f.Config.Season {
			return
		}
		tr.Find("td:nth-child(6) a").Each(func(a int, link *goquery.Selection) {
			if href, ok := link.Attr("href"); ok {
				err = f.DownloadFile(Url+strings.ReplaceAll(href, " ", "+"), sp)
				found = true
			}
		})
	})
	if !found {
		return errors.New("not found feliratok.info")
	}
	return err
}
