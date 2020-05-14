package mkv

import (
	"os"
	"time"

	"github.com/remko/go-mkvparse"
)

type parser struct {
	SubLanguage bool
}

func HaveSub(filename string) bool {
	file, err := os.Open(filename)
	if err != nil {
		return false
	}
	defer file.Close()

	handler := parser{}
	err = mkvparse.ParseSections(file, &handler, mkvparse.TracksElement)
	if err != nil {
		return false
	}
	if handler.SubLanguage {
		return true
	}
	return false
}

func (mp *parser) HandleMasterBegin(id mkvparse.ElementID, info mkvparse.ElementInfo) (bool, error) {
	return true, nil
}

func (mp *parser) HandleMasterEnd(id mkvparse.ElementID, info mkvparse.ElementInfo) error {
	return nil
}

func (mp *parser) HandleString(id mkvparse.ElementID, value string, info mkvparse.ElementInfo) error {
	if id == mkvparse.LanguageElement {
		if value == "hun" {
			mp.SubLanguage = true
		}
	}
	return nil
}

func (mp *parser) HandleInteger(id mkvparse.ElementID, value int64, info mkvparse.ElementInfo) error {
	return nil
}

func (mp *parser) HandleFloat(id mkvparse.ElementID, value float64, info mkvparse.ElementInfo) error {
	return nil
}

func (mp *parser) HandleDate(id mkvparse.ElementID, value time.Time, info mkvparse.ElementInfo) error {
	return nil
}

func (mp *parser) HandleBinary(id mkvparse.ElementID, value []byte, info mkvparse.ElementInfo) error {
	return nil
}
