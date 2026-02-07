package domain

import (
	"testing"
)

func TestKeyValidate(t *testing.T) {
	tests := []struct {
		name    string
		keyName string
		wantErr bool
	}{
		{"Valid ASCII", "my-key_123", false},
		{"Valid Japanese", "テストキー", false},
		{"Valid Mixed", "Key-日本語_123", false},
		{"Valid with Spaces", "My Key Name", false},
		{"Too long (ASCII)", "a234567890b234567890c234567890d234567890e2345678901", true},
		{"Too long (Japanese)", "一二三四五六七八九十一二三四五六七八九十一二三四五六七八九十一二三四五六七八九十一二三四五六七八九十一", true},
		{"Valid 50 characters (Japanese)", "あいうえおかきくけこさしすせそたちつてとなにぬねのはひふへほまみむめもやいゆえよらりるれろわいうえを", false},
		{"Empty name", "", true},
		{"Invalid characters", "key!", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := &Key{
				Name: tt.keyName,
				Configuration: &KeyConfiguration{
					Provider: PluginConfig{ID: "openai"},
				},
			}
			err := k.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Key.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
