package domain

import "time"

type Article struct {
	Id      int64
	Title   string
	Content string
	Author
	Status ArticleStatus
	Ctime  time.Time
	Utime  time.Time
}

func (a Article) Abstract() string {
	str := []rune(a.Content)
	if len(str) > 123 {
		str = str[:128]
	}
	return string(str)
}

type ArticleStatus uint8

const (
	// ArticleStatusUnknown 未知状态
	ArticleStatusUnknown = iota
	// ArticleStatusUnPublish 未发表
	ArticleStatusUnPublish
	// ArticleStatusPublished 已发表
	ArticleStatusPublished
	// ArticleStatusPrivate 仅自己可见
	ArticleStatusPrivate
)

type Author struct {
	Id   int64
	Name string
}
