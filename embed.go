package i18n

import (
	"embed"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"path/filepath"
)

// embedLoader load messages from embed.FS
type embedLoader struct {
	*loaderCfg
	FS embed.FS
}

func NewEmbedLoader(fs embed.FS, opts ...LoaderOp) *embedLoader {
	loader := &embedLoader{FS: fs, loaderCfg: &loaderCfg{}}
	loader.apply(opts...)
	return loader
}

func (b *embedLoader) Walk(fn func(path, ext string, lang language.Tag) error) error {
	return b.readDir(b.getRootPath(), fn)
}

func (b *embedLoader) readDir(dirAddress string, fn func(path, ext string, lang language.Tag) error) error {
	dirs, err := b.FS.ReadDir(dirAddress)
	if err != nil {
		return err
	}

	for _, dir := range dirs {
		info, err := dir.Info()
		if err != nil {
			return err
		}

		path := filepath.Join(dirAddress, info.Name())
		if info.IsDir() {
			err := b.readDir(path, fn)
			if err != nil {
				return err
			}
		} else {
			if b.PathFilter.Filter(path) {
				return nil
			}

			tag, ext, err := getLanguage(info.Name())
			if err != nil {
				return err
			}

			if e := fn(path, ext, tag); e != nil {
				return e
			}
		}
	}

	return nil
}

func (b *embedLoader) GetRootPath() string {
	return b.RootPath
}

func (b *embedLoader) GetUnmarshalers() map[string]i18n.UnmarshalFunc {
	return b.Unmarshalers
}

func (b *embedLoader) LoadMessage(path string) ([]byte, error) {
	return b.FS.ReadFile(path)
}
