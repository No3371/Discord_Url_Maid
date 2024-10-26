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
		{"test", args{"||https://www.youtube.com/live/bWgxzE8J0i8?si=uEdOSrcqdVINSDkH  https://www.youtube.com/live/bWgxzE8J0i8?si=uEdOSrcqdVINSDkH ||"}, true},
		{"test", args{"||https://www.youtube.com/live/bWgxzE8J0i8?si=uEdOSrcqdVINSDkH |||| https://www.youtube.com/live/bWgxzE8J0i8?si=uEdOSrcqdVINSDkH ||"}, true},
		{"test", args{"||https://www.youtube.com/live/bWgxzE8J0i8?si=uEdOSrcqdVINSDkH  https://www.youtube.com/live/bWgxzE8J0i8?si=uEdOSrcqdVINSDkH a ||"}, false},
		{"test", args{"||https://www.youtube.com/live/bWgxzE8J0i8?si=uEdOSrcqdVINSDkH"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deSpoiled, err := spoilerFinder.Replace(tt.args.url, " $1 ", -1, -1)
			if err != nil {
				t.Errorf("Failed to despoil message: %e", err)
			}
			notUrlOnly, err := impureUrlsDetector.MatchString(deSpoiled)
			if err != nil {
				t.Errorf("Failed to detect if message is URL only: %v", err)
			}
			if !notUrlOnly != tt.want {
				t.Errorf("%s: %v, want %v", tt.args.url, !notUrlOnly, tt.want)
			}
		})
	}
}
func TestSpoilerFinder(t *testing.T) {
	type args struct {
		input string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "||||",
			args: args{input: "||||"},
			want: "",
		},
		{
			name: "||||a",
			args: args{input: "||||"},
			want: "",
		},
		{
			name: "||||a||",
			args: args{input: "||||a||"},
			want: "||||a||",
		},
		{
			name: `||\n||||a`,
			args: args{input: `||
||||a`},
			want: `||
||`,
		},
		{
			name: "||||||a",
			args: args{input: "||||||a"},
			want: "|||||",
		},
		{
			name: "|| ||a| ||",
			args: args{input: "|| ||a| ||"},
			want: "|| ||",
		},
		{
			name: "||||a| ||",
			args: args{input: "||||a| ||"},
			want: "||||a| ||",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got string
			m, err := spoilerFinder.FindStringMatch(tt.args.input)
			if err != nil {
				t.Errorf("spoilerExtractor.FindStringMatch() error = %v", err)
				return
			}
			if m == nil {
				if tt.want == "" {
					return
				}

				t.Errorf("No match found")
				return
			}
			got = m.String()
			if got != tt.want {
				t.Errorf("Match = \"%v\", want \"%v\"", got, tt.want)
			}
		})
	}
}

func TestEnforceSpoilerEdges(t *testing.T) {
	type args struct {
		input string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "",
			args: args{input: "||https://www.youtube.com/watch?v=qQiVUv7RIPs&t=770||"},
			want: "|| https://www.youtube.com/watch?v=qQiVUv7RIPs&t=770 ||",
		},
		{
			name: "|||||",
			args: args{input: "|||||"},
			want: "|| | ||",
		},
		{
			name: "||||",
			args: args{input: "||||"},
			want: "||||", // NO CHANGE
		},
		{
			name: "|| ||",
			args: args{input: "|| ||"},
			want: "||   ||",
		},
		{
			name: "https://www.youtube.com/watch?v=qQiVUv7RIPs&t=770",
			args: args{input: "https://www.youtube.com/watch?v=qQiVUv7RIPs&t=770"},
			want: "https://www.youtube.com/watch?v=qQiVUv7RIPs&t=770",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := enforceSpoilerPadding(tt.args.input)
			if err != nil {
				t.Errorf("spoilerSpaceEdgeInserter.FindStringMatch() error = %v", err)
				return
			}
			if m != tt.want {
				t.Errorf("Match = \"%v\", want \"%v\"", m, tt.want)
			}
		})
	}
}

func TestDespoil(t *testing.T) {
	type args struct {
		input string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "Pad single spoiler",
			args: args{
				input: "||spoiler||",
			},
			want:    " spoiler ",
			wantErr: false,
		},
		{
			name: "Pad multiple spoilers",
			args: args{
				input: "Here are two spoilers: ||first|| and ||second||.",
			},
			want:    "Here are two spoilers:  first  and  second .",
			wantErr: false,
		},
		{
			name: "No spoilers present",
			args: args{
				input: "This string has no spoilers.",
			},
			want:    "This string has no spoilers.",
			wantErr: false,
		},
		{
			name: "Empty spoiler",
			args: args{
				input: "||||",
			},
			want:    "||||",
			wantErr: false,
		},
		{
			name: "Spoiler with spaces",
			args: args{
				input: "||  spoiler with spaces  ||",
			},
			want:    "   spoiler with spaces   ",
			wantErr: false,
		},
		{
			name: "Nested spoilers",
			args: args{
				input: "||outer ||inner|| outer||",
			},
			want:    " outer  inner  outer ",
			wantErr: false,
		},
		{
			name: "Complex",
			args: args{
				input: `https://www.youtube.com/live/aMM3PQ312L8?si=d8UBZgrEFKJB5FUI https://www.youtube.com/live/aMM3PQ312L8?si=d8UBZgrEFKJB5FUI  ||

https://www.youtube.com/live/aMM3PQ312L8?si=d8UBZgrEFKJB5FUI ||  https://www.youtube.com/live/aMM3PQ312L8?si=d8UBZgrEFKJB5FUI  ||https://www.youtube.com/live/aMM3PQ312L8?si=d8UBZgrEFKJB5FUI||  || https://www.youtube.com/live/aMM3PQ312L8?si=d8UBZgrEFKJB5FUIhttps://www.youtube.com/live/aMM3PQ312L8?si=d8UBZgrEFKJB5FUI ||  ||https://www.youtube.com/live/aMM3PQ312L8?si=d8UBZgrEFKJB5FUI https://www.youtube.com/live/aMM3PQ312L8?si=d8UBZgrEFKJB5FUI ||        ||https://www.youtube.com/live/aMM3PQ312L8?si=d8UBZgrEFKJB5FUI ||  a || https://www.youtube.com/live/aMM3PQ312L8?si=d8UBZgrEFKJB5FUI  || https://www.youtube.com/live/aMM3PQ312L8?si=d8UBZgrEFKJB5FUI ||  ||  https://www.youtube.com/live/aMM3PQ312L8?si=d8UBZgrEFKJB5FUI a  a   a ||`,
			},
			want:    `https://www.youtube.com/live/aMM3PQ312L8?si=d8UBZgrEFKJB5FUI https://www.youtube.com/live/aMM3PQ312L8?si=d8UBZgrEFKJB5FUI   

https://www.youtube.com/live/aMM3PQ312L8?si=d8UBZgrEFKJB5FUI    https://www.youtube.com/live/aMM3PQ312L8?si=d8UBZgrEFKJB5FUI   https://www.youtube.com/live/aMM3PQ312L8?si=d8UBZgrEFKJB5FUI     https://www.youtube.com/live/aMM3PQ312L8?si=d8UBZgrEFKJB5FUIhttps://www.youtube.com/live/aMM3PQ312L8?si=d8UBZgrEFKJB5FUI     https://www.youtube.com/live/aMM3PQ312L8?si=d8UBZgrEFKJB5FUI https://www.youtube.com/live/aMM3PQ312L8?si=d8UBZgrEFKJB5FUI           https://www.youtube.com/live/aMM3PQ312L8?si=d8UBZgrEFKJB5FUI    a   https://www.youtube.com/live/aMM3PQ312L8?si=d8UBZgrEFKJB5FUI    https://www.youtube.com/live/aMM3PQ312L8?si=d8UBZgrEFKJB5FUI       https://www.youtube.com/live/aMM3PQ312L8?si=d8UBZgrEFKJB5FUI a  a   a ||`,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := spoilerFinder.Replace(tt.args.input, " $1 ", -1, -1)
			if (err != nil) != tt.wantErr {
				t.Errorf("Replace() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Replace() = \"%v\", want \"%v\"", got, tt.want)
			}
		})
	}
}
