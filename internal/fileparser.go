package internal

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"regexp"

	"github.com/Ak-Army/xlog"
)

type FileParser struct {
	FilenamePatterns             []string
	SeriesnameYearPattern        string
	ExtraInfoPattern             string
	SeriesnameReplacements       []*Replacements
	ReleasegroupInfoReplacements []*Replacements
	ExtraInfoReplacements        []*Replacements
	ExtensionPattern             string
	EpisodeNumber                string
}

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

func (fp FileParser) Parse(files []string, log xlog.Logger) {
	var patterns []*regexp.Regexp
	for _, p := range fp.FilenamePatterns {
		re, err := regexp.Compile(p)
		if err != nil {
			log.Error("Invalid filename pattern ", p, err)
		}
		patterns = append(patterns, re)

	}
	for _, f := range files {
		for _, p := range patterns {
			if namedGroups := fp.matchWithGroup(p, f); len(namedGroups) > 0 {
				log.Info("# Processing file: ", f)
				episodeNumber := fp.episodeNumber(namedGroups)
				if episodeNumber == "" {
					log.Warn("# Regex does not contain episode number group, should"+
						"contain episodenumber, episodenumber1-9, or"+
						"episodenumberstart and episodenumberend# Pattern"+
						"was: ", p.String())
					break
				}

				seriesName, year := fp.seriesName(namedGroups)
				if seriesName == "" {
					log.Warn("# # Regex must contain seriesname. Pattern was: ", p.String())
					break
				}
				seasonNumber := fp.seasonNumber(namedGroups)
				extraInfo := fp.extraInfo(namedGroups)
				releaseGroup := fp.releaseGroup(namedGroups)
				if realName, err := fp.checkTvMaze(seriesName); err != nil {
					log.Info("Start set sub: ", realName, extraInfo, releaseGroup)
					// subtitle_search(result, year, seasonnumber, episodenumbers, extrainfo, releasegroup, re.sub(Config["extension_pattern"], "", filename), onlypath, fullpath)
				} else if realName, err := fp.checkTvDB(seriesName); err != nil {
					log.Info("Start set sub: ", realName, extraInfo, releaseGroup)
					// subtitle_search(result, year, seasonnumber, episodenumbers, extrainfo, releasegroup, re.sub(Config["extension_pattern"], "", filename), onlypath, fullpath)
				} else {
					log.Infof("# Not found on www.tvmaze.com and thetvdb.com: %s %s %sx%s", seriesName, year, seasonNumber, episodeNumber)
				}
				return
			}
		}
	}
}

func (fp FileParser) episodeNumber(namedgroups map[string]string) string {
	if v, ok := namedgroups["episodenumberstart"]; ok {
		// Multiple episodes, regex specifies start and end number
		return fmt.Sprint(fp.EpisodeNumber+"-"+fp.EpisodeNumber, v, namedgroups["episodenumberend"])
	} else if v, ok := namedgroups["episodenumber"]; ok {
		return fmt.Sprint(fp.EpisodeNumber, v)
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
	if matchYear := fp.matchWithGroup(re, s); len(matchYear) > 0 {
		s = matchYear["extra"]
	}
	for _, r := range fp.ExtraInfoReplacements {
		s = r.Replace(s)
	}
	return s
}

func (fp FileParser) checkTvMaze(name string) (string, error) {
	url.QueryEscape(name)
	resp, err := http.DefaultClient.Get(QuickUrlTvmaze + url.QueryEscape(name))
	if err != nil {
		return "", err
	}
	result := &struct {
		RealName string `json:"name"`
	}{}
	jsonReader := json.NewDecoder(resp.Body)
	if err := jsonReader.Decode(result); err != nil {
		return "", err
	}
	return result.RealName, nil
}
func (fp FileParser) checkTvDB(name string) (string, error) {
	url.QueryEscape(name)
	resp, err := http.DefaultClient.Get(QuickUrlTvdb + url.QueryEscape(name))
	if err != nil {
		return "", err
	}
	result := &struct {
		RealName string `xml:"Data>Series>SeriesName"`
	}{}
	jsonReader := xml.NewDecoder(resp.Body)
	if err := jsonReader.Decode(result); err != nil {
		return "", err
	}
	return result.RealName, nil
}
