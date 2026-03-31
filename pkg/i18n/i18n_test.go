package i18n

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTranslator_T_ZH(t *testing.T) {
	translator := NewTranslator("zh")

	assert.Equal(t, "初始化成功", translator.T("init_success"))
	assert.Equal(t, "正在采集桥梁数据...", translator.T("fetching"))
	assert.Equal(t, "正在验证桥梁可用性...", translator.T("validating"))
}

func TestTranslator_T_EN(t *testing.T) {
	translator := NewTranslator("en")

	assert.Equal(t, "Initialization successful", translator.T("init_success"))
	assert.Equal(t, "Fetching bridge data...", translator.T("fetching"))
	assert.Equal(t, "Validating bridge availability...", translator.T("validating"))
}

func TestTranslator_SetLang(t *testing.T) {
	translator := NewTranslator("zh")
	assert.Equal(t, "初始化成功", translator.T("init_success"))

	translator.SetLang("en")
	assert.Equal(t, "Initialization successful", translator.T("init_success"))
}

func TestTranslator_MissingKey(t *testing.T) {
	translator := NewTranslator("zh")

	key := "non_existent_key"
	result := translator.T(key)

	assert.Equal(t, key, result)
}

func TestTranslator_EmptyLang(t *testing.T) {
	translator := NewTranslator("")
	assert.Equal(t, "初始化成功", translator.T("init_success"))
}

func TestGetTranslator(t *testing.T) {
	translator := GetTranslator("zh")
	assert.NotNil(t, translator)

	translator2 := GetTranslator("en")
	assert.NotNil(t, translator2)
}

func TestNewTranslator(t *testing.T) {
	tests := []struct {
		lang     string
		wantLang string
	}{
		{"zh", "zh"},
		{"en", "en"},
	}

	for _, tt := range tests {
		t.Run(tt.lang, func(t *testing.T) {
			translator := NewTranslator(tt.lang)
			assert.Equal(t, tt.wantLang, translator.lang)
		})
	}
}
