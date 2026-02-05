package contract

type AdminNotificationResp struct {
	ID        int64   `json:"id"`
	Type      string  `json:"type"`
	Title     string  `json:"title"`
	Content   string  `json:"content"`
	Payload   any     `json:"payload,omitempty" swaggertype:"object"`
	IsRead    bool    `json:"is_read"`
	ReadAt    *string `json:"read_at,omitempty"`
	CreatedAt string  `json:"created_at"`
}

type AdminNotificationListResp struct {
	Items []AdminNotificationResp `json:"items"`
	Total int64                   `json:"total"`
	Page  int                     `json:"page"`
	Size  int                     `json:"size"`
}
