package main

import "testing"

func TestMain(m *testing.M) {
	m.Run()
}

func TestCleanMessage(t *testing.T) {
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
		{"test", "https://www.tomtoc.com.tw/t21s1d2", "https://www.tomtoc.com.tw/t21s1d2"},
		{"test", "https://cathaybk.com.tw/cathaybk/personal/product/credit-card/cards/eva/?CUB_SRC=GOOGLE&CUB_CHL1=AD_WORD&CUB_CHL2=01&MA_TK=DB590&CUB_DT=20240101&Cub_ProjectCode=DBB4400001&gad_source=1&gclid=Cj0KCQjw9Km3BhDjARIsAGUb4nygrkAZpfoCJo3YVMkZsSfpMtF8I2aoAy22EAOp8REOOeSlSd5r5d0aAk6zEALw_wcB", "https://cathaybk.com.tw/cathaybk/personal/product/credit-card/cards/eva/?CUB_SRC=GOOGLE&CUB_CHL1=AD_WORD&CUB_CHL2=01&MA_TK=DB590&CUB_DT=20240101&Cub_ProjectCode=DBB4400001"},
		{"test", "https://m.momoshop.com.tw/goods.momo?i_code=10489628&osm=Ad07&utm_source=googleshop&utm_medium=googleshop-pmax-all-mb-feed&utm_content=bn&gclid=Cj0KCQjwwae1BhC_ARIsAK4Jfrw5xTuyBtdUOafMEZCD3bV0d6H77it_bp5Zi0UrzXjK69ztk3Z_hXgaAiFVEALw_wcB", "https://m.momoshop.com.tw/goods.momo?i_code=10489628&osm=Ad07"},
	}
	repo, err := FetchAndLoadJSON(repo)
	if err != nil {
		t.Fatalf("FetchAndLoadJSON() error = %v", err)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CleanMessageAndExtractCleanedUrls(tt.message, repo); got != tt.want {
				t.Errorf("CleanMessage() = %v, want %v", got, tt.want)
			}
		})
	}
}
