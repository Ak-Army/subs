package downloader

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/Ak-Army/subs/internal"
	"github.com/Ak-Army/xlog"
	"github.com/PuerkitoBio/goquery"
)

const Url string = "https://www.feliratok.info"

type Feliratok struct {
	*internal.SeriesParams
	Logger xlog.Logger
}

func (f *Feliratok) Download(lang string) error {
	req, err := http.NewRequest("GET", Url, nil)
	if err != nil {
		return err
	}

	q := req.URL.Query()
	q.Add("search", fmt.Sprintf("%s %s %sx%s %s %s", f.Name, f.Year, f.SeasonNumber, f.EpisodeNumber, f.ExtraInfo, f.ReleaseGroup))
	q.Add("nyelv", lang)
	req.URL.RawQuery = q.Encode()

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

	/*b,_ := ioutil.ReadAll(res.Body)
	f.Logger.Debug("response: ", string(b))*/

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return err
	}
	// Find the review items
	doc.Find("table.result tr").Each(func(i int, tr *goquery.Selection) {
		tr.Find("td:nth-child(6) a").Each(func(a int, link *goquery.Selection) {
			if href, ok := link.Attr("href"); ok {
				f.downloadFile(href, res)
			}
		})
	})
	return nil
}

func (f *Feliratok) downloadFile(href string, resHeaders *http.Response) {
	base, err := url.ParseQuery(href)
	if err != nil {
		return
	}
	f.Logger.Debug("href: ", href, " fnev: ", base.Get("fnev"))
	req, err := http.NewRequest("GET", Url+href, nil)
	if err != nil {
		return
	}
	req.Header = resHeaders.Header
	f.Logger.Info("Download file: ", req.URL.String())
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 400 {
		return
	}
	out, err := os.Create(f.Path + "/" + base.Get("fnev"))
	_, err = io.Copy(out, res.Body)
	if err != nil {
		return
	}
}
