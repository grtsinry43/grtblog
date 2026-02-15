<script lang="ts">
	import { SlideIn, StaggerList } from '$lib/ui/animation';
	import { Quote, Code2, Gamepad2, Coffee, Library, Zap, Sparkles, Github } from 'lucide-svelte';
	import type {
		HomeGitHubStats,
		HomeInspirationIconName,
		HomeInspirationStatItem,
		HomeInspirationStatsData,
		HomeInspirationThemeConfig,
		HomeWordCountStats
	} from './types';

	let { config, stats }: { config?: HomeInspirationThemeConfig; stats?: HomeInspirationStatsData } =
		$props();

	const iconMap: Record<HomeInspirationIconName, typeof Quote> = {
		quote: Quote,
		code2: Code2,
		gamepad2: Gamepad2,
		coffee: Coffee,
		library: Library,
		zap: Zap,
		sparkles: Sparkles,
		github: Github
	};

	const defaultNowItems = [
		{ id: 'coding', label: 'Coding', value: 'grtblog-v2', icon: 'code2' as const },
		{
			id: 'reading',
			label: 'Reading',
			value: 'The Design of Everyday Things',
			icon: 'library' as const
		},
		{ id: 'learning', label: 'Learning', value: 'Svelte 5 & Runes', icon: 'zap' as const }
	];
	const defaultStats: HomeInspirationStatItem[] = [
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
	];

	const sectionTitle = $derived(config?.sectionTitle || '灵感与实验场');
	const quoteText = $derived(
		config?.quote?.text || '“The best way to predict the future is to invent it.”'
	);
	const quoteAuthor = $derived(config?.quote?.author || 'Alan Kay');
	const nowTitle = $derived(config?.now?.title || 'Now / 正在');
	const nowItems = $derived.by(
		() =>
			(config?.now?.items && config.now.items.length > 0
				? config.now.items
				: defaultNowItems) as Array<{
				id: string;
				label: string;
				value: string;
				icon?: HomeInspirationIconName;
			}>
	);
	const energyLabel = $derived(config?.energy?.label || 'High Energy');
	const energyIcon = $derived(resolveIcon(config?.energy?.icon, Sparkles));
	const techTitle = $derived(config?.techStack?.title || 'Tech Stack');
	const techItems = $derived(
		config?.techStack?.items && config.techStack.items.length > 0
			? config.techStack.items
			: ['Java', 'TypeScript', 'Svelte', 'Rust']
	);
	const techIcons = $derived(
		config?.techStack?.icons && config.techStack.icons.length > 0
			? config.techStack.icons.map((item) => resolveIcon(item, Code2))
			: [Code2, Gamepad2]
	);
	const statItems = $derived(
		config?.stats && config.stats.length > 0 ? config.stats : defaultStats
	);
	const wordStats = $derived(stats?.words);
	const githubStats = $derived(stats?.github);

	function resolveIcon(
		name: HomeInspirationIconName | undefined,
		fallback: typeof Quote
	): typeof Quote {
		if (!name) {
			return fallback;
		}
		return iconMap[name] ?? fallback;
	}

	function formatCompact(value: number): string {
		if (!Number.isFinite(value)) {
			return '-';
		}
		return new Intl.NumberFormat('en', {
			notation: 'compact',
			maximumFractionDigits: 1
		}).format(value);
	}

	function resolveDynamicStatValue(
		item: HomeInspirationStatItem,
		words: HomeWordCountStats | undefined,
		github: HomeGitHubStats | undefined
	): string {
		const source = item.source?.type;
		switch (source) {
			case 'words_total':
				return words ? formatCompact(words.total) : '-';
			case 'github_recent_push_commits':
				return github ? formatCompact(github.recentPushCommits) : '-';
			case 'github_followers':
				return github ? formatCompact(github.followers) : '-';
			case 'github_public_repos':
				return github ? formatCompact(github.publicRepos) : '-';
			default:
				return item.value || '-';
		}
	}

	function isStatMissing(
		item: HomeInspirationStatItem,
		words: HomeWordCountStats | undefined,
		github: HomeGitHubStats | undefined
	): boolean {
		const source = item.source?.type;
		switch (source) {
			case 'words_total':
				return !words;
			case 'github_recent_push_commits':
			case 'github_followers':
			case 'github_public_repos':
				return !github;
			default:
				return !item.value;
		}
	}
</script>

<section class="mt-20 md:mt-32">
	<SlideIn direction="up">
		<div class="flex items-center gap-3 mb-10 border-b border-ink-100 dark:border-ink-800 pb-4">
			<span class="h-px w-8 bg-jade-500/40"></span>
			<h2 class="text-xl font-serif font-medium text-ink-900 dark:text-ink-100">{sectionTitle}</h2>
		</div>
	</SlideIn>

	<StaggerList
		staggerDelay={80}
		y={20}
		class="grid grid-cols-1 md:grid-cols-4 lg:grid-cols-6 gap-4 auto-rows-[120px]"
	>
		<div
			class="col-span-1 md:col-span-3 lg:col-span-3 row-span-2 bento-card p-8 flex flex-col justify-center relative overflow-hidden group"
		>
			<Quote
				class="absolute -top-4 -left-4 w-24 h-24 text-ink-100 dark:text-ink-800/50 -rotate-12 transition-transform group-hover:rotate-0 duration-700"
			/>
			<div class="relative z-10">
				<p
					class="text-2xl md:text-3xl font-serif italic text-ink-800 dark:text-ink-100 leading-relaxed"
				>
					{quoteText}
				</p>
				<p class="mt-4 font-mono text-sm text-ink-400">— {quoteAuthor}</p>
			</div>
			<div class="absolute bottom-0 right-0 w-32 h-32 bg-jade-500/5 blur-[80px] rounded-full"></div>
		</div>

		<div
			class="col-span-1 md:col-span-1 lg:col-span-2 row-span-2 bento-card p-6 flex flex-col gap-6"
		>
			<div
				class="flex items-center gap-2 text-xs font-mono uppercase tracking-wider text-jade-600 dark:text-jade-400"
			>
				<span class="relative flex h-2 w-2">
					<span
						class="animate-ping absolute inline-flex h-full w-full rounded-full bg-jade-400 opacity-75"
					></span>
					<span class="relative inline-flex rounded-full h-2 w-2 bg-jade-500"></span>
				</span>
				{nowTitle}
			</div>

			<div class="space-y-4">
				{#each nowItems as item (item.id)}
					<div class="flex items-start gap-3">
						<svelte:component
							this={resolveIcon(item.icon, Code2)}
							size={16}
							class="mt-0.5 text-ink-400"
						/>
						<div>
							<div class="text-[10px] font-mono text-ink-400 uppercase">{item.label}</div>
							<div class="text-sm font-medium">{item.value}</div>
						</div>
					</div>
				{/each}
			</div>
		</div>

		<div
			class="col-span-1 md:col-span-1 lg:col-span-1 row-span-1 bento-card flex items-center justify-center group"
		>
			<div class="text-center">
				<svelte:component
					this={energyIcon}
					size={24}
					class="mx-auto mb-2 text-amber-400 transition-transform group-hover:scale-125 duration-500"
				/>
				<div class="text-xs font-mono">{energyLabel}</div>
			</div>
		</div>

		{#each statItems as stat (stat.id)}
			<div
				class="col-span-1 md:col-span-1 lg:col-span-1 row-span-1 bento-card p-4 flex flex-col justify-between hover:border-jade-200 dark:hover:border-jade-900/50 transition-colors"
			>
				<svelte:component this={resolveIcon(stat.icon, Library)} size={18} class="text-ink-400" />
				<div>
					<div
						class="text-2xl font-serif {isStatMissing(stat, wordStats, githubStats)
							? 'text-cinnabar-500 dark:text-cinnabar-400'
							: stat.colorClass || 'text-jade-500'}"
					>
						{resolveDynamicStatValue(stat, wordStats, githubStats)}
					</div>
					<div class="text-[10px] font-mono uppercase tracking-tighter text-ink-400">
						{stat.label}
					</div>
				</div>
			</div>
		{/each}

		<div
			class="col-span-1 md:col-span-2 lg:col-span-3 row-span-1 bento-card px-6 flex items-center justify-between overflow-hidden relative"
		>
			<div class="flex flex-col">
				<div class="text-[10px] font-mono text-ink-400 uppercase mb-1">{techTitle}</div>
				<div class="flex gap-4">
					{#each techItems as item (`stack-${item}`)}
						<span class="text-sm font-medium hover:text-jade-500 cursor-default transition-colors">
							{item}
						</span>
					{/each}
				</div>
			</div>
			<div class="flex gap-2">
				{#each techIcons as Icon, idx (`icon-${idx}`)}
					<div
						class="w-8 h-8 rounded-full bg-ink-100 dark:bg-ink-800 flex items-center justify-center transition-colors hover:bg-jade-100 dark:hover:bg-jade-900/30"
					>
						<svelte:component this={Icon} size={14} />
					</div>
				{/each}
			</div>
		</div>
	</StaggerList>
</section>

<style lang="postcss">
	@reference "$routes/layout.css";

	.bento-card {
		@apply relative z-0 rounded-default border border-ink-200/80 bg-ink-50 transition-all duration-500 hover:shadow-glass dark:border-ink-800 dark:bg-ink-900/50 dark:hover:shadow-glass-dark;
	}

	.bento-card::after {
		content: '';
		@apply pointer-events-none absolute inset-0 -z-10 rounded-default opacity-20;
		background-image: var(--texture-noise);
		mix-blend-mode: soft-light;
	}

	:root[class~='dark'] .bento-card::after {
		@apply opacity-10;
		filter: invert(1);
	}

	.bento-card:hover {
		@apply -translate-y-1 border-ink-200 dark:border-ink-700;
	}
</style>
