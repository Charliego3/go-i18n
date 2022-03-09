package i18n

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v2"
	"net/http"
)

type LocalizeConfig = i18n.LocalizeConfig
type UnmarshalFunc = i18n.UnmarshalFunc
type Option interface{ apply(*I18n) }
type ContextHandler struct{}
type LangHandler interface {
	Language(*http.Request) language.Tag
}

type loader struct{ l Loader }
type langKey string
type OptionFunc func(*I18n)
type langHandler struct{ LangHandler }
type LangHandlerFunc func(*http.Request) language.Tag

func (l loader) apply(i *I18n)                                  { i.AddLoader(l.l) }
func (l langKey) apply(i *I18n)                                 { i.langKey = string(l) }
func (f OptionFunc) apply(i *I18n)                              { f(i) }
func (h langHandler) apply(i *I18n)                             { i.langHandler = h }
func (f LangHandlerFunc) Language(r *http.Request) language.Tag { return f(r) }

const languageKey = "Accept-Language"

func (_ ContextHandler) ServeHTTP(_ http.ResponseWriter, r *http.Request) {
	lang := i.langHandler.Language(r)
	if i.ctx != nil {
		if tag, ok := i.ctx.Value(languageKey).(language.Tag); ok && tag == lang {
			return
		}
	}

	ctx := context.WithValue(r.Context(), languageKey, lang)
	i.ctx = ctx
}

// WithLoader Register the Loader interface to *I18n.bundle
//
// Example:
//
//	//go:embed examples/lan2/*
//	var langFS embed.FS
//	i18n.Localize(language.Chinese, i18n.NewLoaderWithPath("language_file_path"))
//	i18n.Localize(language.Chinese, i18n.NewLoaderWithFS(langFS, i18n.WithUnmarshal("json", json.Unmarshal)))
func WithLoader(l Loader) Option { return loader{l} }

// WithLangHandler get the language from *http.Request,
// default LangHandler the order of acquisition is: header(always get the value of Accept-Language) -> cookie -> query -> form -> postForm
// you can use WithLangKey change the default lang key
//
// Example:
//
//	loader := i18n.NewLoaderWithPath("language_file_path")
//	i18n.Localize(language.Chinese, loader, i18n.WithLangHandler(i18n.LangHandlerFunc(func(r *http.Request) language.Tag {
//		lang := r.Header.Get("Accept-Language")
//		tag, err := language.Parse(lang)
//		if err != nil {
//			return language.Chinese
//		}
//		return tag
//	})))
func WithLangHandler(handler LangHandler) Option { return langHandler{handler} }

// WithLangKey specifies the default language key when obtained from the LangHandler
// Except from the Header, there is no limit if you specify LangHandler manually
//
// Example:
//
//	i18n.loader :=i18n.NewLoaderWithPath("language_file_path")
//	i18n.Localize(language.Chinese, loader, i18n.WithLangKey("default_language_key"))
func WithLangKey(key string) Option { return langKey(key) }

type I18n struct {
	bundle      *i18n.Bundle
	language    language.Tag
	localizes   map[language.Tag]*i18n.Localizer
	langHandler LangHandler
	langKey     string
	ctx         context.Context
}

func (i *I18n) SetDefaultLang(lang language.Tag) {
	i.language = lang
	i.bundle = i18n.NewBundle(lang)
}

func (i *I18n) AddLoader(loader Loader) {
	err := loader.ParseMessage(i)
	if err != nil {
		panic(err)
	}
	i.SetLocalizer(i.language)
}

func (i *I18n) Tr(p interface{}) (string, error) {
	var lang language.Tag
	if i.ctx != nil {
		if value, ok := i.ctx.Value(languageKey).(language.Tag); ok {
			lang = value
		}
	}

	if lang.IsRoot() {
		lang = i.language
	}
	lr := i.getLocalizer(lang)

	var config *i18n.LocalizeConfig

	switch t := p.(type) {
	case string:
		config = &i18n.LocalizeConfig{MessageID: t}
	case i18n.LocalizeConfig:
		config = &t
	case *i18n.LocalizeConfig:
		config = t
	default:
		return "", fmt.Errorf("unsupported param %T", p)
	}

	translated, err := lr.Localize(config)
	if err != nil || len(translated) == 0 {
		return config.MessageID, err
	}
	return translated, nil
}

func (i *I18n) registerUnmarshalFunc(format string) {
	var fn UnmarshalFunc
	switch format {
	case "json":
		fn = json.Unmarshal
	case "toml":
		fn = toml.Unmarshal
	case "yaml":
		fn = yaml.Unmarshal
	}

	if fn != nil {
		i.RegisterUnmarshalFunc(format, fn)
	}
}

func (i *I18n) RegisterUnmarshalFunc(format string, unmarshalFunc UnmarshalFunc) {
	i.bundle.RegisterUnmarshalFunc(format, unmarshalFunc)
}

func (i *I18n) MastParseMessageFileBytes(buf []byte, path string) {
	i.bundle.MustParseMessageFileBytes(buf, path)
}

func (i *I18n) SetLocalizer(lang language.Tag) {
	if _, ok := i.localizes[lang]; ok {
		return
	}

	if i.localizes == nil {
		i.localizes = make(map[language.Tag]*i18n.Localizer)
	}

	langs := []string{lang.String()}
	if lang != i.language {
		langs = append(langs, i.language.String())
	}

	i.localizes[lang] = i18n.NewLocalizer(i.bundle, langs...)
}

func (i *I18n) getLocalizer(lang language.Tag) *i18n.Localizer {
	if lr, ok := i.localizes[lang]; ok {
		return lr
	}

	return i.localizes[i.language]
}

var i *I18n

// Localize initialize i18n...
func Localize(defaultLang language.Tag, opts ...Option) ContextHandler {
	i = &I18n{}
	i.SetDefaultLang(defaultLang)
	for _, opt := range opts {
		opt.apply(i)
	}

	if i.langHandler == nil {
		i.langHandler = &defaultLangHandler{}
	}

	if len(i.langKey) == 0 {
		i.langKey = "lang"
	}
	return ContextHandler{}
}

// Tr translate messageId to target language
//
// Example:
//
//	Tr("hello")
//	Tr(LocalizeConfig{
//		MessageID: "HelloName",
//		TemplateData: map[string]string{
//			"Name": "I18n",
//		},
//	})
func Tr(messageId interface{}) (string, error) {
	return i.Tr(messageId)
}

// MustTr called Tr but ignore error
func MustTr(messageId interface{}) string {
	message, _ := Tr(messageId)
	return message
}

type defaultLangHandler struct{}

// Language header -> cookie -> query -> form -> postForm
func (g *defaultLangHandler) Language(r *http.Request) language.Tag {
	lan := getLan(func() string {
		return r.Header.Get(languageKey)
	}, func() string {
		c, err := r.Cookie(i.langKey)
		if err != nil {
			return ""
		}
		return c.Value
	}, func() string {
		return r.URL.Query().Get(i.langKey)
	}, func() string {
		return r.FormValue(i.langKey)
	}, func() string {
		return r.PostFormValue(i.langKey)
	})

	if len(lan) > 0 {
		tag, err := language.Parse(lan)
		if err != nil {
			return i.language
		}
		return tag
	}

	return i.language
}

func getLan(fns ...func() string) string {
	for _, fn := range fns {
		lan := fn()
		if len(lan) > 0 {
			return lan
		}
	}
	return ""
}
