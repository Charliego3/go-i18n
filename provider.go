package i18n

import (
	"net/http"

	"golang.org/x/text/language"
)

func parseLanguage(val string) language.Tag {
	if len(val) == 0 {
		return language.Und
	}

	tag, err := language.Parse(val)
	if err != nil {
		return language.Und
	}
	return tag
}

func ParseFromHeader(val string) language.Tag {
	tags, _, err := language.ParseAcceptLanguage(val)
	if err != nil {
		return language.Und
	}

	idx := -1
	for i, tag := range tags {
		if _, ok := g.localizes[tag]; ok {
			idx = i
			break
		}
	}

	if idx < 0 {
		return language.Und
	}
	return tags[idx]
}

func HeaderProvider(r *http.Request, key string) language.Tag {
	return ParseFromHeader(r.Header.Get(key))
}

func CookieProvider(r *http.Request, key string) language.Tag {
	val, err := r.Cookie(key)
	if err != nil {
		return language.Und
	}
	return parseLanguage(val.Value)
}

func QueryProvider(r *http.Request, key string) language.Tag {
	return parseLanguage(r.URL.Query().Get(key))
}

func FormProvider(r *http.Request, key string) language.Tag {
	return parseLanguage(r.FormValue(key))
}

func PostFormProvider(r *http.Request, key string) language.Tag {
	return parseLanguage(r.PostFormValue(key))
}
