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

func TestURLOnly(t *testing.T) {
	type args struct {
		url string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"test", args{"https://www.youtube.com/live/bWgxzE8J0i8?si=uEdOSrcqdVINSDkH"}, true},
		{"test", args{"https://www.youtube.com/live/bWgxzE8J0i8?si=uEdOSrcqdVINSDkH https://www.youtube.com/live/bWgxzE8J0i8?si=uEdOSrcqdVINSDkH"}, true},
		{"test", args{"https://www.youtube.com/live/bWgxzE8J0i8?si=uEdOSrcqdVINSDkHhttps://www.youtube.com/live/bWgxzE8J0i8?si=uEdOSrcqdVINSDkH"}, true},
		{"test", args{"https://www.youtube.com/live/bWgxzE8J0i8?si=uEdOSrcqdVINSDkH a"}, false},
		{"test", args{"https://www.youtube.com/live/bWgxzE8J0i8?si=uEdOSrcqdVINSDkH a https://www.youtube.com/live/bWgxzE8J0i8?si=uEdOSrcqdVINSDkH a"}, false},
		{"test", args{`
		https://www.youtube.com/live/bWgxzE8J0i8?si=uEdOSrcqdVINSDkH
		https://www.youtube.com/live/bWgxzE8J0i8?si=uEdOSrcqdVINSDkH
		https://www.youtube.com/live/bWgxzE8J0i8?si=uEdOSrcqdVINSDkH https://www.youtube.com/live/bWgxzE8J0i8?si=uEdOSrcqdVINSDkH`}, true},
		{"test", args{"||https://www.youtube.com/live/bWgxzE8J0i8?si=uEdOSrcqdVINSDkH  https://www.youtube.com/live/bWgxzE8J0i8?si=uEdOSrcqdVINSDkH ||"}, true },
		{"test", args{"||https://www.youtube.com/live/bWgxzE8J0i8?si=uEdOSrcqdVINSDkH |||| https://www.youtube.com/live/bWgxzE8J0i8?si=uEdOSrcqdVINSDkH ||"}, true },
		{"test", args{"||https://www.youtube.com/live/bWgxzE8J0i8?si=uEdOSrcqdVINSDkH  https://www.youtube.com/live/bWgxzE8J0i8?si=uEdOSrcqdVINSDkH a ||"}, false },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			notUrlOnly, err := impureUrlsDetector.MatchString(tt.args.url)
			if err != nil {
				t.Errorf("Failed to detect if message is URL only: %v", err)
				t.Fail()
			}
			if !notUrlOnly != tt.want {
				t.Errorf("%s: %v, want %v", tt.args.url, !notUrlOnly, tt.want)
			}
		})
	}
}

func TestUrlExtractor(t *testing.T) {
	type args struct {
		input string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "Single URL",
			args: args{input: "Check out this link: https://example.com/page"},
			want: []string{"https://example.com/page"},
		},
		{
			name: "Multiple URLs",
			args: args{input: "Here are two links: https://example.com and https://test.com/page"},
			want: []string{"https://example.com", "https://test.com/page"},
		},
		{
			name: "URL with Discord spoiler",
			args: args{input: "Secret link ||https://secret.com|| is hidden"},
			want: []string{"||https://secret.com||"},
		},
		{
			name: "Mixed URLs with and without spoilers",
			args: args{input: "Public link https://public.com and secret ||https://secret.com||"},
			want: []string{"https://public.com", "||https://secret.com||"},
		},
		{
			name: "No URLs",
			args: args{input: "This string contains no links."},
			want: []string{},
		},
		{
			name: "Invalid URLs",
			args: args{input: "Invalid link https:/bad.com and valid https://good.com"},
			want: []string{"https://good.com"},
		},
		{
			name: "Multiple URLs with various formats",
			args: args{input: "Links: ||https://example.com||, https://test.com/page, and ||https://another.com||"},
			want: []string{"||https://example.com||", "https://test.com/page", "||https://another.com||"},
		},
		{
			name: "URLs with trailing characters",
			args: args{input: "Check these: https://example.com/page?param=value, ||https://secret.com||!"},
			want: []string{"https://example.com/page?param=value", "||https://secret.com||"},
		},
		{
			name: "URLs separated by newlines",
			args: args{input: "First link: https://first.com\nSecond link: ||https://second.com||"},
			want: []string{"https://first.com", "||https://second.com||"},
		},
		{
			name: "URLs with pipe characters",
			args: args{input: "Multiple ||https://example.com|| ||https://test.com|| links."},
			want: []string{"||https://example.com||", "||https://test.com||"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got []string
			m, err := urlExtractor.FindStringMatch(tt.args.input)
			if err != nil {
				t.Errorf("urlExtractor.FindStringMatch() error = %v", err)
				return
			}
			for m != nil {
				got = append(got, m.String())
				m, err = urlExtractor.FindNextMatch(m)
				if err != nil {
					t.Errorf("urlExtractor.FindNextMatch() error = %v", err)
					break
				}
			}

			if len(got) != len(tt.want) {
				t.Errorf("Number of matches = %d, want %d", len(got), len(tt.want))
			}
			for i := range tt.want {
				if i >= len(got) || got[i] != tt.want[i] {
					t.Errorf("Match %d = \"%v\", want \"%v\"", i, got[i], tt.want[i])
				}
			}
		})
	}
}
