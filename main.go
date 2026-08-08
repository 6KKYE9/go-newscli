package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

// 内置几个公开的新闻 RSS 源，都是不用鉴权就能拉的中文源。
var defaultFeeds = []string{
	"https://www.zhihu.com/rss",
	"https://sspai.com/feed",
	"https://www.ruanyifeng.com/blog/atom.xml",
}

func main() {
	var (
		feeds   = flag.String("feeds", "", "逗号分隔的 RSS 地址，不填用内置几个源")
		keyword = flag.String("kw", "", "按关键词过滤标题，多个词用空格分开，命中任意一个就留")
		limit   = flag.Int("n", 10, "每个源最多展示几条")
		save    = flag.String("save", "", "把结果写到这个文件（JSON）")
		timeout = flag.Int("timeout", 10, "单个源拉取超时秒数")
	)
	flag.Parse()

	urls := defaultFeeds
	if strings.TrimSpace(*feeds) != "" {
		urls = strings.Split(*feeds, ",")
		for i := range urls {
			urls[i] = strings.TrimSpace(urls[i])
		}
	}

	kws := strings.Fields(*keyword)

	items, err := fetchAll(urls, *timeout, *limit)
	if err != nil {
		fmt.Fprintln(os.Stderr, "拉取出错:", err)
		os.Exit(1)
	}

	if len(kws) > 0 {
		items = filter(items, kws)
	}

	if *save != "" {
		if err := saveJSON(*save, items); err != nil {
			fmt.Fprintln(os.Stderr, "保存失败:", err)
			os.Exit(1)
		}
		fmt.Printf("已保存 %d 条到 %s\n", len(items), *save)
		return
	}

	printItems(items)
}
