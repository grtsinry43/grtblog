package contract

import (
	"encoding/json"
	"strings"
	"time"
)

// CreateArticleReq 创建文章请求。
type CreateArticleReq struct {
	Title        string     `json:"title" validate:"required,max=255"`
	Summary      string     `json:"summary"`
	AISummary    *string    `json:"aiSummary,omitempty"`
	LeadIn       *string    `json:"leadIn,omitempty"`
	Content      string     `json:"content" validate:"required"`
	Cover        *string    `json:"cover,omitempty"`
	CategoryID   *int64     `json:"categoryId,omitempty"`
	TagIDs       []int64    `json:"tagIds,omitempty"`
	ShortURL     *string    `json:"shortUrl"`
	IsPublished  bool       `json:"isPublished" validate:"required"`
	IsTop        bool       `json:"isTop"`
	AllowComment *bool      `json:"allowComment,omitempty"`
	IsOriginal   bool       `json:"isOriginal"`
	ExtInfo      *JSONRaw   `json:"extInfo,omitempty" swaggertype:"object"`
	CreatedAt    *time.Time `json:"createdAt,omitempty"` // 可以自定义发布时间
}

type createArticleReqJSON struct {
	Title        string   `json:"title"`
	Summary      string   `json:"summary"`
	AISummary    *string  `json:"aiSummary"`
	LeadIn       *string  `json:"leadIn"`
	Content      string   `json:"content"`
	Cover        *string  `json:"cover"`
	CategoryID   *int64   `json:"categoryId"`
	TagIDs       []int64  `json:"tagIds"`
	ShortURL     *string  `json:"shortUrl"`
	IsPublished  bool     `json:"isPublished"`
	IsTop        bool     `json:"isTop"`
	AllowComment *bool    `json:"allowComment"`
	IsOriginal   bool     `json:"isOriginal"`
	ExtInfo      *JSONRaw `json:"extInfo" swaggertype:"object"`
	CreatedAt    *string  `json:"createdAt"`
}

func (r *CreateArticleReq) UnmarshalJSON(data []byte) error {
	var aux createArticleReqJSON
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	r.Title = aux.Title
	r.Summary = aux.Summary
	r.AISummary = aux.AISummary
	r.LeadIn = aux.LeadIn
	r.Content = aux.Content
	r.Cover = aux.Cover
	r.CategoryID = aux.CategoryID
	r.TagIDs = aux.TagIDs
	r.ShortURL = aux.ShortURL
	r.IsPublished = aux.IsPublished
	r.IsTop = aux.IsTop
	r.AllowComment = aux.AllowComment
	r.IsOriginal = aux.IsOriginal
	r.ExtInfo = aux.ExtInfo

	if aux.CreatedAt == nil {
		r.CreatedAt = nil
		return nil
	}
	if strings.TrimSpace(*aux.CreatedAt) == "" {
		now := time.Now()
		r.CreatedAt = &now
		return nil
	}
	parsed, err := time.Parse(time.RFC3339, *aux.CreatedAt)
	if err != nil {
		return err
	}
	r.CreatedAt = &parsed
	return nil
}

// UpdateArticleReq 更新文章请求。
type UpdateArticleReq struct {
	Title        string   `json:"title" validate:"required,max=255"`
	Summary      string   `json:"summary"`
	AISummary    *string  `json:"aiSummary,omitempty"`
	LeadIn       *string  `json:"leadIn,omitempty"`
	Content      string   `json:"content" validate:"required"`
	Cover        *string  `json:"cover,omitempty"`
	CategoryID   *int64   `json:"categoryId,omitempty"`
	TagIDs       []int64  `json:"tagIds,omitempty"`
	ShortURL     string   `json:"shortUrl" validate:"required"`
	IsPublished  bool     `json:"isPublished"`
	IsTop        bool     `json:"isTop"`
	AllowComment *bool    `json:"allowComment,omitempty"`
	IsOriginal   bool     `json:"isOriginal"`
	ExtInfo      *JSONRaw `json:"extInfo,omitempty" swaggertype:"object"`
}

// ListArticlesReq 文章列表查询请求。
type ListArticlesReq struct {
	Page       int     `json:"page" validate:"min=1"`
	PageSize   int     `json:"pageSize" validate:"min=1,max=100"`
	CategoryID *int64  `json:"categoryId,omitempty"`
	TagID      *int64  `json:"tagId,omitempty"`
	AuthorID   *int64  `json:"authorId,omitempty"`
	Published  *bool   `json:"published,omitempty"`
	Search     *string `json:"search,omitempty"`
}

// CheckArticleLatestReq 文章版本校验请求。
type CheckArticleLatestReq struct {
	Hash string `json:"hash"`
}

// BatchSetArticlePublishedReq 批量切换文章发布状态请求。
type BatchSetArticlePublishedReq struct {
	IDs         []int64 `json:"ids"`
	IsPublished bool    `json:"isPublished"`
}

// BatchSetArticleTopReq 批量切换文章置顶状态请求。
type BatchSetArticleTopReq struct {
	IDs   []int64 `json:"ids"`
	IsTop bool    `json:"isTop"`
}

// BatchDeleteArticleReq 批量删除文章请求。
type BatchDeleteArticleReq struct {
	IDs []int64 `json:"ids"`
}
