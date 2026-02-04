package contract

type AdminEventGroupResp struct {
	Category string   `json:"category"`
	Events   []string `json:"events"`
}

type AdminEventListResp struct {
	Groups []AdminEventGroupResp `json:"groups"`
}

type AdminEventFieldResp struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Required    bool   `json:"required"`
	Description string `json:"description"`
}

type AdminEventDescriptorResp struct {
	Name        string                `json:"name"`
	Title       string                `json:"title"`
	Category    string                `json:"category"`
	Description string                `json:"description"`
	PublicEmail bool                  `json:"publicEmail"`
	Channels    []string              `json:"channels"`
	Fields      []AdminEventFieldResp `json:"fields"`
}

type AdminEventCatalogResp struct {
	Items []AdminEventDescriptorResp `json:"items"`
}
