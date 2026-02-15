import type { WebsiteInfoMap } from '$lib/features/website-info/types';
import type {
	HomeActivityPulseThemeConfig,
	HomeHeroSocialLink,
	HomeHeroTemplateNode,
	HomeInspirationIconName,
	HomeInspirationNowItem,
	HomeInspirationStatItem,
	HomeInspirationStatSource,
	HomeInspirationThemeConfig,
	HomeThemeConfig
} from './types';

const allowedHeroTemplateTypes = new Set(['h1', 'span', 'code', 'br']);
const allowedInspirationIconNames = new Set([
	'quote',
	'code2',
	'gamepad2',
	'coffee',
	'library',
	'zap',
	'sparkles',
	'github'
]);
const allowedInspirationSourceTypes = new Set([
	'static',
	'words_total',
	'github_recent_push_commits',
	'github_followers',
	'github_public_repos'
]);

const defaultHeroTemplate: HomeHeroTemplateNode[] = [
	{ type: 'h1', text: 'Hi! 👋', variant: 'hero_h1_highlight' },
	{ type: 'br' },
	{ type: 'h1', text: "I'm grtsinry43", variant: 'hero_h1_primary' }
];

const defaultHeroSocials: HomeHeroSocialLink[] = [
	{ icon: 'github', name: 'GitHub', href: 'https://github.com/grtinry43' },
	{ icon: 'mail', name: 'Email', href: 'mailto:grtsinry43@outlook.com' },
	{ icon: 'rss', name: 'RSS', href: '/feed' }
];

const defaultThemeConfig: HomeThemeConfig = {
	hero: {
		avatarUrl: 'https://dogeoss.grtsinry43.com/img/author.jpeg',
		description: 'Java & JavaScript full-stack developer committed to crafting excellent software.',
		titleTemplate: defaultHeroTemplate,
		mottoLines: [
			'热衷于在逻辑与感性的缝隙中构建数字花园。',
			'也许，代码是现代的诗歌，而文字是思想的快照。'
		],
		socials: defaultHeroSocials
	},
	activityPulse: {
		title: '创作律动',
		subtitle: '近一年的数字足迹：逻辑的向上生长，感性的向下扎根。',
		rangeLabelStart: '365 Days Ago',
		rangeLabelEnd: 'Today',
		legend: {
			posts: 'Article',
			moments: 'Moment'
		},
		rangeDays: 365
	},
	inspiration: {
		sectionTitle: '灵感与实验场',
		quote: {
			text: '“The best way to predict the future is to invent it.”',
			author: 'Alan Kay'
		},
		now: {
			title: 'Now / 正在',
			items: [
				{ id: 'coding', label: 'Coding', value: 'grtblog-v2', icon: 'code2' },
				{
					id: 'reading',
					label: 'Reading',
					value: 'The Design of Everyday Things',
					icon: 'library'
				},
				{ id: 'learning', label: 'Learning', value: 'Svelte 5 & Runes', icon: 'zap' }
			]
		},
		energy: {
			label: 'High Energy',
			icon: 'sparkles'
		},
		stats: [
			{
				id: 'words',
				label: 'Words',
				icon: 'library',
				colorClass: 'text-jade-500',
				source: { type: 'words_total' }
			},
			{
				id: 'commits',
				label: 'Commits',
				icon: 'github',
				colorClass: 'text-ink-900 dark:text-ink-100',
				source: { type: 'github_recent_push_commits' }
			},
			{
				id: 'coffee',
				label: 'Coffee',
				value: '∞',
				icon: 'coffee',
				colorClass: 'text-amber-500',
				source: { type: 'static' }
			}
		],
		techStack: {
			title: 'Tech Stack',
			items: ['Java', 'TypeScript', 'Svelte', 'Rust'],
			icons: ['code2', 'gamepad2']
		},
		github: {
			username: 'grtsinry43'
		}
	}
};

const isRecord = (value: unknown): value is Record<string, unknown> =>
	typeof value === 'object' && value !== null;

const toStringValue = (value: unknown): string | undefined => {
	if (typeof value !== 'string') {
		return undefined;
	}
	const trimmed = value.trim();
	return trimmed.length > 0 ? trimmed : undefined;
};

const toPositiveInt = (value: unknown): number | undefined => {
	if (typeof value !== 'number' || !Number.isFinite(value)) {
		return undefined;
	}
	const parsed = Math.floor(value);
	return parsed > 0 ? parsed : undefined;
};

const toRangeDaysValue = (value: unknown): number | 'all' | undefined => {
	const str = toStringValue(value);
	if (str && str.toLowerCase() === 'all') {
		return 'all';
	}
	return toPositiveInt(value);
};

const parseHeroTemplate = (value: unknown): HomeHeroTemplateNode[] | undefined => {
	if (!Array.isArray(value)) {
		return undefined;
	}
	const nodes: HomeHeroTemplateNode[] = [];
	for (const item of value) {
		if (!isRecord(item)) {
			continue;
		}
		const typeRaw = toStringValue(item.type);
		if (!typeRaw || !allowedHeroTemplateTypes.has(typeRaw)) {
			continue;
		}
		nodes.push({
			type: typeRaw as HomeHeroTemplateNode['type'],
			text: toStringValue(item.text),
			variant: toStringValue(item.variant),
			className: toStringValue(item.className ?? item.class)
		});
	}
	return nodes.length > 0 ? nodes : undefined;
};

const parseSocials = (value: unknown): HomeHeroSocialLink[] | undefined => {
	if (!Array.isArray(value)) {
		return undefined;
	}
	const items: HomeHeroSocialLink[] = [];
	for (const item of value) {
		if (!isRecord(item)) {
			continue;
		}
		const icon = toStringValue(item.icon);
		const name = toStringValue(item.name);
		const href = toStringValue(item.href);
		if (!icon || !href) {
			continue;
		}
		items.push({
			icon,
			name: name ?? '',
			href
		});
	}
	return items.length > 0 ? items : undefined;
};

const parseStringList = (value: unknown): string[] | undefined => {
	if (!Array.isArray(value)) {
		return undefined;
	}
	const items = value
		.map((item) => toStringValue(item))
		.filter((item): item is string => Boolean(item));
	return items.length > 0 ? items : undefined;
};

const parseInspirationIcon = (value: unknown): HomeInspirationIconName | undefined => {
	const icon = toStringValue(value);
	if (!icon || !allowedInspirationIconNames.has(icon)) {
		return undefined;
	}
	return icon as HomeInspirationIconName;
};

const parseInspirationNowItems = (value: unknown): HomeInspirationNowItem[] | undefined => {
	if (!Array.isArray(value)) {
		return undefined;
	}
	const items: HomeInspirationNowItem[] = [];
	for (const item of value) {
		if (!isRecord(item)) {
			continue;
		}
		const id = toStringValue(item.id) ?? toStringValue(item.key);
		const label = toStringValue(item.label);
		const val = toStringValue(item.value);
		if (!id || !label || !val) {
			continue;
		}
		items.push({
			id,
			label,
			value: val,
			icon: parseInspirationIcon(item.icon)
		});
	}
	return items.length > 0 ? items : undefined;
};

const parseInspirationStatSource = (value: unknown): HomeInspirationStatSource | undefined => {
	const sourceType = isRecord(value) ? toStringValue(value.type) : toStringValue(value);
	if (!sourceType || !allowedInspirationSourceTypes.has(sourceType)) {
		return undefined;
	}
	return { type: sourceType as HomeInspirationStatSource['type'] };
};

const parseInspirationStats = (value: unknown): HomeInspirationStatItem[] | undefined => {
	if (!Array.isArray(value)) {
		return undefined;
	}
	const items: HomeInspirationStatItem[] = [];
	for (const item of value) {
		if (!isRecord(item)) {
			continue;
		}
		const id = toStringValue(item.id) ?? toStringValue(item.key);
		const label = toStringValue(item.label);
		if (!id || !label) {
			continue;
		}
		items.push({
			id,
			label,
			value: toStringValue(item.value),
			icon: parseInspirationIcon(item.icon),
			colorClass: toStringValue(item.colorClass ?? item.color),
			source: parseInspirationStatSource(item.source ?? item.sourceType)
		});
	}
	return items.length > 0 ? items : undefined;
};

const parseActivityPulse = (value: unknown): HomeActivityPulseThemeConfig | undefined => {
	if (!isRecord(value)) {
		return undefined;
	}
	const legend = isRecord(value.legend)
		? {
				posts: toStringValue(value.legend.posts),
				moments: toStringValue(value.legend.moments)
			}
		: undefined;
	return {
		title: toStringValue(value.title),
		subtitle: toStringValue(value.subtitle),
		statusLabel: toStringValue(value.statusLabel),
		rangeLabelStart: toStringValue(value.rangeLabelStart),
		rangeLabelEnd: toStringValue(value.rangeLabelEnd),
		legend,
		rangeDays: toRangeDaysValue(value.rangeDays)
	};
};

const parseInspiration = (value: unknown): HomeInspirationThemeConfig | undefined => {
	if (!isRecord(value)) {
		return undefined;
	}
	const quote = isRecord(value.quote)
		? {
				text: toStringValue(value.quote.text),
				author: toStringValue(value.quote.author)
			}
		: undefined;
	const now = isRecord(value.now)
		? {
				title: toStringValue(value.now.title),
				items: parseInspirationNowItems(value.now.items)
			}
		: undefined;
	const energy = isRecord(value.energy)
		? {
				label: toStringValue(value.energy.label),
				icon: parseInspirationIcon(value.energy.icon)
			}
		: undefined;
	const techStack = isRecord(value.techStack)
		? {
				title: toStringValue(value.techStack.title),
				items: parseStringList(value.techStack.items),
				icons: Array.isArray(value.techStack.icons)
					? value.techStack.icons
							.map((item) => parseInspirationIcon(item))
							.filter((item): item is HomeInspirationIconName => Boolean(item))
					: undefined
			}
		: undefined;
	const github = isRecord(value.github)
		? {
				username: toStringValue(value.github.username)
			}
		: undefined;

	return {
		sectionTitle: toStringValue(value.sectionTitle),
		quote,
		now,
		energy,
		techStack,
		stats: parseInspirationStats(value.stats),
		github
	};
};

export const resolveHomeThemeConfig = (
	websiteInfo: WebsiteInfoMap | null | undefined
): HomeThemeConfig => {
	const themeRaw = websiteInfo?.theme_extend_info;
	if (!isRecord(themeRaw)) {
		return defaultThemeConfig;
	}
	const homeRoot = isRecord(themeRaw.home) ? themeRaw.home : themeRaw;
	const heroRaw = isRecord(homeRoot.hero) ? homeRoot.hero : {};
	const activityRaw = isRecord(homeRoot.activityPulse) ? homeRoot.activityPulse : {};
	const inspirationRaw = isRecord(homeRoot.inspiration) ? homeRoot.inspiration : {};
	const parsedInspiration = parseInspiration(inspirationRaw);

	const hero = {
		avatarUrl: toStringValue(heroRaw.avatarUrl) ?? defaultThemeConfig.hero?.avatarUrl,
		description: toStringValue(heroRaw.description) ?? defaultThemeConfig.hero?.description,
		titleTemplate:
			parseHeroTemplate(isRecord(heroRaw.title) ? heroRaw.title.template : heroRaw.titleTemplate) ??
			defaultThemeConfig.hero?.titleTemplate,
		mottoLines: parseStringList(heroRaw.mottoLines) ?? defaultThemeConfig.hero?.mottoLines,
		socials: parseSocials(heroRaw.socials) ?? defaultThemeConfig.hero?.socials
	};

	const activity = {
		...defaultThemeConfig.activityPulse,
		...parseActivityPulse(activityRaw)
	};
	const inspiration = {
		...defaultThemeConfig.inspiration,
		...parsedInspiration,
		now: {
			...defaultThemeConfig.inspiration?.now,
			...(parsedInspiration?.now ?? {})
		},
		quote: {
			...defaultThemeConfig.inspiration?.quote,
			...(parsedInspiration?.quote ?? {})
		},
		energy: {
			...defaultThemeConfig.inspiration?.energy,
			...(parsedInspiration?.energy ?? {})
		},
		techStack: {
			...defaultThemeConfig.inspiration?.techStack,
			...(parsedInspiration?.techStack ?? {})
		},
		github: {
			...defaultThemeConfig.inspiration?.github,
			...(parsedInspiration?.github ?? {})
		}
	};
	if (
		!inspiration.now?.items ||
		!Array.isArray(inspiration.now.items) ||
		inspiration.now.items.length === 0
	) {
		inspiration.now = {
			...inspiration.now,
			items: defaultThemeConfig.inspiration?.now?.items
		};
	}
	if (!inspiration.stats || !Array.isArray(inspiration.stats) || inspiration.stats.length === 0) {
		inspiration.stats = defaultThemeConfig.inspiration?.stats;
	}
	if (
		!inspiration.techStack?.items ||
		!Array.isArray(inspiration.techStack.items) ||
		inspiration.techStack.items.length === 0
	) {
		inspiration.techStack = {
			...inspiration.techStack,
			items: defaultThemeConfig.inspiration?.techStack?.items
		};
	}
	if (
		!inspiration.techStack?.icons ||
		!Array.isArray(inspiration.techStack.icons) ||
		inspiration.techStack.icons.length === 0
	) {
		inspiration.techStack = {
			...inspiration.techStack,
			icons: defaultThemeConfig.inspiration?.techStack?.icons
		};
	}

	return {
		hero,
		activityPulse: activity,
		inspiration
	};
};
