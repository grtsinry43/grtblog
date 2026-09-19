<script lang="ts">
	import { Gamepad2 } from 'lucide-svelte';
	import type { OwnerStatusGame } from '$lib/features/owner-status/types';
	import { gameCatalogStore } from '$lib/features/owner-status/game-catalog.svelte';

	let { process, game = null }: { process: string; game?: OwnerStatusGame | null } = $props();

	const DEFAULT_ACCENT = '#3f8f6b';

	type CardData = {
		name: string;
		slogan: string;
		desc: string;
		accent: string;
		headerImage?: string;
		url?: string;
	};

	// 上报端规则里维护的游戏信息优先；没配才懒查服务端兜底源（Steam）。
	// 卡片标题与搜索关键词优先用规则里配置的游戏标准名，回退到 process 显示名。
	const gameName = $derived((game?.name ?? '').trim() || process);

	const hasReporterGame = $derived(
		Boolean(
			game && (game.name || game.cover || game.slogan || game.desc || game.accent || game.url)
		)
	);

	$effect(() => {
		if (!hasReporterGame && gameName.trim()) {
			gameCatalogStore.ensure(gameName);
		}
	});

	const lookup = $derived(gameCatalogStore.get(gameName));

	const card = $derived.by<CardData | 'loading'>(() => {
		if (hasReporterGame && game) {
			return {
				name: gameName,
				slogan: game.slogan ?? '',
				desc: game.desc ?? '',
				accent: game.accent || DEFAULT_ACCENT,
				headerImage: game.cover || undefined,
				url: game.url || undefined
			};
		}
		if (lookup.status === 'done' && lookup.data.found) {
			// Steam 兜底卡片不伪造「手写 slogan」，只展示官方信息。
			return {
				name: lookup.data.name || gameName,
				slogan: '',
				desc: lookup.data.shortDescription ?? '',
				accent: DEFAULT_ACCENT,
				headerImage: lookup.data.headerImage,
				url: lookup.data.storeUrl
			};
		}
		if (lookup.status === 'idle' || lookup.status === 'loading') {
			return 'loading';
		}
		// 兜底：查不到也至少亮个相，保持「正在游戏中」的现场感。
		return {
			name: gameName,
			slogan: '',
			desc: '',
			accent: DEFAULT_ACCENT
		};
	});

	let imageFailed = $state(false);
	$effect(() => {
		void (card !== 'loading' ? card.headerImage : undefined);
		imageFailed = false;
	});
</script>

{#if card === 'loading'}
	<div class="h-24 animate-pulse rounded-default bg-ink-100 dark:bg-ink-800"></div>
{:else}
	<svelte:element
		this={card.url ? 'a' : 'div'}
		href={card.url}
		target={card.url ? '_blank' : undefined}
		rel={card.url ? 'noreferrer noopener' : undefined}
		class="group/game block overflow-hidden rounded-default border border-ink-100 transition-colors dark:border-ink-800 {card.url
			? 'hover:border-jade-300 dark:hover:border-jade-800'
			: ''}"
	>
		<div
			class="relative h-24 overflow-hidden"
			style:background="linear-gradient(135deg, {card.accent}33, {card.accent}0d 45%, transparent)"
		>
			{#if card.headerImage && !imageFailed}
				<img
					src={card.headerImage}
					alt={card.name}
					loading="lazy"
					class="h-full w-full object-cover transition-transform duration-500 group-hover/game:scale-105"
					onerror={() => (imageFailed = true)}
				/>
			{:else}
				<div class="absolute inset-0 flex items-center justify-center">
					<Gamepad2 class="h-8 w-8 opacity-30" style="color: {card.accent}" />
				</div>
			{/if}
			<div
				class="absolute inset-0 bg-gradient-to-t from-black/70 via-black/10 to-transparent"
			></div>
			<span
				class="absolute right-2 top-2 inline-flex items-center gap-1 rounded-full bg-black/45 px-1.5 py-0.5 text-[9px] font-medium tracking-wide text-white backdrop-blur-sm"
			>
				<span class="relative flex h-1.5 w-1.5">
					<span
						class="absolute inline-flex h-full w-full animate-ping rounded-full opacity-75"
						style:background-color={card.accent}
					></span>
					<span
						class="relative inline-flex h-1.5 w-1.5 rounded-full"
						style:background-color={card.accent}
					></span>
				</span>
				游玩中
			</span>
			<div class="absolute bottom-1.5 left-2.5 right-2.5">
				<p class="game-name truncate font-serif text-sm font-semibold text-white">
					{card.name}
				</p>
			</div>
		</div>
		{#if card.slogan || card.desc}
			<div class="bg-white/60 px-2.5 py-2 dark:bg-ink-900/40">
				{#if card.slogan}
					<p class="font-serif text-[12px] italic leading-4" style:color={card.accent}>
						「{card.slogan}」
					</p>
				{/if}
				{#if card.desc}
					<p
						class="line-clamp-2 text-[11px] leading-4 text-ink-500 dark:text-ink-400 {card.slogan
							? 'mt-1'
							: ''}"
					>
						{card.desc}
					</p>
				{/if}
			</div>
		{/if}
	</svelte:element>
{/if}

<style lang="postcss">
	@reference "$routes/layout.css";

	.game-name {
		text-shadow: 0 1px 4px rgb(0 0 0 / 0.55);
	}
</style>
