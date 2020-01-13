package hosszupuska

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/Ak-Army/subs/internal"
	"github.com/Ak-Army/subs/internal/downloader"
	"github.com/PuerkitoBio/goquery"
)

const Url string = "http://hosszupuskasub.com"

type Hosszupuska struct {
	*downloader.BaseDownloader
}

func (h *Hosszupuska) Download(sp *internal.SeriesParams) error {
	h.Logger.Info("Download: ", sp.Name)
	req, err := h.NewRequest("GET", Url+"/sorozatok.php", nil)
	if err != nil {
		return err
	}
	syear := ""
	if sp.Year != "" {
		syear = fmt.Sprintf(" (%s)", sp.Year)
	}
	q := req.URL.Query()
	q.Add("cim", fmt.Sprintf("%s%s", sp.Name, syear))
	q.Add("evad", fmt.Sprintf("s%s", sp.SeasonNumber))
	q.Add("resz", fmt.Sprintf("e%s", sp.EpisodeNumber))
	q.Add("nyelvtipus", h.LanguageNumber)
	q.Add("x", "13")
	q.Add("y", "14")
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
	doc.Find("#stranka center table td:nth-child(2) table tr").Each(func(i int, tr *goquery.Selection) {
		if i == 0 || tr.Find("td").Length() < 7 || found {
			return
		}
		td := tr.Find("td").Slice(1, 2)
		name := td.Text()
		h.Logger.Debug(td.Text(), fmt.Sprintf("%s-%s", strings.ToLower(sp.ExtraInfo), strings.ToLower(sp.ReleaseGroup)))
		if strings.Contains(name, sp.Name) &&
			strings.Contains(name, fmt.Sprintf("s%se%s", sp.SeasonNumber, sp.EpisodeNumber)) &&
			strings.Contains(strings.ToLower(name), fmt.Sprintf("%s-%s", strings.ToLower(sp.ExtraInfo), strings.ToLower(sp.ReleaseGroup))) {
			link := tr.Find("td").Slice(6, 7).Find("a[target=\"_parent\"]")
			if href, ok := link.Attr("href"); ok {
				err = h.DownloadFile(href, sp.Path)
				found = true
				return
			}
		}
	})
	if !found {
		return errors.New("not found")
	}
	return err
}
