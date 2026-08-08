# go-newscli

天天刷好几个新闻源，其实挺烦的。这个工具一次把几个 RSS 源拉下来，归一成统一的格式，按关键词筛一下，要么直接打印要么存成 JSON 慢慢看。纯 Go 标准库，没装任何依赖。

## 直接跑

```powershell
# 用内置的几个源（知乎/少数派/阮一峰），每个源取前 10 条
go run .

# 只看跟 Go 有关的
go run . -kw go

# 多关键词，命中任意一个就留
go run . -kw "go 区块链 ai"

# 指定自己的源
go run . -feeds "https://a.com/rss,https://b.com/atom.xml" -n 5

# 存成 JSON 以后看
go run . -save news.json
```

## 参数

| 参数 | 默认值 | 说明 |
|------|--------|------|
| `-feeds` | 内置几个源 | 逗号分隔的 RSS/Atom 地址 |
| `-kw` | 空 | 关键词过滤，空格分开，命中标题任意一个就留 |
| `-n` | 10 | 每个源最多取几条 |
| `-save` | 空 | 结果写进这个 JSON 文件 |
| `-timeout` | 10 | 单个源拉取超时（秒） |

## 怎么解析的

RSS 和 Atom 字段名差很多，统一归一成 `Item{Title, Link, Description, Source, Published}`。先按 RSS 2.0 解，解不出来再按 Atom 解。摘要里常带着 `<p>` 这种标签，会先 strip 掉再收一下空白，打印出来才干净。

单个源拉挂了不影响其他的，会跳过并打印原因，最后把能拿到的都汇总回来。

## 没做的事

- 只支持 RSS/Atom，那种纯 HTML 列表页（没有 feed）的源抓不了，得自己写选择器
- 没做增量：每次都全量拉，不记上次看到哪
- 没落库：`-save` 只是覆盖写 JSON，不会去重

## 测试

```powershell
go test ./...
```

RSS/Atom 解析、去标签、关键词过滤都用手写样例测了。
