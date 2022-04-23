package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/zh"
	"github.com/go-playground/universal-translator"
	"gopkg.in/go-playground/validator.v9"
	en_translations "gopkg.in/go-playground/validator.v9/translations/en"
	zh_translations "gopkg.in/go-playground/validator.v9/translations/zh"
	"reflect"
	"regexp"
	"strings"
	"wechatGin/public"
)

//设置Translation
func TranslationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		//参照：https://github.com/go-playground/validator/blob/v9/_examples/translations/main.go

		//设置支持语言
		en := en.New()
		zh := zh.New()

		//设置国际化翻译器
		uni := ut.New(zh, zh, en)
		val := validator.New()

		//根据参数取翻译器实例
		locale := c.DefaultQuery("locale", "zh")
		trans, _ := uni.GetTranslator(locale)

		//翻译器注册到validator
		switch locale {
		case "en":
			en_translations.RegisterDefaultTranslations(val, trans)
			val.RegisterTagNameFunc(func(fld reflect.StructField) string {
				return fld.Tag.Get("en_comment")
			})
			break
		default:
			zh_translations.RegisterDefaultTranslations(val, trans)
			val.RegisterTagNameFunc(func(fld reflect.StructField) string {
				return fld.Tag.Get("comment")
			})

			// 自定义验证方法
			// 验证邮箱
			val.RegisterValidation("validaEmail", func(fl validator.FieldLevel) bool {
				matched, _ := regexp.Match(`[^@ \t\r\n]+@[^@ \t\r\n]+\.[^@ \t\r\n]+`, []byte(strings.ToUpper(fl.Field().String())))
				return matched
			})

			// 自定义验证器
			// 验证邮箱是否符合要求
			val.RegisterTranslation("validaEmail", trans, func(ut ut.Translator) error {
				return ut.Add("validaEmail", "{0} 格式不正确", true)
			}, func(ut ut.Translator, fe validator.FieldError) string {
				t, _ := ut.T("validaEmail", fe.Field())
				return t
			})
			break
		}
		c.Set(public.TranslatorKey, trans)
		c.Set(public.ValidatorKey, val)
		c.Next()
	}
}
