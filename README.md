# go-i18n

[![GitHub Workflow Status (branch)](https://img.shields.io/github/workflow/status/Charliego93/go-i18n/Go/main?logo=github)](https://github.com/Charliego93/go-i18n/actions/workflows/go.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/Charliego93/go-i18n)](https://goreportcard.com/report/github.com/Charliego93/go-i18n)
[![Coveralls](https://img.shields.io/coveralls/github/Charliego93/go-i18n?logo=coveralls&color=8345D5)](https://coveralls.io/github/Charliego93/go-i18n?branch=main)
[![GitHub tag (latest SemVer pre-release)](https://img.shields.io/github/v/tag/Charliego93/go-i18n?include_prereleases)](https://github.com/Charliego93/go-i18n/tags)
[![GitHub](https://img.shields.io/github/license/Charliego93/go-i18n?color=D96BA2)](https://github.com/Charliego93/go-i18n/blob/main/LICENSE)
[![GoDoc](https://godoc.org/github.com/Charliego93/go-i18n?status.svg)](https://godoc.org/github.com/Charliego93/go-i18n)

Provides simplicity and ease of use, no specific framework restrictions, easy access to any framework based on `http.Handler`

# Table of Contents

- [Installation](#installation)
- [Usage](#usage)
- [Customize Loader](#customize-loader)
- [License](#license)

## Installation

```shell
go get github.com/Charliego93/go-i18n
```

## Usage

```go
package main

import (
   "embed"
   "fmt"
   "github.com/Charliego93/go-i18n"
   "github.com/gin-gonic/gin"
   "net/http"
)

//go:embed examples/lan2/*
var langFS embed.FS

func main() {
   engine := gin.New()

   // returns the default language if the header and language key are not specified or if the language does not exist
   engine.Use(gin.WrapH(i18n.Localize(language.Chinese,
      i18n.WithLoader(i18n.NewLoaderWithPath("./examples/simple")))))

   // Use multi loader provider
   // Built-in load from file and load from fs.FS
   // engine.Use(gin.WrapH(i18n.Localize(language.Chinese,
   // 	i18n.WithLoader(i18n.NewLoaderWithFS(langFS),
   // 		i18n.NewLoaderWithPath("./examples/lan1")))))

   // curl -H "Accept-Language: en" 'http://127.0.0.1:9090/Hello'  returns "hello"
   // curl -H "Accept-Language: uk" 'http://127.0.0.1:9090/Hello'  returns "Бонгу"
   // curl 'http://127.0.0.1:9090/Hello?lang=en'  returns "hello"
   // curl 'http://127.0.0.1:9090/Hello?lang=uk'  returns "Бонгу"
   engine.GET("/:messageId", func(ctx *gin.Context) {
      ctx.String(http.StatusOK, i18n.MustTr(ctx.Param("messageId")))
   })

   // curl -H "Accept-Language: en" 'http://127.0.0.1:9090/HelloName/I18n'  returns "hello I18n"
   // curl -H "Accept-Language: uk" 'http://127.0.0.1:9090/HelloName/I18n'  returns "Бонгу I18n"
   // curl 'http://127.0.0.1:9090/HelloName/I18n?lang=en'  returns "hello I18n"
   // curl 'http://127.0.0.1:9090/HelloName/I18n?lang=uk'  returns "Бонгу I18n"
   engine.GET("/:messageId/:name", func(ctx *gin.Context) {
      ctx.String(http.StatusOK, i18n.MustTr(&i18n.LocalizeConfig{
         MessageID: ctx.Param("messageId"),
         TemplateData: map[string]string{
            "Name": ctx.Param("name"),
         },
      }))
   })

   fmt.Println(engine.Run())
}
```

## Customize Loader

You can implement your own `Loader` by yourself, and even pull the language files from any 
possible place to use, just pay attention when implementing the `ParseMessage(i *I18n) error` function:
1. At least need to call `i.SetLocalizer(language.Tag)` and `i.MastParseMessageFileBytes([]byte, string)` to register with `Bundle`
    - `[]byte` is the file content
    - `string` is the file path: mainly used to parse the language and serialization type, for example: `en.yaml`
2. Sometimes it is necessary to call `i.RegisterUnmarshalFunc(string, UnmarshalFunc)` to register the deserialization function:
   - `string` is the format type. eg: `yaml`
   - `UnmarshalFunc` eg: `json.Unmarshal` or `yaml.Unmarshal`

## License

[MIT © Charliego93.](LICENSE)
