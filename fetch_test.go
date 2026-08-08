package main

import (
	"strings"
	"testing"
)

func TestStripTags(t *testing.T) {
	got := stripTags("<p>hello <b>world</b></p>")
	if got != "hello world" {
		t.Fatalf("期望 hello world, 得到 %q", got)
	}
}

func TestClean(t *testing.T) {
	got := clean("  a\n  b\t  c  ")
	if got != "a b c" {
		t.Fatalf("期望 'a b c', 得到 %q", got)
	}
}

func TestFilter(t *testing.T) {
	items := []Item{
		{Title: "Go 1.23 发布"},
		{Title: "今天天气不错"},
		{Title: "用 Go 写爬虫"},
	}
	got := filter(items, []string{"go"})
	if len(got) != 2 {
		t.Fatalf("期望命中 2 条, 得到 %d 条: %v", len(got), got)
	}
	for _, it := range got {
		if !strings.Contains(strings.ToLower(it.Title), "go") {
			t.Fatalf("过滤漏了不应命中的: %q", it.Title)
		}
	}
}

func TestParseRSS(t *testing.T) {
	body := []byte(`<?xml version="1.0"?><rss version="2.0"><channel><title>t</title>
		<item><title>甲</title><link>http://x/1</link><description>&lt;p&gt;摘要&lt;/p&gt;</description><pubDate>now</pubDate></item>
		<item><title>乙</title><link>http://x/2</link><description>纯文本</description></item>
		</channel></rss>`)
	items, err := parseFeed("src", body, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("期望 2 条, 得到 %d", len(items))
	}
	if items[0].Description != "摘要" {
		t.Fatalf("摘要去标签失败: %q", items[0].Description)
	}
}

func TestParseAtom(t *testing.T) {
	body := []byte(`<?xml version="1.0"?><feed xmlns="http://www.w3.org/2005/Atom">
		<title>t</title>
		<entry><title>甲</title><link href="http://x/1"/><summary>摘要</summary><updated>now</updated></entry>
		</feed>`)
	items, err := parseFeed("src", body, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].Link != "http://x/1" {
		t.Fatalf("Atom 解析异常: %+v", items)
	}
}
