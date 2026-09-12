package client

import "time"

// User 当前登录用户。
type User struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Email    string `json:"email"`
	Avatar   string `json:"avatar"`
	IsActive bool   `json:"isActive"`
	IsAdmin  bool   `json:"isAdmin"`
}

// Tag 标签。
type Tag struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	ArticleCount int64  `json:"articleCount,omitempty"`
}

// Category 文章分类。
type Category struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	ShortURL  string `json:"shortUrl"`
	CreatedAt string `json:"createdAt"`
}

// Column 手记专栏。
type Column struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	ShortURL  string `json:"shortUrl"`
	CreatedAt string `json:"createdAt"`
}

// ArticleListItem 文章列表项。
type ArticleListItem struct {
	ID           int64     `json:"id"`
	Title        string    `json:"title"`
	ShortURL     string    `json:"shortUrl"`
	Summary      string    `json:"summary"`
	Views        int64     `json:"views"`
	CategoryName string    `json:"categoryName"`
	Tags         []string  `json:"tags"`
	Likes        int       `json:"likes"`
	Comments     int       `json:"comments"`
	IsTop        bool      `json:"isTop"`
	IsPublished  bool      `json:"isPublished"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// ArticleList 文章分页列表。
type ArticleList struct {
	Items []ArticleListItem `json:"items"`
	Total int64             `json:"total"`
	Page  int               `json:"page"`
	Size  int               `json:"size"`
}

// Article 文章详情。
type Article struct {
	ID           int64     `json:"id"`
	Title        string    `json:"title"`
	Summary      string    `json:"summary"`
	Content      string    `json:"content"`
	Cover        *string   `json:"cover"`
	CategoryID   *int64    `json:"categoryId"`
	CategoryName string    `json:"categoryName"`
	ShortURL     string    `json:"shortUrl"`
	IsPublished  bool      `json:"isPublished"`
	IsTop        bool      `json:"isTop"`
	AllowComment bool      `json:"allowComment"`
	IsOriginal   bool      `json:"isOriginal"`
	Tags         []Tag     `json:"tags"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// ArticleUpsertReq 创建/更新文章请求（更新时 ShortURL 必填）。
type ArticleUpsertReq struct {
	Title        string  `json:"title"`
	Summary      string  `json:"summary"`
	Content      string  `json:"content"`
	Cover        *string `json:"cover,omitempty"`
	CategoryID   *int64  `json:"categoryId,omitempty"`
	TagIDs       []int64 `json:"tagIds,omitempty"`
	ShortURL     *string `json:"shortUrl,omitempty"`
	IsPublished  bool    `json:"isPublished"`
	IsTop        bool    `json:"isTop"`
	AllowComment *bool   `json:"allowComment,omitempty"`
	IsOriginal   bool    `json:"isOriginal"`
}

// MomentListItem 手记列表项。
type MomentListItem struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	ShortURL    string    `json:"shortUrl"`
	Summary     string    `json:"summary"`
	Image       []string  `json:"image"`
	Views       int64     `json:"views"`
	ColumnName  string    `json:"columnName"`
	Topics      []string  `json:"topics"`
	Likes       int       `json:"likes"`
	Comments    int       `json:"comments"`
	IsTop       bool      `json:"isTop"`
	IsPublished bool      `json:"isPublished"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// MomentList 手记分页列表。
type MomentList struct {
	Items []MomentListItem `json:"items"`
	Total int64            `json:"total"`
	Page  int              `json:"page"`
	Size  int              `json:"size"`
}

// Moment 手记详情。
type Moment struct {
	ID           int64     `json:"id"`
	Title        string    `json:"title"`
	Summary      string    `json:"summary"`
	Content      string    `json:"content"`
	Image        []string  `json:"image"`
	ColumnID     *int64    `json:"columnId"`
	ColumnName   string    `json:"columnName"`
	ShortURL     string    `json:"shortUrl"`
	IsPublished  bool      `json:"isPublished"`
	IsTop        bool      `json:"isTop"`
	AllowComment bool      `json:"allowComment"`
	IsOriginal   bool      `json:"isOriginal"`
	Topics       []Tag     `json:"topics"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// MomentUpsertReq 创建/更新手记请求。
type MomentUpsertReq struct {
	Title        string   `json:"title"`
	Summary      string   `json:"summary"`
	Content      string   `json:"content"`
	Image        []string `json:"image,omitempty"`
	ColumnID     *int64   `json:"columnId,omitempty"`
	TopicIDs     []int64  `json:"topicIds,omitempty"`
	ShortURL     *string  `json:"shortUrl,omitempty"`
	IsPublished  bool     `json:"isPublished"`
	IsTop        bool     `json:"isTop"`
	AllowComment *bool    `json:"allowComment,omitempty"`
	IsOriginal   bool     `json:"isOriginal"`
}

// Thinking 思考。
type Thinking struct {
	ID           int64     `json:"id"`
	Content      string    `json:"content"`
	AuthorID     int64     `json:"authorId"`
	AllowComment bool      `json:"allowComment"`
	Views        int64     `json:"views"`
	Likes        int       `json:"likes"`
	Comments     int       `json:"comments"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// ThinkingList 思考列表。
type ThinkingList struct {
	Items []Thinking `json:"items"`
	Total int64      `json:"total"`
}

// ThinkingUpsertReq 创建/更新思考请求。
type ThinkingUpsertReq struct {
	Content      string `json:"content"`
	AllowComment *bool  `json:"allowComment,omitempty"`
}

// BatchIDsReq 批量 ID 请求。
type BatchIDsReq struct {
	IDs []int64 `json:"ids"`
}

// BatchPublishedReq 批量发布状态请求。
type BatchPublishedReq struct {
	IDs         []int64 `json:"ids"`
	IsPublished bool    `json:"isPublished"`
}

// BatchTopReq 批量置顶请求。
type BatchTopReq struct {
	IDs   []int64 `json:"ids"`
	IsTop bool    `json:"isTop"`
}

// Comment 管理端评论。
type Comment struct {
	ID        string     `json:"id"`
	AreaID    int64      `json:"areaId"`
	AreaTitle *string    `json:"areaTitle"`
	Content   *string    `json:"content"`
	NickName  *string    `json:"nickName"`
	Location  *string    `json:"location"`
	Status    string     `json:"status"`
	IsViewed  bool       `json:"isViewed"`
	IsTop     bool       `json:"isTop"`
	ParentID  *string    `json:"parentId"`
	CreatedAt time.Time  `json:"createdAt"`
	DeletedAt *time.Time `json:"deletedAt,omitempty"`
}

// CommentList 评论分页列表。
type CommentList struct {
	Items []Comment `json:"items"`
	Total int64     `json:"total"`
	Page  int       `json:"page"`
	Size  int       `json:"size"`
}

// CommentStatusReq 更新评论状态请求。
type CommentStatusReq struct {
	Status string `json:"status"`
}

// CommentReplyReq 回复评论请求。
type CommentReplyReq struct {
	Content string `json:"content"`
}

// CommentViewedReq 标记评论已读请求。
type CommentViewedReq struct {
	IDs      []string `json:"ids"`
	IsViewed *bool    `json:"isViewed,omitempty"`
}

// UploadFile 上传文件。
type UploadFile struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Path      string    `json:"path"`
	PublicURL string    `json:"publicUrl"`
	Type      string    `json:"type"`
	Size      int64     `json:"size"`
	CreatedAt time.Time `json:"createdAt"`
}

// UploadFileList 上传文件分页列表。
type UploadFileList struct {
	Items []UploadFile `json:"items"`
	Total int64        `json:"total"`
	Page  int          `json:"page"`
	Size  int          `json:"size"`
}

// UploadRenameReq 重命名上传文件请求。
type UploadRenameReq struct {
	Name string `json:"name"`
}

// UploadSyncResult 同步索引结果。
type UploadSyncResult struct {
	Scanned           int `json:"scanned"`
	Indexed           int `json:"indexed"`
	Created           int `json:"created"`
	Updated           int `json:"updated"`
	Deleted           int `json:"deleted"`
	SkippedDuplicates int `json:"skippedDuplicates"`
}

// SysConfigTree 系统配置树。
type SysConfigTree struct {
	Groups []SysConfigGroup `json:"groups"`
	Items  []SysConfigItem  `json:"items,omitempty"`
}

// SysConfigGroup 配置分组。
type SysConfigGroup struct {
	Key      string           `json:"key"`
	Path     string           `json:"path"`
	Label    string           `json:"label"`
	Children []SysConfigGroup `json:"children,omitempty"`
	Items    []SysConfigItem  `json:"items,omitempty"`
}

// SysConfigItem 配置项。
type SysConfigItem struct {
	Key         string `json:"key"`
	GroupPath   string `json:"groupPath"`
	Label       string `json:"label"`
	Description string `json:"description"`
	ValueType   string `json:"valueType"`
	IsSensitive bool   `json:"isSensitive"`
	Value       any    `json:"value"`
}

// SysConfigBatchUpdateReq 批量更新配置请求。
type SysConfigBatchUpdateReq struct {
	Items []SysConfigUpdateItem `json:"items"`
}

// SysConfigUpdateItem 单条配置更新。
type SysConfigUpdateItem struct {
	Key   string `json:"key"`
	Value any    `json:"value"`
}

// TaxonomyReq 分类/专栏创建与更新请求。
type TaxonomyReq struct {
	Name     string  `json:"name"`
	ShortURL *string `json:"shortUrl,omitempty"`
}

// TagReq 标签创建与更新请求。
type TagReq struct {
	Name string `json:"name"`
}

// SystemStatus 系统运行状态（仅 CLI 关心的字段）。
type SystemStatus struct {
	App struct {
		Version   string `json:"version"`
		Commit    string `json:"commit"`
		GoVersion string `json:"goVersion"`
		Uptime    string `json:"uptime"`
	} `json:"app"`
	Memory struct {
		Alloc uint64 `json:"alloc"`
		Sys   uint64 `json:"sys"`
	} `json:"memory"`
	Disk struct {
		All  uint64 `json:"all"`
		Used uint64 `json:"used"`
		Free uint64 `json:"free"`
	} `json:"disk"`
	Storage struct {
		Size uint64 `json:"size"`
	} `json:"storage"`
	Database struct {
		Status  string `json:"status"`
		Version string `json:"version"`
	} `json:"database"`
	Redis struct {
		Status     string `json:"status"`
		Version    string `json:"version"`
		UsedMemory string `json:"usedMemory"`
	} `json:"redis"`
	Components []struct {
		Name    string `json:"name"`
		Status  string `json:"status"`
		Healthy bool   `json:"healthy"`
	} `json:"components"`
	Update struct {
		HasUpdate      bool   `json:"hasUpdate"`
		LatestVersion  string `json:"latestVersion"`
		CurrentVersion string `json:"currentVersion"`
	} `json:"update"`
	HealthMode string `json:"healthMode"`
}
