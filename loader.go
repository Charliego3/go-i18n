package i18n

import (
	"fmt"
	"golang.org/x/text/language"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type Loader interface {
	ParseMessage(i *I18n) error
}

type LoaderOp interface{ apply(cfg *LoaderConfig) }
type LoaderOpFunc func(cfg *LoaderConfig)
type fss struct{ fs fs.FS }
type unmarshalls map[string]UnmarshalFunc
type unmarshal struct {
	format string
	fn     UnmarshalFunc
}

func (f fss) apply(cfg *LoaderConfig)          { cfg.fs = f.fs }
func (u unmarshalls) apply(cfg *LoaderConfig)  { cfg.ums = u }
func (c LoaderOpFunc) apply(cfg *LoaderConfig) { c(cfg) }
func (u unmarshal) apply(cfg *LoaderConfig) {
	if cfg.ums == nil {
		cfg.ums = make(map[string]UnmarshalFunc)
	}
	cfg.ums[u.format] = u.fn
}

func WithUnmarshalls(fns map[string]UnmarshalFunc) LoaderOp  { return unmarshalls(fns) }
func WithUnmarshal(format string, fn UnmarshalFunc) LoaderOp { return unmarshal{format, fn} }

func NewLoaderWithPath(path string, opts ...LoaderOp) Loader {
	loader := &LoaderConfig{}
	opts = append(opts, fss{os.DirFS(path)})
	for _, opt := range opts {
		opt.apply(loader)
	}
	return loader
}

func NewLoaderWithFS(fs fs.FS, opts ...LoaderOp) Loader {
	loader := &LoaderConfig{}
	opts = append(opts, fss{fs})
	for _, opt := range opts {
		opt.apply(loader)
	}
	return loader
}

type LoaderConfig struct {
	fs  fs.FS
	ums map[string]UnmarshalFunc
}

func (c *LoaderConfig) ParseMessage(i *I18n) error {
	for format, ufn := range c.ums {
		i.RegisterUnmarshalFunc(format, ufn)
	}

	return c.parseMessage(i, ".")
}

func (c *LoaderConfig) parse(name string, buf []byte) error {
	ns := strings.Split(name, ".")
	if len(name) == 0 || len(ns) < 2 {
		return fmt.Errorf("the file %s not ext", name)
	}

	format := ns[1]
	if _, ok := c.ums[format]; !ok {
		i.registerUnmarshalFunc(format)
	}

	tag, err := language.Parse(ns[0])
	if err != nil {
		return err
	}
	i.SetLocalizer(tag)
	i.MastParseMessageFileBytes(buf, name)
	return nil
}

func (c *LoaderConfig) parseMessage(i *I18n, path string) error {
	entries, err := fs.ReadDir(c.fs, path)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		name := entry.Name()
		fp := filepath.Join(path, name)
		if entry.IsDir() {
			err := c.parseMessage(i, fp)
			if err != nil {
				return err
			}
		} else {
			buf, err := fs.ReadFile(c.fs, fp)
			if err != nil {
				return err
			}
			err = c.parse(name, buf)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
