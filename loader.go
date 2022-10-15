package i18n

import (
	"fmt"
	"golang.org/x/text/language"
	"io/fs"
	"os"
	"path/filepath"
)

type Loader interface {
	ParseMessage(*options) error
}

func NewLoaderWithPath(path string, opts ...LOpt) Option {
	return NewLoaderWithFS(os.DirFS(path), opts...)
}

func NewLoaderWithFS(fs fs.FS, opts ...LOpt) Option {
	l := &FSLoader{fs: fs}
	for _, opt := range opts {
		opt(l)
	}
	return WithLoader(l)
}

type FSLoader struct {
	fs  fs.FS
	ums map[string]UnmarshalFunc
}

func (c *FSLoader) ParseMessage(o *options) error {
	for format, ufn := range c.ums {
		o.RegisterUnmarshalFunc(format, ufn)
	}

	return c.parseMessage(o, ".")
}

func (c *FSLoader) parse(o *options, name string, buf []byte) error {
	ext := filepath.Ext(name)
	if len(name) == 0 || len(ext) == 0 {
		return fmt.Errorf("the file %s not ext", name)
	}

	format := ext[1:]
	if _, ok := c.ums[format]; !ok {
		o.registerUnmarshalFunc(format)
	}

	name = filepath.Base(name)
	tag, err := language.Parse(name[:len(name)-len(ext)])
	if err != nil {
		return err
	}
	o.SetLocalizer(tag)
	o.MastParseMessageFileBytes(buf, name)
	return nil
}

func (c *FSLoader) parseMessage(o *options, path string) error {
	entries, err := fs.ReadDir(c.fs, path)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		name := entry.Name()
		fp := filepath.Join(path, name)
		if entry.IsDir() {
			err := c.parseMessage(o, fp)
			if err != nil {
				return err
			}
		} else {
			buf, err := fs.ReadFile(c.fs, fp)
			if err != nil {
				return err
			}
			err = c.parse(o, name, buf)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
