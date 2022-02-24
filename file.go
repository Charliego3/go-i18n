package i18n

import (
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"io/fs"
	"io/ioutil"
	"path/filepath"
)

// fileLoader load messages from local file
type fileLoader struct {
	*loaderCfg
}

func NewFileLoader(opts ...LoaderOp) *fileLoader {
	loader := &fileLoader{&loaderCfg{}}
	loader.apply(opts...)
	return loader
}

func (b *fileLoader) Walk(fn func(path, ext string, lang language.Tag) error) error {
	return filepath.Walk(b.getRootPath(), func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() || b.PathFilter.Filter(path) {
			return nil
		}

		tag, ext, err := getLanguage(info.Name())
		if err != nil {
			return err
		}

		if e := fn(path, ext, tag); e != nil {
			return e
		}
		return err
	})
}

func (b *fileLoader) GetRootPath() string {
	return b.RootPath
}

func (b *fileLoader) GetUnmarshalers() map[string]i18n.UnmarshalFunc {
	return b.Unmarshalers
}

func (b *fileLoader) LoadMessage(path string) ([]byte, error) {
	return ioutil.ReadFile(path)
}
