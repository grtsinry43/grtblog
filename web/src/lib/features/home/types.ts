export type HomeSubscriptionPreference = 'posts' | 'moments' | 'thinkings';

export type PublicEmailEventName = 'article.published' | 'moment.published' | 'thinking.created';

export type SubscribeEmailPayload = {
	email: string;
	eventNames: PublicEmailEventName[];
};

export type EmailSubscriptionItem = {
	id: number;
	email: string;
	eventName: string;
	createdAt: string;
	updatedAt: string;
};

export type SubscribeEmailResponse = {
	items: EmailSubscriptionItem[];
};

export type HomeHeroTemplateNodeType = 'h1' | 'span' | 'code' | 'br';

export type HomeHeroTemplateNode = {
	type: HomeHeroTemplateNodeType;
	text?: string;
	variant?: string;
	className?: string;
};

export type HomeHeroSocialLink = {
	icon: string;
	name: string;
	href: string;
};

export type HomeHeroAlignMode = 'default' | 'center';

export type HomeHeroThemeConfig = {
	avatarUrl?: string;
	description?: string;
	titleTemplate?: HomeHeroTemplateNode[];
	mottoLines?: string[];
	mottoLinesAlign?: HomeHeroAlignMode;
	socials?: HomeHeroSocialLink[];
	socialsAlign?: HomeHeroAlignMode;
};

export type HomeActivityPulseThemeConfig = {
	title?: string;
	subtitle?: string;
	statusLabel?: string;
	rangeLabelStart?: string;
	rangeLabelEnd?: string;
	legend?: {
		posts?: string;
		moments?: string;
	};
	rangeDays?: number | 'all';
};

export type HomeInspirationIconName =
	| 'quote'
	| 'code2'
	| 'gamepad2'
	| 'coffee'
	| 'library'
	| 'zap'
	| 'sparkles'
	| 'github';

export type HomeInspirationNowItem = {
	id: string;
	label: string;
	value: string;
	icon?: HomeInspirationIconName;
};

export type HomeInspirationStatSourceType =
	| 'static'
	| 'words_total'
	| 'github_recent_push_commits'
	| 'github_followers'
	| 'github_public_repos';

export type HomeInspirationStatSource = {
	type: HomeInspirationStatSourceType;
};

export type HomeInspirationStatItem = {
	id: string;
	label: string;
	value?: string;
	icon?: HomeInspirationIconName;
	colorClass?: string;
	source?: HomeInspirationStatSource;
};

export type HomeInspirationThemeConfig = {
	sectionTitle?: string;
	quote?: {
		text?: string;
		author?: string;
	};
	now?: {
		title?: string;
		items?: HomeInspirationNowItem[];
	};
	energy?: {
		label?: string;
		icon?: HomeInspirationIconName;
	};
	techStack?: {
		title?: string;
		items?: string[];
		icons?: HomeInspirationIconName[];
	};
	stats?: HomeInspirationStatItem[];
	github?: {
		username?: string;
	};
};

export type HomeThemeConfig = {
	hero?: HomeHeroThemeConfig;
	activityPulse?: HomeActivityPulseThemeConfig;
	inspiration?: HomeInspirationThemeConfig;
};

export type HomeActivityPulsePoint = {
	date: string;
	posts: number;
	moments: number;
};

export type HomeActivityPulseData = {
	days: number;
	startDate: string;
	endDate: string;
	totalPosts: number;
	totalMoments: number;
	statusLabel: string;
	points: HomeActivityPulsePoint[];
};

export type HomeWordCountStats = {
	total: number;
	articles: number;
	moments: number;
	pages: number;
	thinkings: number;
};

export type HomeGitHubStats = {
	username: string;
	profileUrl: string;
	avatarUrl: string;
	followers: number;
	publicRepos: number;
	recentPushCommits: number;
	fetchedAt: string;
};

export type HomeInspirationStatsData = {
	words: HomeWordCountStats;
	github?: HomeGitHubStats;
	githubError?: string;
};
