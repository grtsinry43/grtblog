export interface SiteSearchItemResp {
	id: number;
	title: string;
	summary: string;
	snippet: string;
	shortUrl?: string; // Optional because the pointer in Go can be nil
	path: string; // The constructed frontend path
	score: number;
	createdAt: string; // ISO date string
}

export interface SiteSearchResp {
	query: string;
	keywords: string[];
	cached: boolean;
	articles: SiteSearchItemResp[];
	moments: SiteSearchItemResp[];
	pages: SiteSearchItemResp[];
	thinkings: SiteSearchItemResp[];
}
