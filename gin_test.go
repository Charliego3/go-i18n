package i18n

import (
	"embed"
	"github.com/gin-gonic/gin"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"testing"
)

func init() {
	gin.SetMode(gin.ReleaseMode)
}

type ginServer struct {
	*gin.Engine
}

func newGinServer(fn func(engine *ginServer), middleware ...gin.HandlerFunc) *ginServer {
	engine := gin.New()
	engine.Use(middleware...)

	s := &ginServer{engine}
	fn(s)
	return s
}

func (s *ginServer) request(lan language.Tag, messageId, data string, count int64) string {
	path := "/" + filepath.Join(messageId, data) + "?count=" + strconv.FormatInt(count, 10)
	req, _ := http.NewRequest("GET", path, nil)
	req.Header.Add("Accept-Language", lan.String())

	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)

	return w.Body.String()
}

func TestSimple(t *testing.T) {
	i = nil
	business := func(engine *ginServer) {
		engine.GET("/:messageId", func(ctx *gin.Context) {
			ctx.String(http.StatusOK, MustTr(ctx.Param("messageId")))
		})

		engine.GET("/:messageId/:name", func(ctx *gin.Context) {
			ctx.String(http.StatusOK, MustTr(&i18n.LocalizeConfig{
				MessageID: ctx.Param("messageId"),
				TemplateData: map[string]string{
					"Name": ctx.Param("name"),
				},
			}))
		})
	}

	gs := newGinServer(business, gin.WrapH(Localize(language.Chinese, NewLoaderWithPath("./examples/simple"))))

	type args struct {
		lng       language.Tag
		messageId string
		name      string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// Hello
		{name: "chinese_hello", args: args{messageId: "Hello", lng: language.Chinese}, want: "你好"},
		{name: "english_hello", args: args{messageId: "Hello", lng: language.English}, want: "hello"},
		{name: "ukrainian_hello", args: args{messageId: "Hello", lng: language.Ukrainian}, want: "Бонгу"},
		// HelloName
		{name: "chinese_hello_name", args: args{messageId: "HelloName", name: "尼克", lng: language.Chinese}, want: "你好尼克"},
		{name: "english_hello_name", args: args{messageId: "HelloName", name: "Nick", lng: language.English}, want: "hello Nick"},
		{name: "ukrainian_hello_name", args: args{messageId: "HelloName", name: "Nick", lng: language.Ukrainian}, want: "Бонгу Nick"},
		// PersonCats
		{name: "chinese_hello", args: args{messageId: "PersonCats", name: "尼克", lng: language.Chinese}, want: "尼克有几只猫"},
		{name: "english_hello", args: args{messageId: "PersonCats", name: "Nick", lng: language.English}, want: "Nick has a few cats"},
		{name: "ukrainian_hello", args: args{messageId: "PersonCats", name: "Nick", lng: language.Ukrainian}, want: "Nick Є кілька котів"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := gs.request(tt.args.lng, tt.args.messageId, tt.args.name, 0)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

//go:embed examples/lan2/*
var lan2Embed embed.FS

func TestLocalize(t *testing.T) {
	i = nil
	business := func(engine *ginServer) {
		engine.GET("/default/:messageId", func(ctx *gin.Context) {
			ctx.String(http.StatusOK, MustTr(ctx.Param("messageId")))
		})

		engine.GET("/:messageId", func(ctx *gin.Context) {
			ctx.String(http.StatusOK, MustTr(&i18n.LocalizeConfig{
				MessageID:   ctx.Param("messageId"),
				PluralCount: ctx.Query("count"),
			}))
		})

		engine.GET("/:messageId/:name", func(ctx *gin.Context) {
			count := ctx.Query("count")
			ctx.String(http.StatusOK, MustTr(&i18n.LocalizeConfig{
				MessageID: ctx.Param("messageId"),
				TemplateData: map[string]string{
					"Name":  ctx.Param("name"),
					"Count": count,
				},
				PluralCount: count,
			}))
		})
	}
	gs := newGinServer(business,
		gin.WrapH(Localize(language.Chinese, NewLoaderWithFS(lan2Embed),
			NewLoaderWithPath("examples/lan1"))))

	t.Parallel()

	type args struct {
		lng       language.Tag
		messageId string
		name      string
		count     int64
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// Chinese
		{name: "chinese_hello", args: args{messageId: "Hello", lng: language.Chinese}, want: "你好"},
		{name: "chinese_hello_name", args: args{messageId: "HelloName", name: "尼克", lng: language.Chinese}, want: "你好尼克"},
		{name: "chinese_person_cats", args: args{messageId: "PersonCats", name: "尼克", count: 5, lng: language.Chinese}, want: "尼克有5只猫"},
		{name: "chinese_hello_default", args: args{messageId: "default/Hello", lng: language.Chinese}, want: "你好"},

		// English
		{name: "english_hello_one", args: args{messageId: "Hello", count: 1, lng: language.English}, want: "hello"},
		{name: "english_hello_other", args: args{messageId: "Hello", count: 0, lng: language.English}, want: "hello other"},
		{name: "english_hello_default", args: args{messageId: "default/Hello", lng: language.English}, want: "hello other"},
		// English HelloName
		{name: "english_hello_name_one", args: args{messageId: "HelloName", name: "Nick", count: 1, lng: language.English}, want: "hello Nick"},
		{name: "english_hello_name_other", args: args{messageId: "HelloName", name: "Nick", count: 0, lng: language.English}, want: "hello Nick other"},
		{name: "english_hello_name_default", args: args{messageId: "default/HelloName", lng: language.English}, want: "hello <no value> other"},
		// English PersonCats
		{name: "english_person_cats_one", args: args{messageId: "PersonCats", name: "Nick", count: 1, lng: language.English}, want: "Nick has 1 cat."},
		{name: "english_person_cats_other", args: args{messageId: "PersonCats", name: "Nick", count: 0, lng: language.English}, want: "Nick has 0 cats other."},
		{name: "english_person_cats_default", args: args{messageId: "default/PersonCats", lng: language.English}, want: "<no value> has <no value> cats other."},

		// Latvian Hello
		{name: "latvian_hello_one", args: args{messageId: "Hello", count: 1, lng: language.Latvian}, want: "Sveiki"},
		{name: "latvian_hello_other", args: args{messageId: "Hello", count: 3, lng: language.Latvian}, want: "Sveiki other."},
		{name: "latvian_hello_zero", args: args{messageId: "Hello", count: 0, lng: language.Latvian}, want: "Sveiki 0"},
		{name: "latvian_hello_default", args: args{messageId: "default/Hello", lng: language.Latvian}, want: "Sveiki other."},
		// Latvian HelloName
		{name: "latvian_hello_name_one", args: args{messageId: "HelloName", name: "Nick", count: 1, lng: language.Latvian}, want: "Nick Sveiki"},
		{name: "latvian_hello_name_other", args: args{messageId: "HelloName", name: "Nick", count: 3, lng: language.Latvian}, want: "Nick Sveiki other."},
		{name: "latvian_hello_name_zero", args: args{messageId: "HelloName", name: "Nick", count: 0, lng: language.Latvian}, want: "Nick Sveiki 0"},
		{name: "latvian_hello_name_default", args: args{messageId: "default/HelloName", lng: language.Latvian}, want: "<no value> Sveiki other."},
		// Latvian PersonCats
		{name: "latvian_person_cats_one", args: args{messageId: "PersonCats", name: "Nick", count: 1, lng: language.Latvian}, want: "Nick Ir 1 kat"},
		{name: "latvian_person_cats_other", args: args{messageId: "PersonCats", name: "Nick", count: 3, lng: language.Latvian}, want: "Nick Ir daudzskaitlis 3 kat other."},
		{name: "latvian_person_cats_zero", args: args{messageId: "PersonCats", name: "Nick", count: 0, lng: language.Latvian}, want: "Nick nav kaķa"},
		{name: "latvian_person_cats_default", args: args{messageId: "default/PersonCats", lng: language.Latvian}, want: "<no value> Ir daudzskaitlis <no value> kat other."},

		// Ukrainian
		{name: "ukrainian_hello_one", args: args{messageId: "Hello", count: 1, lng: language.Ukrainian}, want: "Бонгу one"},
		{name: "ukrainian_hello_few", args: args{messageId: "Hello", count: 2, lng: language.Ukrainian}, want: "Бонгу few"},
		{name: "ukrainian_hello_many", args: args{messageId: "Hello", count: 0, lng: language.Ukrainian}, want: "Бонгу many"},
		// Ukrainian HelloName
		{name: "ukrainian_hello_name_one", args: args{messageId: "HelloName", name: "Nick", count: 1, lng: language.Ukrainian}, want: "Nick Бонгу one"},
		{name: "ukrainian_hello_name_few", args: args{messageId: "HelloName", name: "Nick", count: 2, lng: language.Ukrainian}, want: "Nick Бонгу few"},
		{name: "ukrainian_hello_name_many", args: args{messageId: "HelloName", name: "Nick", count: 0, lng: language.Ukrainian}, want: "Nick Бонгу many"},
		// Ukrainian PersonCats
		{name: "ukrainian_person_cats_one", args: args{messageId: "PersonCats", name: "Nick", count: 1, lng: language.Ukrainian}, want: "Nick Є 1 кішки one"},
		{name: "ukrainian_person_cats_few", args: args{messageId: "PersonCats", name: "Nick", count: 2, lng: language.Ukrainian}, want: "Nick Є 2 кішки few"},
		{name: "ukrainian_person_cats_many", args: args{messageId: "PersonCats", name: "Nick", count: 0, lng: language.Ukrainian}, want: "Nick Є 0 кішки many"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := gs.request(tt.args.lng, tt.args.messageId, tt.args.name, tt.args.count)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
