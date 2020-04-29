package config

import (
	"os"
	"time"

	"github.com/remko/go-mkvparse"
)

type mkvParser struct {
	SubLanguage bool
}

func MkvSub(filename string) bool {
	file, err := os.Open(filename)
	defer file.Close()

	if err != nil {
		return false
	}

	handler :=  mkvParser{}
	err = mkvparse.ParseSections(file, &handler, mkvparse.TracksElement)

	if err != nil {
		return false
	}
	if handler.SubLanguage {
		return true
	}
	return false
}

func (mp *mkvParser) HandleMasterBegin(id mkvparse.ElementID, info mkvparse.ElementInfo) (bool, error) {
	return true, nil
}

func (mp *mkvParser) HandleMasterEnd(id mkvparse.ElementID, info mkvparse.ElementInfo) error {
	return nil
}

func (mp *mkvParser) HandleString(id mkvparse.ElementID, value string, info mkvparse.ElementInfo) (error) {
	if id == mkvparse.LanguageElement {
		if value == "hun" {
			mp.SubLanguage = true
		}
	}
	return nil
}

func (mp *mkvParser) HandleInteger(id mkvparse.ElementID, value int64, info mkvparse.ElementInfo) error {
	return nil
}

func (mp *mkvParser) HandleFloat(id mkvparse.ElementID, value float64, info mkvparse.ElementInfo) error {
	return nil
}

func (mp *mkvParser) HandleDate(id mkvparse.ElementID, value time.Time, info mkvparse.ElementInfo) error {
	return nil
}

func (mp *mkvParser) HandleBinary(id mkvparse.ElementID, value []byte, info mkvparse.ElementInfo) error {
	return nil
}
