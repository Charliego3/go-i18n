package i18n

import (
	"fmt"
	"golang.org/x/text/language"
	"io/fs"
	"os"
	"path/filepath"
)

type Entry struct {
	Tag   language.Tag
	Name  string
	Bytes []byte
}

type Result struct {
	Funcs   map[string]UnmarshalFunc
	Entries []Entry
}

type Loader interface {
	Load() (*Result, error)
}

func NewLoaderWithPath(path string, opts ...LOpt) Option {
	return NewLoaderWithFS(os.DirFS(path), opts...)
}

func NewLoaderWithFS(fs fs.FS, opts ...LOpt) Option {
	l := &fsLoader{fs: fs, rs: &Result{Funcs: make(map[string]UnmarshalFunc)}}
	for _, opt := range opts {
		opt(l)
	}
	return WithLoader(l)
}

type fsLoader struct {
	fs fs.FS
	rs *Result
}

func (c *fsLoader) Load() (*Result, error) {
	err := c.load(".")
	return c.rs, err
}

func (c *fsLoader) loadFile(path string) error {
	buf, err := fs.ReadFile(c.fs, path)
	if err != nil {
		return err
	}

	ext := filepath.Ext(path)
	if len(ext) == 0 {
		return fmt.Errorf("the file %s not ext", path)
	}

	format := ext[1:]
	if _, ok := c.rs.Funcs[format]; !ok {
		if fn := getUnmarshaler(format); fn != nil {
			c.rs.Funcs[format] = fn
		}
	}

	name := filepath.Base(path)
	tag, err := language.Parse(name[:len(name)-len(ext)])
	if err != nil {
		return err
	}

	c.rs.Entries = append(c.rs.Entries, Entry{
		Tag:   tag,
		Name:  path,
		Bytes: buf,
	})
	return nil
}

func (c *fsLoader) load(path string) error {
	entries, err := fs.ReadDir(c.fs, path)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		name := entry.Name()
		fp := filepath.Join(path, name)
		if entry.IsDir() {
			err = c.load(fp)
			if err != nil {
				return err
			}
		} else {
			err = c.loadFile(fp)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
