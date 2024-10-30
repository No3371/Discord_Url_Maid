package main

import (
	"reflect"
	"testing"
)

func TestTryCleanString(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name           string
		args           args
		wantUrlMap     []processedUrl
		wantCleaned    int
		wantRedirects  int
		wantMasks      int
		wantNotUrlOnly bool
		wantErr        bool
	}{
		{
			name: "NotUrlOnlySpoiler",
			args: args{
				str: `https://www.youtube.com/live/aMM3PQ312L8?si=d8UBZgrEFKJB5FUI https://www.youtube.com/live/aMM3PQ312L8?si=d8UBZgrEFKJB5FUI  ||

https://www.youtube.com/live/aMM3PQ312L8?si=d8UBZgrEFKJB5FUI ||  https://www.youtube.com/live/aMM3PQ312L8?si=d8UBZgrEFKJB5FUI  ||https://www.youtube.com/live/aMM3PQ312L8?si=d8UBZgrEFKJB5FUI||  || https://www.youtube.com/live/aMM3PQ312L8?si=d8UBZgrEFKJB5FUIhttps://www.youtube.com/live/aMM3PQ312L8?si=d8UBZgrEFKJB5FUI ||  ||https://www.youtube.com/live/aMM3PQ312L8?si=d8UBZgrEFKJB5FUI https://www.youtube.com/live/aMM3PQ312L8?si=d8UBZgrEFKJB5FUI ||        ||https://www.youtube.com/live/aMM3PQ312L8?si=d8UBZgrEFKJB5FUI ||  a || https://www.youtube.com/live/aMM3PQ312L8?si=d8UBZgrEFKJB5FUI  || https://www.youtube.com/live/aMM3PQ312L8?si=d8UBZgrEFKJB5FUI ||  ||  https://www.youtube.com/live/aMM3PQ312L8?si=d8UBZgrEFKJB5FUI a  a   a ||`,
			},
			wantUrlMap: []processedUrl{
				{
					Raw:       "https://www.youtube.com/live/aMM3PQ312L8?si=d8UBZgrEFKJB5FUI",
					Processed: "https://www.youtube.com/live/aMM3PQ312L8",
					IsSpoiler: true,
				},
			},
			wantCleaned:    1,
			wantRedirects:  0,
			wantMasks:      0,
			wantNotUrlOnly: true,
			wantErr:        false,
		},
		{
			name: "NotUrlOnlySpoiler",
			args: args{
				str: `https://www.youtube.com/live/aMM3PQ312L8?si=d8UBZgrEFKJB5FUI https://www.youtube.com/live/aMM3PQ312L8?si=d8UBZgrEFKJB5FUI  ||

https://www.youtube.com/live/aMM3PQ312L8?si=d8UBZgrEFKJB5FUI ||  https://www.youtube.com/live/aMM3PQ312L8?si=d8UBZgrEFKJB5FUI  ||https://www.youtube.com/live/aMM3PQ312L8?si=d8UBZgrEFKJB5FUI||  || https://www.youtube.com/live/aMM3PQ312L8?si=d8UBZgrEFKJB5FUIhttps://www.youtube.com/live/aMM3PQ312L8?si=d8UBZgrEFKJB5FUI ||  ||https://www.youtube.com/live/aMM3PQ312L8?si=d8UBZgrEFKJB5FUI https://www.youtube.com/live/aMM3PQ312L8?si=d8UBZgrEFKJB5FUI ||        ||https://www.youtube.com/live/aMM3PQ312L8?si=d8UBZgrEFKJB5FUI ||  a || https://www.youtube.com/live/aMM3PQ312L8?si=d8UBZgrEFKJB5FUI  || https://www.youtube.com/live/aMM3PQ312L8?si=d8UBZgrEFKJB5FUI ||  ||  https://www.youtube.com/live/aMM3PQ312L8?si=d8UBZgrEFKJB5FUI ||`,
			},
			wantUrlMap: []processedUrl{
				{
					Raw:       "https://www.youtube.com/live/aMM3PQ312L8?si=d8UBZgrEFKJB5FUI",
					Processed: "https://www.youtube.com/live/aMM3PQ312L8",
					IsSpoiler: true,
				},
			},
			wantCleaned:    1,
			wantRedirects:  0,
			wantMasks:      0,
			wantNotUrlOnly: true,
			wantErr:        false,
		},
		{
			name: "UrlOnlySpoiler",
			args: args{
				str: `https://www.youtube.com/live/aMM3PQ312L8?si=d8UBZgrEFKJB5FUI https://www.youtube.com/live/aMM3PQ312L8?si=d8UBZgrEFKJB5FUI  ||

https://www.youtube.com/live/aMM3PQ312L8?si=d8UBZgrEFKJB5FUI ||  https://www.youtube.com/live/aMM3PQ312L8?si=d8UBZgrEFKJB5FUI  ||https://www.youtube.com/live/aMM3PQ312L8?si=d8UBZgrEFKJB5FUI||  || https://www.youtube.com/live/aMM3PQ312L8?si=d8UBZgrEFKJB5FUIhttps://www.youtube.com/live/aMM3PQ312L8?si=d8UBZgrEFKJB5FUI ||  ||https://www.youtube.com/live/aMM3PQ312L8?si=d8UBZgrEFKJB5FUI https://www.youtube.com/live/aMM3PQ312L8?si=d8UBZgrEFKJB5FUI ||        ||https://www.youtube.com/live/aMM3PQ312L8?si=d8UBZgrEFKJB5FUI ||   || https://www.youtube.com/live/aMM3PQ312L8?si=d8UBZgrEFKJB5FUI  || https://www.youtube.com/live/aMM3PQ312L8?si=d8UBZgrEFKJB5FUI ||  ||  https://www.youtube.com/live/aMM3PQ312L8?si=d8UBZgrEFKJB5FUI`,
			},
			wantUrlMap: []processedUrl{
				{
					Raw:       "https://www.youtube.com/live/aMM3PQ312L8?si=d8UBZgrEFKJB5FUI",
					Processed: "https://www.youtube.com/live/aMM3PQ312L8",
					IsSpoiler: true,
				},
			},
			wantCleaned:    1,
			wantRedirects:  0,
			wantMasks:      0,
			wantNotUrlOnly: false,
			wantErr:        false,
		},
		{
			name: "test",
			args: args{
				str: `到底要多久https://x.com/horo_27/status/1845408056445972628?s=19
https://twitcasting.tv/kurokumo_01?t=你好 a
https://www.youtube.com/watch?v=qQiVUv7RIPs&t=770
https://www.youtube.com/live/5VL4lFPQuc4?si=h2GlP0Dxjn23UiML
https://news.ltn.com.tw/news/life/breakingnews/4826075?fbclid=IwZXh0bgNhZW0CMTEAAR21sLbgLCKNGg1qFqOHPkGnKiINqzN3MyT1gtfuBY6Tlph-iIu06J5bgD4_aem_9oBjNcuqObVpJ-8towvPIA&prev=1 

到底要多久`,
			},
			wantUrlMap: []processedUrl{
				{
					Raw:       "https://x.com/horo_27/status/1845408056445972628?s=19",
					Processed: "https://x.com/horo_27/status/1845408056445972628",
					IsSpoiler: false,
				}, // V
				{
					Raw:       "https://twitcasting.tv/kurokumo_01?t=你好",
					Processed: "https://twitcasting.tv/kurokumo_01?t=你好",
					IsSpoiler: false,
				}, // X
				{
					Raw:       "https://www.youtube.com/watch?v=qQiVUv7RIPs&t=770",
					Processed: "https://www.youtube.com/watch?v=qQiVUv7RIPs&t=770",
					IsSpoiler: false,
				}, // X
				{
					Raw:       "https://www.youtube.com/live/5VL4lFPQuc4?si=h2GlP0Dxjn23UiML",
					Processed: "https://www.youtube.com/live/5VL4lFPQuc4",
					IsSpoiler: false,
				}, // V
				{
					Raw:       "https://news.ltn.com.tw/news/life/breakingnews/4826075?fbclid=IwZXh0bgNhZW0CMTEAAR21sLbgLCKNGg1qFqOHPkGnKiINqzN3MyT1gtfuBY6Tlph-iIu06J5bgD4_aem_9oBjNcuqObVpJ-8towvPIA&prev=1",
					Processed: "https://news.ltn.com.tw/news/life/breakingnews/4826075?&prev=1",
					IsSpoiler: false,
				}, // V
			},
			wantCleaned:    3,
			wantRedirects:  0,
			wantMasks:      0,
			wantNotUrlOnly: true,
			wantErr:        false,
		},
		{
			name: "Redirect",
			args: args{
				str: `https://www.youtube.com/redirect?event=video_description&redir_token=QUFFLUhqbUlwZ3hybmEyZnd5bnpTR0N5VWFnN3J4MFE1Z3xBQ3Jtc0trY2tQMzA1NDdCcnphVm5oMGlfYVB1TU5VYjZaYVZSUGFzak1hLTJ2SGN1MkZCdmx1VU9zY1l3Tl91cXpuc19yVTBZYVhNTGdzMEtDaUJjX0lXaHJSYUtvdFNiQjBGV0NkRzBvUjZXejhFblVIRV93OA&q=https%3A%2F%2Fx.com%2Fi%2Fspaces%2F1lPKqOyrXWLJb&v=eqVjAWxlxbk`,
			},
			wantUrlMap: []processedUrl{
				{
					Raw:        "https://www.youtube.com/redirect?event=video_description&redir_token=QUFFLUhqbUlwZ3hybmEyZnd5bnpTR0N5VWFnN3J4MFE1Z3xBQ3Jtc0trY2tQMzA1NDdCcnphVm5oMGlfYVB1TU5VYjZaYVZSUGFzak1hLTJ2SGN1MkZCdmx1VU9zY1l3Tl91cXpuc19yVTBZYVhNTGdzMEtDaUJjX0lXaHJSYUtvdFNiQjBGV0NkRzBvUjZXejhFblVIRV93OA&q=https%3A%2F%2Fx.com%2Fi%2Fspaces%2F1lPKqOyrXWLJb&v=eqVjAWxlxbk",
					Processed:  "https://www.youtube.com/redirect?event=video_description&redir_token=QUFFLUhqbUlwZ3hybmEyZnd5bnpTR0N5VWFnN3J4MFE1Z3xBQ3Jtc0trY2tQMzA1NDdCcnphVm5oMGlfYVB1TU5VYjZaYVZSUGFzak1hLTJ2SGN1MkZCdmx1VU9zY1l3Tl91cXpuc19yVTBZYVhNTGdzMEtDaUJjX0lXaHJSYUtvdFNiQjBGV0NkRzBvUjZXejhFblVIRV93OA&q=https%3A%2F%2Fx.com%2Fi%2Fspaces%2F1lPKqOyrXWLJb&v=eqVjAWxlxbk",
					IsSpoiler:  false,
					IsRedirect: true,
				},
			},
			wantCleaned:    0,
			wantRedirects:  1,
			wantMasks:      0,
			wantNotUrlOnly: false,
			wantErr:        false,
		},
		{
			name: "Masks",
			args: args{
				str: `[ 123 ](https://123.com)`,
			},
			wantUrlMap: []processedUrl{
				{
					Raw:        "https://123.com",
					Processed:  "https://123.com",
					IsSpoiler:  false,
					IsRedirect: false,
					Mask: " 123 ",
				},
			},
			wantCleaned:    0,
			wantRedirects:  0,
			wantMasks:      1,
			wantNotUrlOnly: true,
			wantErr:        false,
		},
	}
	providers, err := FetchAndLoadJSON(repo)
	if err != nil {
		t.Fatalf("FetchAndLoadJSON() error = %v", err)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotUrlMap, gotCleaned, gotRedirects, gotMasks, gotNotUrlOnly, err := TryCleanString(tt.args.str, providers)
			if (err != nil) != tt.wantErr {
				t.Errorf("TryCleanString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotUrlMap, tt.wantUrlMap) {
				t.Errorf("TryCleanString() gotUrlMap = %v, want %v", gotUrlMap, tt.wantUrlMap)
			}
			if gotCleaned != tt.wantCleaned {
				t.Errorf("TryCleanString() gotCleaned = %v, want %v", gotCleaned, tt.wantCleaned)
			}
			if gotRedirects != tt.wantRedirects {
				t.Errorf("TryCleanString() gotContainsRedirect = %v, want %v", gotRedirects, tt.wantRedirects)
			}
			if gotMasks != tt.wantMasks {
				t.Errorf("TryCleanString() gotMasks = %v, want %v", gotMasks, tt.wantMasks)
			}
			if gotNotUrlOnly != tt.wantNotUrlOnly {
				t.Errorf("TryCleanString() gotNotUrlOnly = %v, want %v", gotNotUrlOnly, tt.wantNotUrlOnly)
			}
		})
	}
}

func TestPrepareReply(t *testing.T) {
	type args struct {
		urlMap []processedUrl
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test",
			args: args{
				urlMap: []processedUrl{
					{
						Raw:       "https://x.com/horo_27/status/1845408056445972628?s=19",
						Processed: "https://x.com/horo_27/status/1845408056445972628",
						IsSpoiler: false,
					},
					{
						Raw:       "https://twitcasting.tv/kurokumo_01?t=你好",
						Processed: "https://twitcasting.tv/kurokumo_01?t=你好",
						IsSpoiler: false,
					},
					{
						Raw:       "https://www.youtube.com/watch?v=qQiVUv7RIPs&t=770",
						Processed: "https://www.youtube.com/watch?v=qQiVUv7RIPs&t=770",
						IsSpoiler: false,
					},
					{
						Raw:       "https://www.youtube.com/live/5VL4lFPQuc4?si=h2GlP0Dxjn23UiML",
						Processed: "https://www.youtube.com/live/5VL4lFPQuc4",
						IsSpoiler: false,
					},
					{
						Raw:       "https://news.ltn.com.tw/news/life/breakingnews/4826075?fbclid=IwZXh0bgNhZW0CMTEAAR21sLbgLCKNGg1qFqOHPkGnKiINqzN3MyT1gtfuBY6Tlph-iIu06J5bgD4_aem_9oBjNcuqObVpJ-8towvPIA&prev=1",
						Processed: "https://news.ltn.com.tw/news/life/breakingnews/4826075?&prev=1",
						IsSpoiler: false,
					},
				},
			},
			want: `https://x.com/horo_27/status/1845408056445972628
https://twitcasting.tv/kurokumo_01?t=你好
https://www.youtube.com/watch?v=qQiVUv7RIPs&t=770
https://www.youtube.com/live/5VL4lFPQuc4
https://news.ltn.com.tw/news/life/breakingnews/4826075?&prev=1`,
		},
		{
			name: "redirect",
			args: args{
				urlMap: []processedUrl{
					{
						Raw: "https://www.youtube.com/redirect?event=video_description&redir_token=QUFFLUhqbUlwZ3hybmEyZnd5bnpTR0N5VWFnN3J4MFE1Z3xBQ3Jtc0trY2tQMzA1NDdCcnphVm5oMGlfYVB1TU5VYjZaYVZSUGFzak1hLTJ2SGN1MkZCdmx1VU9zY1l3Tl91cXpuc19yVTBZYVhNTGdzMEtDaUJjX0lXaHJSYUtvdFNiQjBGV0NkRzBvUjZXejhFblVIRV93OA&q=https%3A%2F%2Fx.com%2Fi%2Fspaces%2F1lPKqOyrXWLJb&v=eqVjAWxlxbk",
						Processed:  "https://www.youtube.com/redirect?event=video_description&redir_token=QUFFLUhqbUlwZ3hybmEyZnd5bnpTR0N5VWFnN3J4MFE1Z3xBQ3Jtc0trY2tQMzA1NDdCcnphVm5oMGlfYVB1TU5VYjZaYVZSUGFzak1hLTJ2SGN1MkZCdmx1VU9zY1l3Tl91cXpuc19yVTBZYVhNTGdzMEtDaUJjX0lXaHJSYUtvdFNiQjBGV0NkRzBvUjZXejhFblVIRV93OA&q=https%3A%2F%2Fx.com%2Fi%2Fspaces%2F1lPKqOyrXWLJb&v=eqVjAWxlxbk",
						IsSpoiler:  false,
						IsRedirect: true,
					},
				},
			},
			want: `↪️ Redirect Found / 可能自動轉向未知站點`,
		},
		{
			name: "redirect+clean",
			args: args{
				urlMap: []processedUrl{
					{
						Raw:       "https://x.com/horo_27/status/1845408056445972628?s=19",
						Processed: "https://x.com/horo_27/status/1845408056445972628",
						IsSpoiler: false,
					},
					{
						Raw:       "https://twitcasting.tv/kurokumo_01?t=你好",
						Processed: "https://twitcasting.tv/kurokumo_01?t=你好",
						IsSpoiler: false,
					},
					{
						Raw:       "https://www.youtube.com/live/5VL4lFPQuc4?si=h2GlP0Dxjn23UiML",
						Processed: "https://www.youtube.com/live/5VL4lFPQuc4",
						IsSpoiler: true,
					},
					{
						Raw: "https://www.youtube.com/redirect?event=video_description&redir_token=QUFFLUhqbUlwZ3hybmEyZnd5bnpTR0N5VWFnN3J4MFE1Z3xBQ3Jtc0trY2tQMzA1NDdCcnphVm5oMGlfYVB1TU5VYjZaYVZSUGFzak1hLTJ2SGN1MkZCdmx1VU9zY1l3Tl91cXpuc19yVTBZYVhNTGdzMEtDaUJjX0lXaHJSYUtvdFNiQjBGV0NkRzBvUjZXejhFblVIRV93OA&q=https%3A%2F%2Fx.com%2Fi%2Fspaces%2F1lPKqOyrXWLJb&v=eqVjAWxlxbk",
						Processed:  "https://www.youtube.com/redirect?event=video_description&redir_token=QUFFLUhqbUlwZ3hybmEyZnd5bnpTR0N5VWFnN3J4MFE1Z3xBQ3Jtc0trY2tQMzA1NDdCcnphVm5oMGlfYVB1TU5VYjZaYVZSUGFzak1hLTJ2SGN1MkZCdmx1VU9zY1l3Tl91cXpuc19yVTBZYVhNTGdzMEtDaUJjX0lXaHJSYUtvdFNiQjBGV0NkRzBvUjZXejhFblVIRV93OA&q=https%3A%2F%2Fx.com%2Fi%2Fspaces%2F1lPKqOyrXWLJb&v=eqVjAWxlxbk",
						IsSpoiler:  false,
						IsRedirect: true,
					},
				},
			},
			want: `https://x.com/horo_27/status/1845408056445972628
https://twitcasting.tv/kurokumo_01?t=你好
||https://www.youtube.com/live/5VL4lFPQuc4||
https://www.youtube.com/redirect?event=video_description&redir_token=QUFFLUhqbUlwZ3hybmEyZnd5bnpTR0N5VWFnN3J4MFE1Z3xBQ3Jtc0trY2tQMzA1NDdCcnphVm5oMGlfYVB1TU5VYjZaYVZSUGFzak1hLTJ2SGN1MkZCdmx1VU9zY1l3Tl91cXpuc19yVTBZYVhNTGdzMEtDaUJjX0lXaHJSYUtvdFNiQjBGV0NkRzBvUjZXejhFblVIRV93OA&q=https%3A%2F%2Fx.com%2Fi%2Fspaces%2F1lPKqOyrXWLJb&v=eqVjAWxlxbk ↪️ Redirect Found / 可能自動轉向未知站點`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PrepareReply(tt.args.urlMap); got != tt.want {
				t.Errorf("PrepareReply() = \n%v\n, want \n%v", got, tt.want)
			}
		})
	}
}

func TestCleanUrl(t *testing.T) {
	tests := []struct {
		name    string
		message string
		want    string
	}{
		{"test", "https://x.com/4009_0825900/status/1840979404572213471?t=iy49kBSlrMutQ0QwNW4YyA&s=19", "https://x.com/4009_0825900/status/1840979404572213471"},
		{"test", "https://youtu.be/aVpJGGQHSqc?si=az72VbWhlionVl4c", "https://youtu.be/aVpJGGQHSqc"},
		{"test", "https://www.youtube.com/watch?v=n-su1KVKlGk", "https://www.youtube.com/watch?v=n-su1KVKlGk"},
		{"test", "https://www.reddit.com/r/MechanicalKeyboards/comments/156he48/attention_new_issue_with_gmk_keycaps_know_before/", "https://www.reddit.com/r/MechanicalKeyboards/comments/156he48/attention_new_issue_with_gmk_keycaps_know_before/"},
		{"test", "https://travel.ettoday.net/amp/amp_news.php7?news_id=2738515&ref=mw&from=google.com", "https://travel.ettoday.net/amp/amp_news.php7?news_id=2738515"},
		{"test", "https://www.china-airlines.com/zh-tw/tpe_20240906_autumn?gad_source=1&gclid=Cj0KCQjwo8S3BhDeARIsAFRmkOPMStVGw370iMxB8F3mTp9CB2ZgPRWsc2B1I5rkczwm_fcGrYrkD5EaAmDHEALw_wcB", "https://www.china-airlines.com/zh-tw/tpe_20240906_autumn"},
		{"test", "https://www.tomtoc.com.tw/t21s1d2?_gl=1*1j8iyju*_up*MQ..&gclid=EAIaIQobChMImJvOisbaiAMVYcRMAh3OZSGtEAAYASAAEgJcVPD_BwE", "https://www.tomtoc.com.tw/t21s1d2"},
		{"test", "https://cathaybk.com.tw/cathaybk/personal/product/credit-card/cards/eva/?CUB_SRC=GOOGLE&CUB_CHL1=AD_WORD&CUB_CHL2=01&MA_TK=DB590&CUB_DT=20240101&Cub_ProjectCode=DBB4400001&gad_source=1&gclid=Cj0KCQjw9Km3BhDjARIsAGUb4nygrkAZpfoCJo3YVMkZsSfpMtF8I2aoAy22EAOp8REOOeSlSd5r5d0aAk6zEALw_wcB", "https://cathaybk.com.tw/cathaybk/personal/product/credit-card/cards/eva/?CUB_SRC=GOOGLE&CUB_CHL1=AD_WORD&CUB_CHL2=01&MA_TK=DB590&CUB_DT=20240101&Cub_ProjectCode=DBB4400001"},
		{"test", "https://m.momoshop.com.tw/goods.momo?i_code=10489628&osm=Ad07&utm_source=googleshop&utm_medium=googleshop-pmax-all-mb-feed&utm_content=bn&gclid=Cj0KCQjwwae1BhC_ARIsAK4Jfrw5xTuyBtdUOafMEZCD3bV0d6H77it_bp5Zi0UrzXjK69ztk3Z_hXgaAiFVEALw_wcB", "https://m.momoshop.com.tw/goods.momo?i_code=10489628"},
	}
	providers, err := FetchAndLoadJSON(repo)
	if err != nil {
		t.Fatalf("FetchAndLoadJSON() error = %v", err)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := CleanUrl(tt.message, providers); got != tt.want {
				t.Errorf("Got= %v, want %v", got, tt.want)
			}
		})
	}
}
