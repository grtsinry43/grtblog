package contract

type AdminUserListResp struct {
	Items []UserResp `json:"items"`
	Total int64      `json:"total"`
	Page  int        `json:"page"`
	Size  int        `json:"size"`
}

type UpdateAdminUserReq struct {
	Nickname string `json:"nickname"`
	Email    string `json:"email"`
	IsActive bool   `json:"isActive"`
	IsAdmin  bool   `json:"isAdmin"`
}
