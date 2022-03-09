// Package i18n provides simplicity and ease of use, no specific framework restrictions, easy access to any framework based on http.Handler
//
// Installation:
//
// 	go get github.com/Charliego93/go-i18n
//
// Example:
//
// 	import (
// 	 	"embed"
// 		"encoding/json"
// 		"fmt"
// 		"github.com/BurntSushi/toml"
// 		"github.com/Charliego93/go-i18n"
// 	 	"github.com/gin-gonic/gin"
// 		"golang.org/x/text/language"
// 		"gopkg.in/yaml.v2"
// 		"net/http"
// 	)
//
// 	//go:embed examples/lan2/*
// 	var langFS embed.FS
//
// 	func main() {
// 		engine := gin.New()
//
// 		// returns the default language if the header and language key are not specified or if the language does not exist
// 		engine.Use(gin.WrapH(i18n.Localize(language.Chinese,
// 		i18n.WithLoader(i18n.NewLoaderWithPath("./examples/simple")))))
//
// 		// Use multi loader provider
// 		// Built-in load from file and load from fs.FS
// 		// engine.Use(gin.WrapH(i18n.Localize(language.Chinese,
// 		// 	i18n.WithLoader(i18n.NewLoaderWithFS(langFS),
// 		// 		i18n.NewLoaderWithPath("./examples/lan1")))))
//
// 		// curl -H "Accept-Language: en" 'http://127.0.0.1:9090/Hello'  returns "hello"
// 		// curl -H "Accept-Language: uk" 'http://127.0.0.1:9090/Hello'  returns "Бонгу"
// 		// curl 'http://127.0.0.1:9090/Hello?lang=en'  returns "hello"
// 		// curl 'http://127.0.0.1:9090/Hello?lang=uk'  returns "Бонгу"
// 		engine.GET("/:messageId", func(ctx *gin.Context) {
// 			ctx.String(http.StatusOK, i18n.MustTr(ctx.Param("messageId")))
// 		})
//
// 		// curl -H "Accept-Language: en" 'http://127.0.0.1:9090/HelloName/I18n'  returns "hello I18n"
// 		// curl -H "Accept-Language: uk" 'http://127.0.0.1:9090/HelloName/I18n'  returns "Бонгу I18n"
// 		// curl 'http://127.0.0.1:9090/HelloName/I18n?lang=en'  returns "hello I18n"
// 		// curl 'http://127.0.0.1:9090/HelloName/I18n?lang=uk'  returns "Бонгу I18n"
// 		engine.GET("/:messageId/:name", func(ctx *gin.Context) {
// 			ctx.String(http.StatusOK, i18n.MustTr(&i18n.LocalizeConfig{
// 			   MessageID: ctx.Param("messageId"),
// 			   TemplateData: map[string]string{
// 				  "Name": ctx.Param("name"),
// 			   },
// 			}))
// 		})
//
// 		fmt.Println(engine.Run())
// 	}
//
// Customize Loader
//
// You can implement your own Loader by yourself, and even pull the language files from any
// possible place to use, just pay attention when implementing the ParseMessage(i *I18n) error function:
// 1. At least need to call i.SetLocalizer(language.Tag) and i.MastParseMessageFileBytes([]byte, string) to register with Bundle
//    - []byte is the file content
//    - string is the file path: mainly used to parse the language and serialization type, for example: en.yaml
// 2. Sometimes it is necessary to call i.RegisterUnmarshalFunc(string, UnmarshalFunc) to register the deserialization function:
//   - string is the format type. eg: yaml
//   - UnmarshalFunc eg: json.Unmarshal or yaml.Unmarshal
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

type OptionFunc func(*I18n)
type loaders struct{ ls []Loader }
type langHandler struct{ LangHandler }
type LangHandlerFunc func(*http.Request) language.Tag
type langKey string

func (l langKey) apply(i *I18n)                                 { i.langKey = string(l) }
func (f OptionFunc) apply(i *I18n)                              { f(i) }
func (h langHandler) apply(i *I18n)                             { i.langHandler = h }
func (f LangHandlerFunc) Language(r *http.Request) language.Tag { return f(r) }
func (l loaders) apply(i *I18n) {
	for _, l := range l.ls {
		i.AddLoader(l)
	}
}

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
//	i18n.Localize(language.Chinese, i18n.WithLoader(i18n.NewLoaderWithPath("language_file_path")))
//	i18n.Localize(language.Chinese, i18n.WithLoader(i18n.NewLoaderWithFS(langFS, i18n.WithUnmarshal("json", json.Unmarshal))))
func WithLoader(ls ...Loader) Option { return loaders{ls} }

// WithLangHandler get the language from *http.Request,
// default LangHandler the order of acquisition is: header(always get the value of Accept-Language) -> cookie -> query -> form -> postForm
// you can use WithLangKey change the default lang key
//
// Example:
//
//	loader := i18n.WithLoader(i18n.NewLoaderWithPath("language_file_path"))
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
//	i18n.loader := i18n.WithLoader(i18n.NewLoaderWithPath("language_file_path"))
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
