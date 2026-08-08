package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// Item 是一条新闻的归一化结构。RSS 和 Atom 字段不一样，解析时都归一成这个。
type Item struct {
	Title       string
	Link        string
	Description string
	Source      string
	Published   string
}

// rssFeed 覆盖 RSS 2.0 的结构。
type rssFeed struct {
	Channel struct {
		Title string `xml:"title"`
		Items []struct {
			Title       string `xml:"title"`
			Link        string `xml:"link"`
			Description string `xml:"description"`
			PubDate     string `xml:"pubDate"`
		} `xml:"item"`
	} `xml:"channel"`
}

// atomFeed 覆盖 Atom 的结构，字段名和 RSS 差挺多。
type atomFeed struct {
	Title string `xml:"title"`
	Entry []struct {
		Title   string `xml:"title"`
		Link    []link `xml:"link"`
		Summary string `xml:"summary"`
		Updated string `xml:"updated"`
	} `xml:"entry"`
}

type link struct {
	Href string `xml:"href,attr"`
}

// fetchAll 拉一批源，单个源出错不影响其他的，最后汇总返回。
func fetchAll(urls []string, timeoutSec, perLimit int) ([]Item, error) {
	client := &http.Client{Timeout: time.Duration(timeoutSec) * time.Second}
	var out []Item
	var firstErr error

	for _, u := range urls {
		items, err := fetchOne(client, u, perLimit)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			fmt.Fprintf(os.Stderr, "跳过 %s: %v\n", u, err)
			continue
		}
		out = append(out, items...)
	}
	return out, firstErr
}

func fetchOne(client *http.Client, u string, perLimit int) ([]Item, error) {
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}
	// 不少源会按 UA 决定返回内容，不带 UA 容易被挡。
	req.Header.Set("User-Agent", "go-newscli/1.0 (+https://github.com/6KKYE9/go-newscli)")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("状态码 %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return nil, err
	}

	return parseFeed(u, body, perLimit)
}

// parseFeed 先试 RSS，失败再试 Atom。两种格式都兜底成 Item。
func parseFeed(source string, body []byte, perLimit int) ([]Item, error) {
	var rss rssFeed
	if err := xml.Unmarshal(body, &rss); err == nil && rss.Channel.Title != "" {
		n := perLimit
		if n > len(rss.Channel.Items) {
			n = len(rss.Channel.Items)
		}
		out := make([]Item, 0, n)
		for _, it := range rss.Channel.Items[:n] {
			out = append(out, Item{
				Title:       clean(it.Title),
				Link:        it.Link,
				Description: clean(stripTags(it.Description)),
				Source:      source,
				Published:   it.PubDate,
			})
		}
		return out, nil
	}

	var atom atomFeed
	if err := xml.Unmarshal(body, &atom); err == nil && atom.Title != "" {
		n := perLimit
		if n > len(atom.Entry) {
			n = len(atom.Entry)
		}
		out := make([]Item, 0, n)
		for _, e := range atom.Entry[:n] {
			link := ""
			if len(e.Link) > 0 {
				link = e.Link[0].Href
			}
			out = append(out, Item{
				Title:       clean(e.Title),
				Link:        link,
				Description: clean(stripTags(e.Summary)),
				Source:      source,
				Published:   e.Updated,
			})
		}
		return out, nil
	}

	return nil, fmt.Errorf("既不是 RSS 也不是 Atom")
}

// stripTags 把 HTML 标签去掉，不然摘要里会带着一堆 <p> 之类的。
func stripTags(s string) string {
	var b strings.Builder
	inTag := false
	for _, r := range s {
		switch {
		case r == '<':
			inTag = true
		case r == '>':
			inTag = false
		case !inTag:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// clean 收一下空白，去掉首尾和中间多余换行。
func clean(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\t", " ")
	for strings.Contains(s, "  ") {
		s = strings.ReplaceAll(s, "  ", " ")
	}
	return s
}
