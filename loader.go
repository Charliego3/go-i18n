package i18n

import (
	"fmt"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"path/filepath"
	"strings"
)

type Loader interface {
	GetUnmarshalers() map[string]i18n.UnmarshalFunc
	GetRootPath() string
	Walk(func(path, ext string, lang language.Tag) error) error
	LoadMessage(string) ([]byte, error)
}

type PathFilter interface{ Filter(path string) bool }
type PathFilterFunc func(path string) bool
type LoaderOp interface{ Apply(cfg *loaderCfg) }
type LoaderOpFunc func(cfg *loaderCfg)
type rootPathOp string
type unmarshalls map[string]i18n.UnmarshalFunc
type filter struct{ PathFilter }
type unmarshal struct {
	format string
	fn     i18n.UnmarshalFunc
}

func (f filter) Apply(cfg *loaderCfg)            { cfg.PathFilter = f.PathFilter }
func (u unmarshalls) Apply(cfg *loaderCfg)       { cfg.Unmarshalers = u }
func (r rootPathOp) Apply(cfg *loaderCfg)        { cfg.RootPath = string(r) }
func (c LoaderOpFunc) Apply(cfg *loaderCfg)      { c(cfg) }
func (f PathFilterFunc) Filter(path string) bool { return f(path) }
func (u unmarshal) Apply(cfg *loaderCfg) {
	if cfg.Unmarshalers == nil {
		cfg.Unmarshalers = make(map[string]i18n.UnmarshalFunc)
	}
	cfg.Unmarshalers[u.format] = u.fn
}

func LoaderRootPath(path string) LoaderOp                           { return rootPathOp(path) }
func LoaderUnmarshalls(fns map[string]i18n.UnmarshalFunc) LoaderOp  { return unmarshalls(fns) }
func LoaderFilter(f PathFilter) LoaderOp                            { return filter{f} }
func LoaderUnmarshal(format string, fn i18n.UnmarshalFunc) LoaderOp { return unmarshal{format, fn} }

type loaderCfg struct {
	RootPath     string
	Unmarshalers map[string]i18n.UnmarshalFunc
	PathFilter   PathFilter
}

var defaultFilter = PathFilterFunc(func(path string) bool {
	return len(filepath.Ext(path)) == 0
})

func (c *loaderCfg) apply(opts ...LoaderOp) {
	for _, opt := range opts {
		opt.Apply(c)
	}
	if c.PathFilter == nil {
		c.PathFilter = defaultFilter
	}
}

func (c *loaderCfg) getRootPath() string {
	root := c.RootPath
	if len(root) == 0 {
		root = "."
	}
	return root
}

func getLanguage(path string) (language.Tag, string, error) {
	lans := strings.SplitN(path, ".", 3)
	llen := len(lans)
	if llen < 2 {
		return language.Tag{}, "", fmt.Errorf("file type err: %s", path)
	}

	tag, err := language.Parse(lans[llen-2])
	if err != nil {
		return language.Tag{}, "", err
	}
	return tag, lans[llen-1], nil
}
