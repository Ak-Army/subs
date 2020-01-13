package feliratok

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/Ak-Army/subs/internal"
	"github.com/Ak-Army/subs/internal/downloader"
	"github.com/PuerkitoBio/goquery"
)

const Url string = "https://www.feliratok.info"

type Feliratok struct {
	*downloader.BaseDownloader
}

func (f *Feliratok) Download(sp *internal.SeriesParams) error {
	f.Logger.Info("Start download: ", sp.Name)
	req, err := f.NewRequest("GET", Url, nil)
	if err != nil {
		return err
	}
	q := req.URL.Query()
	q.Add("search", fmt.Sprintf("%s %s %sx%s %s %s", sp.Name, sp.Year, sp.SeasonNumber, sp.EpisodeNumber, sp.ExtraInfo, sp.ReleaseGroup))
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
		if found {
			return
		}
		tr.Find("td:nth-child(6) a").Each(func(a int, link *goquery.Selection) {
			if found {
				return
			}
			if href, ok := link.Attr("href"); ok {
				base, err := url.ParseQuery(href)
				if err != nil {
					return
				}
				err = f.DownloadFile(Url+href, f.GetSrtPath(base.Get("fnev"), sp.Path))
				found = true
			}
		})
	})
	if !found {
		return errors.New("not found")
	}
	return err
}
