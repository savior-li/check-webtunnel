package i18n

import (
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

var bundle *i18n.Bundle
var currentLang string = "en"

func Init() error {
	bundle = i18n.NewBundle(language.English)

	bundle.AddMessages(language.English, &i18n.Message{
		ID:    "app_name",
		Other: "Tor Bridge Collector",
	})
	bundle.AddMessages(language.English, &i18n.Message{
		ID:    "init_success",
		Other: "Initialization completed successfully",
	})
	bundle.AddMessages(language.English, &i18n.Message{
		ID:    "fetch_started",
		Other: "Fetching bridges...",
	})
	bundle.AddMessages(language.English, &i18n.Message{
		ID:    "fetch_success",
		Other: "Successfully fetched {{.Count}} bridges",
	})
	bundle.AddMessages(language.English, &i18n.Message{
		ID:    "fetch_failed",
		Other: "Failed to fetch bridges: {{.Error}}",
	})
	bundle.AddMessages(language.English, &i18n.Message{
		ID:    "validate_started",
		Other: "Validating bridges...",
	})
	bundle.AddMessages(language.English, &i18n.Message{
		ID:    "validate_success",
		Other: "Validation completed: {{.Valid}} valid, {{.Invalid}} invalid",
	})
	bundle.AddMessages(language.English, &i18n.Message{
		ID:    "stats_title",
		Other: "Bridge Statistics",
	})
	bundle.AddMessages(language.English, &i18n.Message{
		ID:    "serve_started",
		Other: "Server started on {{.Host}}:{{.Port}}",
	})

	bundle.AddMessages(language.Chinese, &i18n.Message{
		ID:    "app_name",
		Other: "Tor Bridge 采集器",
	})
	bundle.AddMessages(language.Chinese, &i18n.Message{
		ID:    "init_success",
		Other: "初始化完成",
	})
	bundle.AddMessages(language.Chinese, &i18n.Message{
		ID:    "fetch_started",
		Other: "正在采集 bridges...",
	})
	bundle.AddMessages(language.Chinese, &i18n.Message{
		ID:    "fetch_success",
		Other: "成功采集 {{.Count}} 个 bridges",
	})
	bundle.AddMessages(language.Chinese, &i18n.Message{
		ID:    "fetch_failed",
		Other: "采集 bridges 失败: {{.Error}}",
	})
	bundle.AddMessages(language.Chinese, &i18n.Message{
		ID:    "validate_started",
		Other: "正在验证 bridges...",
	})
	bundle.AddMessages(language.Chinese, &i18n.Message{
		ID:    "validate_success",
		Other: "验证完成: {{.Valid}} 个有效, {{.Invalid}} 个无效",
	})
	bundle.AddMessages(language.Chinese, &i18n.Message{
		ID:    "stats_title",
		Other: "Bridge 统计信息",
	})
	bundle.AddMessages(language.Chinese, &i18n.Message{
		ID:    "serve_started",
		Other: "服务器已在 {{.Host}}:{{.Port}} 启动",
	})

	return nil
}

func GetMessage(key string, lang string, args map[string]interface{}) string {
	if bundle == nil {
		return key
	}

	localizer := i18n.NewLocalizer(bundle, lang)
	msg, err := localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    key,
			Other: key,
		},
		TemplateData: args,
	})
	if err != nil {
		return key
	}
	return msg
}

func SetLang(lang string) {
	currentLang = lang
}

func GetLang() string {
	return currentLang
}
