export type FooterThemeLink = {
	name: string;
	href: string;
};

export type FooterThemeSection = {
	title: string;
	links: FooterThemeLink[];
};

export type FooterThemeConfig = {
	sections: FooterThemeSection[];
	brandName: string;
	brandTagline: string;
	copyrightStartYear: number;
	copyrightOwner: string;
	beianText: string;
	beianUrl: string;
	beianGongAnText: string;
	designedWithText: string;
	presenceConnectedText: string;
	presenceLoadingText: string;
};
