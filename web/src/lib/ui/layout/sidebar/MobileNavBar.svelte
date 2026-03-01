<script lang="ts">
	import { resolve } from '$app/paths';
	import type { NavMenuItem } from '$lib/features/navigation/types';
	import { buildMomentPath, buildPostPath } from '$lib/shared/utils/content-path';
	import DynamicLucideIcon from '$lib/ui/icons/DynamicLucideIcon.svelte';
	import DetailTocNavList from '$lib/ui/detail/DetailTocNavList.svelte';
	import ThemeIcon from './ThemeIcon.svelte';
	import { X, ChevronDown, List, Calendar, NotebookPen, FileText, Home } from 'lucide-svelte';
	import { page } from '$app/state';
	import { browser } from '$app/environment';
	import { tick } from 'svelte';
	import { fly, fade } from 'svelte/transition';
	import { cubicOut } from 'svelte/easing';
	import { userStore } from '$lib/shared/stores/userStore';
	import { windowStore } from '$lib/shared/stores/windowStore.svelte';
	import { authModalStore } from '$lib/shared/stores/authModalStore';
	import { ownerStatusStore } from '$lib/features/owner-status/store.svelte';
	import { User } from 'lucide-svelte';
	import { websiteInfoCtx } from '$lib/features/website-info/context';
	import { detailPanelCtx } from '$lib/shared/detail-panel/context';

	let { menuTree = [] } = $props<{ menuTree: NavMenuItem[] }>();

	let isMobileMenuOpen = $state(false);
	let isTocOpen = $state(false);
	let expandedMobileItems = $state<string[]>([]);
	let scrollY = $state(0);
	let isMenuAnimating = $state(false);

	// Interpolation progress: 0 (top/capsule) -> 1 (scrolled/full)
	let navProgress = $derived(Math.max(0, Math.min((scrollY || 0) / 50, 1)));
	let domDetailTitle = $state('');

	const websiteNameStore = websiteInfoCtx.selectModelData(
		(data) => data?.website_name || '墨 手记'
	);
	const detailKindStore = detailPanelCtx.selectModelData((data) => data?.kind ?? null);
	const detailTitleStore = detailPanelCtx.selectModelData((data) => data?.title ?? '');
	const detailTocStore = detailPanelCtx.selectModelData((data) => data?.toc ?? []);
	const relatedMomentsStore = detailPanelCtx.selectModelData((data) => data?.relatedMoments ?? []);
	const relatedPostsStore = detailPanelCtx.selectModelData((data) => data?.relatedPosts ?? []);
	const tocContentRootStore = detailPanelCtx.selectModelData((data) => data?.contentRoot ?? null, {
		equals: (a, b) => a === b
	});
	const tocActiveAnchorStore = detailPanelCtx.selectModelData((data) => data?.activeAnchor ?? null);
	const { updateModelData: updateDetailPanel } = detailPanelCtx.useModelActions();

	const currentTitle = $derived($detailTitleStore || domDetailTitle);
	const showPageTitle = $derived(navProgress > 0.5 && Boolean(currentTitle));

	type TocTone = 'jade' | 'cinnabar' | 'ink';
	const currentTocTone = $derived.by<TocTone>(() => {
		if ($detailKindStore === 'moment') return 'cinnabar';
		if ($detailKindStore === 'post') return 'jade';
		return 'ink';
	});
	const hasRelatedContent = $derived(
		($detailKindStore === 'post' && $relatedMomentsStore.length > 0) ||
			($detailKindStore === 'moment' && $relatedPostsStore.length > 0)
	);
	const ownerStatus = $derived(ownerStatusStore.status);
	const ownerOnline = $derived(ownerStatus.ok === 1);
	const adminPanelOnline = $derived(ownerStatus.adminPanelOnline === true);

	function formatOwnerTime(timestamp?: number): string {
		if (!timestamp || !Number.isFinite(timestamp) || timestamp <= 0) return '未知';
		return new Date(timestamp * 1000).toLocaleString('zh-CN', { hour12: false });
	}

	const isActive = (href: string) =>
		page.url.pathname === href || page.url.pathname.startsWith(href + '/');

	const isHomePage = $derived(page.url.pathname === '/' || page.url.pathname === '');

	const isParentActive = (item: NavMenuItem) => {
		if (isActive(item.url)) return true;
		return item.children?.some((child) => isActive(child.url));
	};

	function toggleMobileSubmenu(e: Event, name: string) {
		e.stopPropagation();
		if (expandedMobileItems.includes(name)) {
			expandedMobileItems = expandedMobileItems.filter((item) => item !== name);
		} else {
			expandedMobileItems = [...expandedMobileItems, name];
		}
	}

	function handleNavigate() {
		isMobileMenuOpen = false;
	}

	function formatDate(dateStr: string): string {
		const date = new Date(dateStr);
		return `${date.getMonth() + 1}月${date.getDate()}日`;
	}

	function handleRelatedNavigate() {
		isTocOpen = false;
	}

	const handleTocAnchorChange = (anchor: string) => {
		updateDetailPanel((prev) => {
			if (!prev || prev.activeAnchor === anchor) return prev;
			return { ...prev, activeAnchor: anchor };
		});
		isTocOpen = false;
	};

	$effect(() => {
		const menuOpen = isMobileMenuOpen;
		isMenuAnimating = true;
		const timer = setTimeout(() => (isMenuAnimating = false), menuOpen ? 500 : 300);
		return () => clearTimeout(timer);
	});

	$effect(() => {
		const pathname = page.url.pathname;
		const detailTitle = $detailTitleStore;
		const kind = $detailKindStore;
		if (!browser) {
			domDetailTitle = '';
			return;
		}
		if (detailTitle) {
			domDetailTitle = '';
			return;
		}
		// Only pick up DOM h1 on detail pages (post/moment), not on list/home pages
		if (!kind) {
			domDetailTitle = '';
			return;
		}
		tick().then(() => {
			if (page.url.pathname !== pathname) return;
			const heading = document.querySelector('main article h1, main h1');
			domDetailTitle = heading?.textContent?.trim() ?? '';
		});
	});

	$effect(() => {
		const pathname = page.url.pathname;
		if (pathname) {
			isTocOpen = false;
		}
	});
</script>

<svelte:window bind:scrollY />

<div
	class="fixed z-50 flex justify-center transition-all ease-[cubic-bezier(0.23,1,0.32,1)] md:hidden"
	class:duration-0={!isMobileMenuOpen && !isMenuAnimating}
	class:duration-500={isMobileMenuOpen || isMenuAnimating}
	class:top-0={isMobileMenuOpen}
	class:inset-x-0={isMobileMenuOpen}
	style:top={isMobileMenuOpen ? undefined : `${16 * (1 - navProgress)}px`}
	style:left={isMobileMenuOpen ? undefined : `${16 * (1 - navProgress)}px`}
	style:right={isMobileMenuOpen ? undefined : `${16 * (1 - navProgress)}px`}
>
	<div
		class="relative mx-auto w-full overflow-hidden transition-all ease-[cubic-bezier(0.23,1,0.32,1)]"
		class:duration-0={!isMobileMenuOpen && !isMenuAnimating}
		class:duration-500={isMobileMenuOpen || isMenuAnimating}
		class:shadow-glass-lg={isMobileMenuOpen}
		class:rounded-none={isMobileMenuOpen}
		class:h-screen={isMobileMenuOpen}
		style:border-radius={isMobileMenuOpen ? undefined : `${24 * (1 - navProgress)}px`}
	>
		<!-- Background Layer -->
		<div
			class="shadow-glass absolute inset-0 bg-white/90 backdrop-blur-xl transition-all ease-[cubic-bezier(0.23,1,0.32,1)] dark:bg-ink-900/90"
			class:duration-0={!isMobileMenuOpen && !isMenuAnimating}
			class:duration-500={isMobileMenuOpen || isMenuAnimating}
			style:opacity={1}
			style:height={isMobileMenuOpen ? '100vh' : '3rem'}
			style:min-height={isMobileMenuOpen ? '100vh' : '3rem'}
		></div>

		<!-- 1. Collapsed Header -->
		<div class="relative z-10 flex h-12 items-center justify-between px-3">
			<!-- Left: Avatar & Title -->
			<div class="flex items-center gap-3 overflow-hidden">
					<button
						onclick={(e) => {
							e.stopPropagation();
							isMobileMenuOpen = !isMobileMenuOpen;
						}}
						class="relative flex h-9 w-9 shrink-0 items-center justify-center rounded-full transition-transform active:scale-90"
					>
						<div class="h-8 w-8 shrink-0 overflow-hidden rounded-full border border-ink-100 dark:border-ink-700">
							<img
								src="https://dogeoss.grtsinry43.com/img/author.jpeg"
								alt="Author"
								class="h-full w-full object-cover"
							/>
						</div>
						<span class="absolute bottom-0 right-0 flex h-2.5 w-2.5">
							<span
								class="absolute inline-flex h-full w-full rounded-full opacity-75 {ownerOnline
									? 'bg-jade-400'
									: 'bg-ink-300 dark:bg-ink-600'}"
							></span>
							<span
								class="relative inline-flex h-2.5 w-2.5 rounded-full border border-white dark:border-ink-900 {ownerOnline
									? 'bg-jade-500'
									: 'bg-ink-400 dark:bg-ink-500'}"
							></span>
						</span>
					</button>

				<div
					class="relative min-w-0 flex-1 flex flex-col justify-center transition-all duration-300"
					class:opacity-0={isMobileMenuOpen}
				>
					{#if showPageTitle}
						<span
							transition:fly={{ y: 10, duration: 300 }}
							class="truncate font-serif text-sm font-bold leading-none text-ink-900 dark:text-jade-100"
						>
							{currentTitle}
						</span>
					{:else}
						<span
							transition:fly={{ y: -10, duration: 300 }}
							class="truncate font-serif text-sm font-bold leading-none text-ink-900 dark:text-jade-100"
						>
							{$websiteNameStore}
						</span>
					{/if}
				</div>
			</div>

			<!-- Right: Actions -->
			<div class="flex items-center gap-1">
				{#if $detailTocStore.length > 0 || hasRelatedContent}
					<button
						onclick={(e) => {
							e.stopPropagation();
							isMobileMenuOpen = false;
							isTocOpen = true;
						}}
						class="flex h-9 w-9 items-center justify-center rounded-full text-ink-600 transition-colors hover:bg-black/5 dark:text-ink-300 dark:hover:bg-white/10"
					>
						<List size={20} />
					</button>
				{/if}
			</div>
		</div>

		<!-- 2. Expanded Content -->
		{#if isMobileMenuOpen}
			<div
				transition:fly={{ y: -12, duration: 380, easing: cubicOut, opacity: 0 }}
				class="no-scrollbar relative z-10 flex max-h-[75vh] flex-col overflow-y-auto px-2 pb-6 pt-0"
			>
				<div class="mb-3 rounded-default border border-ink-200 bg-white/70 px-3 py-2 dark:border-ink-700 dark:bg-ink-900/60">
						<div class="flex items-center justify-between gap-2">
							<div class="text-xs font-medium text-ink-700 dark:text-ink-200">
								{ownerOnline ? '站长在线中' : '站长暂时离线'}
							</div>
							<span
								class="rounded-full px-2 py-0.5 text-[10px] {adminPanelOnline
									? 'bg-jade-100 text-jade-700 dark:bg-jade-900/40 dark:text-jade-300'
									: 'bg-ink-100 text-ink-500 dark:bg-ink-800 dark:text-ink-300'}"
							>
								Admin {adminPanelOnline ? '在线' : '离线'}
							</span>
						</div>
						<div class="mt-1 text-[11px] text-ink-500 dark:text-ink-400">
							{#if ownerOnline}
								正在使用 {ownerStatus.process || '未知应用'}
							{:else}
								暂无实时活动
							{/if}
							· {formatOwnerTime(ownerStatus.timestamp)}
						</div>
					</div>

					<div class="flex flex-col gap-1">
					{#if !isHomePage}
						<a
							href={resolve('/')}
							onclick={handleNavigate}
							class="mb-2 flex items-center gap-3 rounded-default border border-ink-200 bg-ink-50/50 px-3 py-2 text-ink-700 transition-colors hover:border-jade-200 hover:bg-jade-50/50 dark:border-ink-700 dark:bg-ink-800/40 dark:text-ink-200 dark:hover:border-jade-800"
						>
							<Home size={16} />
							<span class="text-sm font-medium">返回首页</span>
						</a>
					{/if}
					{#if $userStore.isLogin}
						<button
							type="button"
							onclick={() => {
								handleNavigate();
								windowStore.open('用户中心', null, 'user-center');
							}}
							class="mb-2 flex w-full items-center gap-3 rounded-default border border-jade-200 bg-jade-50/70 px-3 py-2 text-jade-700 dark:border-jade-800 dark:bg-jade-900/30 dark:text-jade-200"
						>
							<User size={16} />
							<span class="text-sm font-medium">用户中心</span>
						</button>
					{:else}
						<button
							type="button"
							class="mb-2 flex items-center gap-3 rounded-default border border-ink-200 bg-white/60 px-3 py-2 text-ink-700 dark:border-ink-700 dark:bg-ink-800/60 dark:text-ink-200"
							onclick={() => {
								isMobileMenuOpen = false;
								authModalStore.open('mobile-nav');
							}}
						>
							<User size={16} />
							<span class="text-sm font-medium">登录</span>
						</button>
					{/if}

					{#each menuTree as item (item.url)}
						{@const active = isParentActive(item)}
						{@const hasChildren = item.children && item.children.length > 0}
						{@const isExpanded = expandedMobileItems.includes(item.name)}

						<div class="flex flex-col">
							<!-- Main Item -->
							<div
								class="relative flex select-none items-center gap-3 overflow-hidden rounded-xl px-3 py-2 transition-all duration-300
                                {active
									? 'bg-white dark:bg-ink-800'
									: 'hover:bg-white/50 dark:hover:bg-ink-800/50'}"
							>
								{#if active}
									<div
										class="pointer-events-none absolute inset-0 rounded-xl border border-jade-200 dark:border-jade-800"
									></div>
								{/if}

								<a
									href={/^(https?:|\/\/)/i.test(item.url) ? item.url : resolve(item.url)}
									onclick={handleNavigate}
									class="flex min-w-0 flex-1 items-center gap-3 text-left"
								>
									<!-- Icon -->
									<div
										class="flex h-8 w-8 shrink-0 items-center justify-center rounded-full transition-colors duration-300
                                        {active
											? 'bg-jade-100 text-jade-700 dark:bg-jade-900 dark:text-jade-300'
											: 'bg-ink-100 text-ink-500 dark:bg-ink-950 dark:text-ink-400'}"
									>
										{#if item.icon}
											<DynamicLucideIcon name={item.icon} className="w-4 h-4" />
										{/if}
									</div>

									<!-- Text -->
									<div class="min-w-0 flex-1">
										<div
											class="truncate font-serif text-[15px] font-medium {active
												? 'text-jade-800 dark:text-jade-100'
												: 'text-ink-700 dark:text-ink-300'}"
										>
											{item.name}
										</div>
									</div>
								</a>

								<!-- Expand/Collapse Button -->
								{#if hasChildren}
									<button
										type="button"
										onclick={(e) => {
											e.preventDefault();
											e.stopPropagation();
											toggleMobileSubmenu(e, item.name);
										}}
										class="-mr-2 rounded-full p-2 text-ink-400 transition-colors active:scale-90 hover:bg-ink-100 dark:hover:bg-white/10"
									>
										<ChevronDown
											size={16}
											class="transition-transform duration-300 {isExpanded
												? 'rotate-180 text-jade-600'
												: ''}"
										/>
									</button>
								{/if}
							</div>

							<!-- Submenu -->
							{#if hasChildren && isExpanded}
								<div
									transition:fly={{ y: -6, duration: 260, easing: cubicOut, opacity: 0 }}
									class="relative mb-2 mt-1 flex flex-col gap-1"
								>
									<!-- Vertical Line -->
									<div
										class="absolute bottom-4 left-[39px] top-0 w-[1px] bg-ink-200 dark:bg-ink-700"
									></div>

									{#each item.children as sub (sub.url)}
										{@const subActive = isActive(sub.url)}
										<a
											href={/^(https?:|\/\/)/i.test(sub.url) ? sub.url : resolve(sub.url)}
											onclick={handleNavigate}
											class="group/sub relative flex items-center gap-3 rounded-lg ml-2 mr-2 py-2.5 pl-[54px] pr-4 text-left transition-colors
                                            {subActive
												? 'bg-jade-50/50 dark:bg-jade-900/20'
												: 'hover:bg-white/60 dark:hover:bg-white/5'}"
										>
											<!-- Horizontal Line -->
											<div
												class="absolute left-[31px] top-1/2 h-[1px] w-4 bg-ink-200 dark:bg-ink-700"
											></div>

											{#if sub.icon}
												<div
													class="{subActive
														? 'text-jade-600 dark:text-jade-400'
														: 'text-ink-400'} transition-colors"
												>
													<DynamicLucideIcon name={sub.icon} className="w-[14px] h-[14px]" />
												</div>
											{/if}
											<span
												class="text-sm font-medium {subActive
													? 'text-jade-700 dark:text-jade-300'
													: 'text-ink-600 dark:text-ink-400'}"
											>
												{sub.name}
											</span>
										</a>
									{/each}
								</div>
							{/if}
						</div>
					{/each}

					<!-- Extra Actions in Menu -->
					<div
						class="mt-4 flex justify-center gap-4 border-t border-ink-200/50 py-4 dark:border-ink-700/50"
					>
						<ThemeIcon />
					</div>
				</div>
			</div>
		{/if}
	</div>

	<!-- Global Overlay -->
	{#if isMobileMenuOpen}
		<div
			transition:fade={{ duration: 300 }}
			class="fixed inset-0 -z-10 bg-ink-900/20 backdrop-blur-[2px]"
			onclick={() => (isMobileMenuOpen = false)}
			role="presentation"
		></div>
	{/if}
</div>

{#if isTocOpen}
	<div
		transition:fade={{ duration: 220 }}
		class="fixed inset-0 z-[55] bg-ink-900/20 backdrop-blur-[2px] md:hidden"
		onclick={() => (isTocOpen = false)}
		role="presentation"
	></div>
	<aside
		transition:fly={{ x: 32, duration: 260, easing: cubicOut, opacity: 0 }}
		class="fixed inset-y-0 right-0 z-[60] h-full w-[82vw] max-w-[360px] border-l border-ink-200/70 bg-white/95 shadow-glass-lg backdrop-blur-xl dark:border-ink-700/70 dark:bg-ink-900/95 md:hidden"
	>
		<div
			class="flex h-14 items-center justify-between border-b border-ink-100/80 px-4 dark:border-ink-800/70"
		>
			<span class="font-mono text-[10px] font-bold tracking-[0.2em] text-ink-400 uppercase">
				本页导航
			</span>
			<button
				class="rounded-full p-1 text-ink-500 transition-colors hover:bg-ink-100 hover:text-ink-900 dark:text-ink-400 dark:hover:bg-ink-800 dark:hover:text-ink-100"
				onclick={() => (isTocOpen = false)}
			>
				<X size={16} />
			</button>
		</div>
		<div class="h-[calc(100%-3.5rem)] overflow-y-auto p-4">
			<div class="space-y-6">
				{#if $detailTocStore.length > 0}
					<DetailTocNavList
						toc={$detailTocStore}
						contentRoot={$tocContentRootStore}
						activeAnchor={$tocActiveAnchorStore}
						onAnchorChange={handleTocAnchorChange}
						tone={currentTocTone}
						size="md"
					/>
				{/if}
				{#if $detailKindStore === 'post' && $relatedMomentsStore.length > 0}
					<section class="space-y-3 border-t border-ink-100/80 pt-4 dark:border-ink-800/70">
						<div class="flex items-center gap-2 text-cinnabar-600 dark:text-cinnabar-400">
							<NotebookPen size={14} />
							<span class="font-mono text-[10px] font-bold tracking-[0.14em] uppercase">
								同期手记
							</span>
						</div>
						<div class="space-y-2.5">
							{#each $relatedMomentsStore.slice(0, 2) as moment (moment.id)}
								<a
									href={resolve(buildMomentPath(moment.shortUrl, moment.createdAt))}
									onclick={handleRelatedNavigate}
									class="block rounded-default border border-ink-100/80 bg-ink-50/40 p-3 transition-colors hover:border-cinnabar-200 hover:bg-white dark:border-ink-700/70 dark:bg-ink-900/40 dark:hover:border-cinnabar-800"
								>
									<div class="mb-1 flex items-center gap-1 text-[11px] text-ink-400">
										<Calendar size={11} />
										{formatDate(moment.createdAt)}
									</div>
									<p
										class="line-clamp-2 text-[13px] font-semibold leading-5 text-ink-700 dark:text-ink-200"
									>
										{moment.title}
									</p>
									<p class="mt-1 line-clamp-2 text-[12px] leading-5 text-ink-500 dark:text-ink-400">
										{moment.summary}
									</p>
								</a>
							{/each}
						</div>
					</section>
				{:else if $detailKindStore === 'moment' && $relatedPostsStore.length > 0}
					<section class="space-y-3 border-t border-ink-100/80 pt-4 dark:border-ink-800/70">
						<div class="flex items-center gap-2 text-jade-600 dark:text-jade-400">
							<FileText size={14} />
							<span class="font-mono text-[10px] font-bold tracking-[0.14em] uppercase">
								同期文章
							</span>
						</div>
						<div class="space-y-2.5">
							{#each $relatedPostsStore.slice(0, 2) as post (post.id)}
								<a
									href={resolve(buildPostPath(post.shortUrl))}
									onclick={handleRelatedNavigate}
									class="block rounded-default border border-ink-100/80 bg-ink-50/40 p-3 transition-colors hover:border-jade-200 hover:bg-white dark:border-ink-700/70 dark:bg-ink-900/40 dark:hover:border-jade-800"
								>
									<div class="mb-1 flex items-center gap-1 text-[11px] text-ink-400">
										<Calendar size={11} />
										{formatDate(post.createdAt)}
									</div>
									<p
										class="line-clamp-2 text-[13px] font-semibold leading-5 text-ink-700 dark:text-ink-200"
									>
										{post.title}
									</p>
									<p class="mt-1 line-clamp-2 text-[12px] leading-5 text-ink-500 dark:text-ink-400">
										{post.summary}
									</p>
								</a>
							{/each}
						</div>
					</section>
				{/if}
			</div>
		</div>
	</aside>
{/if}

<style>
	/* Use reference if needed, though Tailwind classes usually suffice */
	/* @reference "$routes/layout.css"; */

	/* Hide Scrollbar */
	.no-scrollbar::-webkit-scrollbar {
		display: none;
	}
	.no-scrollbar {
		-ms-overflow-style: none;
		scrollbar-width: none;
	}

	.shadow-glass {
		box-shadow: 0 4px 20px rgba(0, 0, 0, 0.03);
	}
	.shadow-glass-lg {
		box-shadow: 0 15px 30px rgba(0, 0, 0, 0.08);
	}
	:global(.dark) .shadow-glass {
		box-shadow: 0 4px 20px rgba(0, 0, 0, 0.15);
	}
</style>
