package fileparser

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"regexp"

	"github.com/Ak-Army/subs/config"
	"github.com/Ak-Army/xlog"
)

const QuickUrlTvmaze string = "http://api.tvmaze.com/singlesearch/shows?q="
const QuickUrlTvdb string = "http://thetvdb.com/api/GetSeries.php?seriesname="

var cleanupRegex map[*regexp.Regexp]string

func init() {
	cleanupRegex = make(map[*regexp.Regexp]string)
	for k, v := range map[string]string{
		`(\D)[.](\D)`: "$1 $2",
		`(\D)[.]`:     "$1 ",
		`[.](\D)`:     " $1",
		"_":           " ",
		"-$":          "",
	} {
		re, err := regexp.Compile(k)
		if err != nil {
			xlog.Error("Invalid regexp pattern ", k, err)
		}
		cleanupRegex[re] = v
	}
}

type FileParser struct {
	*config.Config
	Logger xlog.Logger
}

func (fp FileParser) Parse(files []string, basePath string) []*SeriesParams {
	var ret []*SeriesParams
	var patterns []*regexp.Regexp
	for _, p := range fp.FilenamePatterns {
		re, err := regexp.Compile(p)
		if err != nil {
			fp.Logger.Error("Invalid filename pattern ", p, err)
		}
		patterns = append(patterns, re)

	}
	for _, path := range files {
		f := filepath.Base(path)
		for _, p := range patterns {
			if namedGroups := fp.matchWithGroup(p, f); len(namedGroups) > 0 {
				fp.Logger.Infof("Processing file: %s", f)
				episodeNumber := fp.episodeNumber(namedGroups)
				if episodeNumber == "" {
					fp.Logger.Warn("Regex does not contain episode number group, should"+
						"contain episodenumber, episodenumber1-9, or"+
						"episodenumberstart and episodenumberend Pattern"+
						"was: ", p.String())
					break
				}

				seriesName, year := fp.seriesName(namedGroups)
				if seriesName == "" {
					fp.Logger.Warnf("Regex must contain seriesname. Pattern was: %s", p.String())
					break
				}
				seasonNumber := fp.seasonNumber(namedGroups)
				extraInfo := fp.extraInfo(namedGroups)
				releaseGroup := fp.releaseGroup(namedGroups)
				sp := &SeriesParams{
					BasePath:      basePath,
					Path:          filepath.Dir(path),
					SeasonNumber:  seasonNumber,
					EpisodeNumber: episodeNumber,
					ExtraInfo:     extraInfo,
					ReleaseGroup:  releaseGroup,
					Year:          year,
				}
				if realName, err := fp.checkTvMaze(seriesName); err == nil {
					fp.Logger.Infof("Found on tvmaze.com: %s %sx%s %s %s", realName, seasonNumber, episodeNumber, extraInfo, releaseGroup)
					sp.Name = realName
					ret = append(ret, sp)
				} else if realName, err := fp.checkTvDB(seriesName); err == nil {
					fp.Logger.Infof("Found on thetvdb.com: %s %sx%s %s %s", realName, seasonNumber, episodeNumber, extraInfo, releaseGroup)
					sp.Name = realName
					ret = append(ret, sp)
				} else {
					fp.Logger.Infof("Not found on www.tvmaze.com and thetvdb.com:  %sx%s %s %s", seasonNumber, episodeNumber, extraInfo, releaseGroup)
				}
				break
			}
		}
	}
	return ret
}

func (fp FileParser) episodeNumber(namedgroups map[string]string) string {
	if v, ok := namedgroups["episodenumberstart"]; ok {
		// Multiple episodes, regex specifies start and end number
		return fmt.Sprintf(fp.EpisodeNumber+"-"+fp.EpisodeNumber, v, namedgroups["episodenumberend"])
	} else if v, ok := namedgroups["episodenumber"]; ok {
		return fmt.Sprintf(fp.EpisodeNumber, v)
	}
	return ""
}

func (fp FileParser) seriesName(namedgroups map[string]string) (string, string) {
	s, ok := namedgroups["seriesname"]
	if !ok {
		return "", ""
	}
	for re, v := range cleanupRegex {
		s = re.ReplaceAllString(s, v)
	}
	re, err := regexp.Compile(fp.SeriesnameYearPattern)
	if err != nil {
		return "", ""
	}
	year := ""
	if matchYear := fp.matchWithGroup(re, s); len(matchYear) > 0 {
		s = matchYear["seriesname"]
		year = matchYear["year"]
	}
	for _, r := range fp.SeriesnameReplacements {
		s = r.Replace(s)
	}
	return s, year
}

func (fp FileParser) matchWithGroup(r *regexp.Regexp, s string) map[string]string {
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

func (fp FileParser) seasonNumber(namedgroups map[string]string) string {
	s, ok := namedgroups["seasonnumber"]
	if !ok {
		return ""
	}
	return s
}

func (fp FileParser) releaseGroup(namedgroups map[string]string) string {
	s, ok := namedgroups["releasegroup"]
	if !ok {
		return ""
	}
	if len(fp.ReleasegroupInfoReplacements) > 0 {
		for _, r := range fp.ReleasegroupInfoReplacements {
			s = r.Replace(s)
		}
	}
	return s
}

func (fp FileParser) extraInfo(namedgroups map[string]string) string {
	s, ok := namedgroups["extrainfo"]
	if !ok {
		return ""
	}
	for re, v := range cleanupRegex {
		s = re.ReplaceAllString(s, v)
	}
	re, err := regexp.Compile(fp.ExtraInfoPattern)
	if err != nil {
		return ""
	}
	if matchExtra := fp.matchWithGroup(re, s); len(matchExtra) > 0 {
		s = matchExtra["extra"]
	}
	for _, r := range fp.ExtraInfoReplacements {
		s = r.Replace(s)
	}
	return s
}

func (fp FileParser) checkTvMaze(name string) (string, error) {
	fp.Logger.Debug("checkTvMaze ", QuickUrlTvmaze+url.QueryEscape(name))
	res, err := http.DefaultClient.Get(QuickUrlTvmaze + url.QueryEscape(name))
	if err != nil {
		return "", err
	}
	if res.StatusCode < 200 || res.StatusCode >= 400 {
		return "", fmt.Errorf("not found")
	}
	fp.Logger.Debug("response", res.Body)
	result := &struct {
		RealName string `json:"name"`
	}{}
	jsonReader := json.NewDecoder(res.Body)
	if err := jsonReader.Decode(result); err != nil {
		return "", err
	}
	return result.RealName, nil
}
func (fp FileParser) checkTvDB(name string) (string, error) {
	fp.Logger.Debug("checkTvDB ", QuickUrlTvdb+url.QueryEscape(name))
	res, err := http.DefaultClient.Get(QuickUrlTvdb + url.QueryEscape(name))
	if err != nil {
		return "", err
	}
	fp.Logger.Debug("response", res.Body)
	result := &struct {
		RealName string `xml:"Data>Series>SeriesName"`
	}{}
	jsonReader := xml.NewDecoder(res.Body)
	if err := jsonReader.Decode(result); err != nil {
		return "", err
	}
	return result.RealName, nil
}
