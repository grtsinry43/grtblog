package album

// AlbumListOptions 公开列表查询参数。
type AlbumListOptions struct {
	Page     int
	PageSize int
	Search   *string
}

// AlbumListOptionsInternal 管理后台列表查询参数。
type AlbumListOptionsInternal struct {
	Page      int
	PageSize  int
	Published *bool
	Search    *string
}
