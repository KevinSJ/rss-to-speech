package helper

import (
	"reflect"
	"testing"
)

func Test_chunksByte(t *testing.T) {
	type args struct {
		s         string
		chunkSize int
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "should return nil if empty string",
			args: args{},
			want: nil,
		},
		{
			name: "should return passed string without chunk",
			args: args{
				s:         "a",
				chunkSize: 2,
			},
			want: []string{"a"},
		},
		{
			name: "should return passed string without chunk",
			args: args{
				s:         "test",
				chunkSize: 5,
			},
			want: []string{"test"},
		},
		{
			name: "should return passed string without chunk",
			args: args{
				s: "世",
				// each chinese char uses 3 byte
				chunkSize: 2,
			},
			want: []string{"世"},
		},
		{
			name: "should create chunked string (zh-CN)",
			args: args{
				s: "世,界",
				// each chinese char uses 3 byte
				chunkSize: 3,
			},
			want: []string{"世", ",", "界"},
		},
		{
			name: "should create chunked string (zh-CN)",
			args: args{
				s: "世界,",
				// each chinese char uses 3 byte
				chunkSize: 4,
			},
			want: []string{"世", "界,"},
		},
		{
			name: "should create chunked string (en-US)",
			args: args{
				s:         "Hello world",
				chunkSize: 5,
			},
			want: []string{"Hello", " worl", "d"},
		},
		{
			name: "should not chunk string (zh-CN)",
			args: args{
				s: "世界",
				// each chinese char uses 3 byte
				chunkSize: 6,
			},
			want: []string{"世界"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := chunksByte(tt.args.s, tt.args.chunkSize); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("chunksByte() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetSanitizedLangCode(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "return sanitized language code",
			args: args{s: "en-us"},
			want: "en-US",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetSanitizedLangCode(tt.args.s); got != tt.want {
				t.Errorf("GetSanitizedLangCode() = %v, want %v", got, tt.want)
			}
		})
	}
}
