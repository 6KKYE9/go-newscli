package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// filter 按关键词过滤，命中标题里任意一个词就保留。关键词不区分大小写。
func filter(items []Item, kws []string) []Item {
	lower := make([]string, len(kws))
	for i, k := range kws {
		lower[i] = strings.ToLower(k)
	}
	var out []Item
	for _, it := range items {
		t := strings.ToLower(it.Title)
		for _, k := range lower {
			if strings.Contains(t, k) {
				out = append(out, it)
				break
			}
		}
	}
	return out
}

// printItems 在终端按来源分组打印，方便扫一眼。
func printItems(items []Item) {
	if len(items) == 0 {
		fmt.Println("没有抓到任何条目。")
		return
	}
	bySource := map[string][]Item{}
	order := []string{}
	for _, it := range items {
		if _, ok := bySource[it.Source]; !ok {
			order = append(order, it.Source)
		}
		bySource[it.Source] = append(bySource[it.Source], it)
	}

	for _, src := range order {
		fmt.Printf("\n=== %s ===\n", src)
		for i, it := range bySource[src] {
			fmt.Printf("%d. %s\n", i+1, it.Title)
			if it.Description != "" {
				fmt.Printf("   %s\n", clip(it.Description, 120))
			}
			if it.Link != "" {
				fmt.Printf("   %s\n", it.Link)
			}
		}
	}
	fmt.Printf("\n共 %d 条\n", len(items))
}

func clip(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…"
}

func saveJSON(path string, items []Item) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(items)
}
