package config

import (
	"path/filepath"
	"reflect"
	"testing"
)

func getPath(s string) string {
	if path, _ := filepath.Abs(s); path != "" {
		return path
	}
	return ""
}

func TestNewConfig(t *testing.T) {
	type args struct {
		fileName string
	}
	tests := []struct {
		name    string
		args    args
		want    *Config
		wantErr bool
	}{
		{
			name:    "Fail on invalid file name",
			args:    args{},
			wantErr: true,
		},
		{
			name: "Fail on empty config file",
			args: args{
				fileName: "./test_configs/empty.yaml",
			},
			wantErr: true,
		},
		{
			name: "Fail on missing credential path",
			args: args{
				fileName: "./test_configs/missing_credential_path.yaml",
			},
			wantErr: true,
		},
		{
			name: "Use DEFAULT_CONFIG if missing certain values",
			args: args{
				fileName: "./test_configs/using_default.yaml",
			},
			want: &Config{
				ConcurrentWorkers: DEFAULT_CONFIG.ConcurrentWorkers,
				MaxItemPerFeed:    DEFAULT_CONFIG.MaxItemPerFeed,
				ItemSince:         DEFAULT_CONFIG.ItemSince,
				Feeds: []string{
					"https://jiangsc.me/feed.xml",
				},
				CredentialPath: "./some_file.json",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewConfig(tt.args.fileName)
			t.Logf("err: %v\n", err)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_getFullCredentialPath(t *testing.T) {
	type fields struct {
		Feeds             []string
		ItemSince         float64
		ConcurrentWorkers int
		CredentialPath    string
		MaxItemPerFeed    int
	}
	path := "./some_file.json"
	tests := []struct {
		name         string
		fields       fields
		wantFullPath string
		wantErr      bool
	}{
		{
			name: "should return full path",
			fields: fields{
				Feeds:             []string{},
				ItemSince:         0,
				ConcurrentWorkers: 0,
				CredentialPath:    path,
				MaxItemPerFeed:    0,
			},
			wantFullPath: getPath(path),
			wantErr:      false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &Config{
				Feeds:             tt.fields.Feeds,
				ItemSince:         tt.fields.ItemSince,
				ConcurrentWorkers: tt.fields.ConcurrentWorkers,
				CredentialPath:    tt.fields.CredentialPath,
				MaxItemPerFeed:    tt.fields.MaxItemPerFeed,
			}
			gotFullPath, err := tr.getFullCredentialPath()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.getFullCredentialPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotFullPath != tt.wantFullPath {
				t.Errorf("Config.getFullCredentialPath() = %v, want %v", gotFullPath, tt.wantFullPath)
			}
		})
	}
}
