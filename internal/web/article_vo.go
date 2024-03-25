package web

type ArticleVo struct {
	Id         int64  `json:"id,omitempty"`
	Title      string `json:"title,omitempty"`
	Content    string `json:"content,omitempty"`
	Abstract   string `json:"abstract,omitempty"`
	AuthorId   int64  `json:"authorId,omitempty"`
	AuthorName string `json:"authorName,omitempty"`
	Status     uint8  `json:"status,omitempty"`
	Ctime      string `json:"ctime,omitempty"`
	Utime      string `json:"utime,omitempty"`
	ReadCnt    int64  `json:"read_cnt,omitempty"`
	LikeCnt    int64  `json:"like_cnt,omitempty"`
	CollectCnt int64  `json:"collect_cnt,omitempty"`
	Liked      bool   `json:"liked,omitempty"`
	Collected  bool   `json:"collected,omitempty"`
}
