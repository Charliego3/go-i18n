package i18n

import "golang.org/x/text/language"

type Option func(*options)

// WithLoader Register the Loader interface to *I18n.bundle
//
// Example:
//
//	//go:embed examples/lan2/*
//	var langFS embed.FS
//	i18n.Localize(http.Handler, i18n.NewLoaderWithPath("language_file_path"))
//	i18n.Localize(http.Handler, i18n.NewLoaderWithFS(langFS, i18n.WithUnmarshal("json", json.Unmarshal)))
func WithLoader(loader Loader) Option {
	return func(o *options) {
		o.loaders = append(o.loaders, loader)
	}
}

// WithLanguageProvider get the language from *http.Request,
// default LanguageProvider the order of acquisition is: header(always get the value of Accept-Language) -> cookie -> query -> form -> postForm
// you can use WithLanguageKey change the default lang key
//
// Example:
//
//	loader := i18n.NewLoaderWithPath("language_file_path")
//	i18n.Localize(http.Handler, i18n.WithLoader(loader),
//	    i18n.WithLanguageProvider(i18n.LangHandlerFunc(func(r *http.Request) language.Tag {
//		    lang := r.Header.Get("Accept-Language")
//		    tag, err := language.Parse(lang)
//		    if err != nil {
//			    return language.Chinese
//		    }
//		    return tag
//	    },
//	)))
func WithLanguageProvider(providers ...LanguageProvider) Option {
	return func(o *options) {
		o.providers = providers
	}
}

// WithLanguageKey specifies the default language key when obtained from the LanguageProvider
// Except from the Header, there is no limit if you specify LanguageProvider manually
//
// Example:
//
//	i18n.loader :=i18n.NewLoaderWithPath("language_file_path")
//	i18n.Localize(http.Handler, i18n.WithLoader(loader), i18n.WithLanguageKey("default_language_key"))
func WithLanguageKey(key string) Option {
	return func(o *options) {
		o.languageKey = key
	}
}

// WithDefaultLanguage specify the default language,
// which is used when it is not available from the LanguageProvider
//
// Example:
//
//	i18n.loader :=i18n.NewLoaderWithPath("language_file_path")
//	i18n.Localize(http.Handler, i18n.WithLoader(loader), i18n.WithDefaultLanguage(language.Chinese))
func WithDefaultLanguage(tag language.Tag) Option {
	return func(o *options) {
		o.defLanguage = tag
	}
}

// Loader Options

type LOpt func(*FSLoader)

// WithUnmarshalls register multi format unmarshal func
func WithUnmarshalls(fns map[string]UnmarshalFunc) LOpt {
	return func(l *FSLoader) {
		l.ums = fns
	}
}

// WithUnmarshal register single format unmarshal func
func WithUnmarshal(format string, fn UnmarshalFunc) LOpt {
	return func(l *FSLoader) {
		if l.ums == nil {
			l.ums = make(map[string]UnmarshalFunc)
		}
		l.ums[format] = fn
	}
}
