package main

import (
	"testing"
)

func TestFetchAndLoadJSON(t *testing.T) {
	type args struct {
		url string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"test", args{repo}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, err := FetchAndLoadJSON(tt.args.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("FetchAndLoadJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			t.Logf("Providers: %+v", d)
		})
	}
}
