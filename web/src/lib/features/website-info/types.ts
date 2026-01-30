export type WebsiteInfoThemeExtendInfo = Record<string, string>;

export interface WebsiteInfoMap {
	api_url?: string;
	description?: string;
	keywords?: string;
	favicon?: string;
	og_description?: string;
	og_image?: string;
	og_site_name?: string;
	og_title?: string;
	og_type?: string;
	og_url?: string;
	public_url?: string;
	theme_extend_info?: WebsiteInfoThemeExtendInfo;
	website_name?: string;
}
