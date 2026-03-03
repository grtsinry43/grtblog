package like

type TrackLikeCmd struct {
	ContentType string
	ContentID   int64
	VisitorID   string
}

type ImportLikeBatchCmd struct {
	ContentType string
	ContentID   int64
	VisitorIDs  []string
}
