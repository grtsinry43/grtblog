<script lang="ts">
	import { resolve } from '$app/paths';
	import { createQuery } from '@tanstack/svelte-query';
	import { searchSite } from './service';
	import { onMount } from 'svelte';
	import { X, Search, ArrowRight, Clock, FileText, Lightbulb, BookOpen } from 'lucide-svelte';
	import { uiState } from '$lib/shared/stores/ui.svelte';
	import { goto } from '$app/navigation';
	import Loading from '$lib/ui/common/Loading.svelte';
	import { buildMomentPath, buildPagePath, buildPostPath } from '$lib/shared/utils/content-path';
	import type { SiteSearchItemResp } from './types';

	let shouldRender = $state(false);
	let isAnimating = $state(false);
	let inputRef: HTMLInputElement | null = $state(null);
	let timer: ReturnType<typeof setTimeout>;
	let searchTerm = $state('');
	let debouncedSearchTerm = $state('');
	let searchHistory: string[] = $state([]);
	let activeItemIndex = $state(-1);

	type KeyboardNavItem = {
		key: string;
		mode: 'history' | 'result';
		term?: string;
		path?: string | null;
	};

	// Debounce logic
	let debounceTimer: ReturnType<typeof setTimeout>;
	$effect(() => {
		const term = searchTerm; // Ensure reactivity tracking
		clearTimeout(debounceTimer);
		debounceTimer = setTimeout(() => {
			debouncedSearchTerm = term.trim();
		}, 500);
		return () => clearTimeout(debounceTimer);
	});

	// Search Query
	const query = createQuery(() => {
		return {
			queryKey: ['site-search', debouncedSearchTerm],
			queryFn: () => {
				return searchSite(undefined, debouncedSearchTerm);
			},
			enabled: !!debouncedSearchTerm,
			staleTime: 1000 * 60 * 5 // 5 minutes
		};
	});

	// Handle open/close transitions
	$effect(() => {
		shouldRender = true;
		document.body.style.overflow = 'hidden';
		// Slight delay to allow DOM to mount before starting transition
		timer = setTimeout(() => {
			isAnimating = true;
			// Focus input after animation starts
			setTimeout(() => inputRef?.focus(), 100);
		}, 50);

		return () => {
			clearTimeout(timer);
			document.body.style.overflow = '';
		};
	});

	function closeWithAnimation() {
		isAnimating = false;
		document.body.style.overflow = '';
		setTimeout(() => {
			uiState.closeSearch();
		}, 500);
	}

	function onClose() {
		closeWithAnimation();
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') {
			onClose();
		}
	}

	// History Logic
	onMount(() => {
		const history = localStorage.getItem('search_history');
		if (history) {
			try {
				searchHistory = JSON.parse(history);
			} catch (e) {
				console.error('Failed to parse search history', e);
			}
		}
	});

	function saveHistory(term: string) {
		if (!term) return;
		const newHistory = [term, ...searchHistory.filter((feed) => feed !== term)].slice(0, 10);
		searchHistory = newHistory;
		localStorage.setItem('search_history', JSON.stringify(newHistory));
	}

	function handleInputKeydown(e: KeyboardEvent) {
		if (e.key === 'ArrowDown') {
			e.preventDefault();
			moveActiveIndex(1);
			return;
		}
		if (e.key === 'ArrowUp') {
			e.preventDefault();
			moveActiveIndex(-1);
			return;
		}
		if (e.key === 'Enter') {
			if (confirmActiveItem()) {
				e.preventDefault();
				return;
			}
		}
		if (e.key === 'Enter' && searchTerm) {
			saveHistory(searchTerm);
		}
	}

	function clearSearchTerm() {
		searchTerm = '';
		debouncedSearchTerm = '';
		activeItemIndex = -1;
		inputRef?.focus();
	}

	function handleResultClick(path: string | null, term?: string) {
		if (!path) return;
		if (term) saveHistory(term);
		else if (debouncedSearchTerm) saveHistory(debouncedSearchTerm);

		closeWithAnimation();
		goto(resolve(path));
	}

	const resolveSearchPath = (
		kind: 'article' | 'moment' | 'page' | 'thinking',
		item: SiteSearchItemResp
	): string | null => {
		if (kind === 'article') {
			return item.shortUrl ? buildPostPath(item.shortUrl) : null;
		}
		if (kind === 'moment') {
			return item.shortUrl ? buildMomentPath(item.shortUrl, item.createdAt) : null;
		}
		if (kind === 'page') {
			return item.shortUrl ? buildPagePath(item.shortUrl) : null;
		}
		return item.path;
	};

	function clearHistory() {
		searchHistory = [];
		localStorage.removeItem('search_history');
	}

	const keyboardNavItems = $derived.by(() => {
		const items: KeyboardNavItem[] = [];
		if (!debouncedSearchTerm) {
			for (const term of searchHistory) {
				items.push({
					key: `history:${term}`,
					mode: 'history',
					term
				});
			}
			return items;
		}
		if (!query.data) {
			return items;
		}

		for (const article of query.data.articles) {
			items.push({
				key: `article:${article.id}`,
				mode: 'result',
				path: resolveSearchPath('article', article)
			});
		}
		for (const moment of query.data.moments) {
			items.push({
				key: `moment:${moment.id}`,
				mode: 'result',
				path: resolveSearchPath('moment', moment)
			});
		}
		for (const thinking of query.data.thinkings) {
			items.push({
				key: `thinking:${thinking.id}`,
				mode: 'result',
				path: resolveSearchPath('thinking', thinking)
			});
		}
		for (const page of query.data.pages) {
			items.push({
				key: `page:${page.id}`,
				mode: 'result',
				path: resolveSearchPath('page', page)
			});
		}

		return items;
	});

	const keyboardNavIndexMap = $derived.by(() => {
		const map = new Map<string, number>();
		keyboardNavItems.forEach((item, index) => {
			map.set(item.key, index);
		});
		return map;
	});

	$effect(() => {
		debouncedSearchTerm;
		activeItemIndex = -1;
	});

	$effect(() => {
		const length = keyboardNavItems.length;
		if (length === 0) {
			activeItemIndex = -1;
			return;
		}
		if (activeItemIndex >= length) {
			activeItemIndex = 0;
		}
	});

	function moveActiveIndex(direction: 1 | -1) {
		const length = keyboardNavItems.length;
		if (!length) return;
		if (activeItemIndex < 0) {
			activeItemIndex = direction === 1 ? 0 : length - 1;
			return;
		}
		activeItemIndex = (activeItemIndex + direction + length) % length;
	}

	function confirmActiveItem() {
		if (activeItemIndex < 0) return false;
		const current = keyboardNavItems[activeItemIndex];
		if (!current) return false;
		if (current.mode === 'history' && current.term) {
			searchTerm = current.term;
			return true;
		}
		if (current.mode === 'result' && current.path) {
			handleResultClick(current.path);
			return true;
		}
		return false;
	}

	function isItemActive(key: string) {
		return keyboardNavIndexMap.get(key) === activeItemIndex;
	}

	function setActiveItem(key: string) {
		const index = keyboardNavIndexMap.get(key);
		if (index === undefined) return;
		activeItemIndex = index;
	}

	// Keyword Highlighting Helper
	function highlightKeywords(text: string, keywords: string[]): string {
		if (!keywords || keywords.length === 0) return text;
		let highlighted = text;
		// Sort keywords by length (descending) to match longest phrases first
		const sortedKeywords = [...keywords].sort((a, b) => b.length - a.length);

		// Escape special regex characters in keywords
		const escapedKeywords = sortedKeywords.map((k) => k.replace(/[.*+?^${}()|[\]\\]/g, '\\$&'));

		// Create a regex to match all keywords, case-insensitive
		const regex = new RegExp(`(${escapedKeywords.join('|')})`, 'gi');

		highlighted = highlighted.replace(
			regex,
			'<span class="text-jade-600 dark:text-jade-400 font-medium">$1</span>'
		);
		return highlighted;
	}
</script>

<svelte:window onkeydown={handleKeydown} />

{#if shouldRender}
	<div
		class="fixed inset-0 z-[100] flex items-start justify-center bg-ink-900/20 dark:bg-black/60 pt-[15vh] px-4 transition-all duration-500 ease-[cubic-bezier(0.16,1,0.3,1)]"
		class:opacity-100={isAnimating}
		class:opacity-0={!isAnimating}
		class:backdrop-blur-[3px]={isAnimating}
		class:backdrop-blur-none={!isAnimating}
	>
		<button type="button" class="absolute inset-0" aria-label="关闭搜索弹窗" onclick={onClose}
		></button>

		<div
			class="relative w-full max-w-2xl overflow-hidden rounded-sm border border-ink-200/80 bg-ink-50/95 shadow-glass dark:border-ink-700/70 dark:bg-ink-900/95 transition-all duration-500 ease-[cubic-bezier(0.16,1,0.3,1)]"
			class:translate-y-0={isAnimating}
			class:translate-y-8={!isAnimating}
			class:scale-100={isAnimating}
			class:scale-95={!isAnimating}
			class:opacity-100={isAnimating}
			class:opacity-0={!isAnimating}
		>
			<!-- Input Header -->
			<div
				class="relative flex items-center gap-3 border-b border-ink-200/80 bg-ink-50 px-4 py-4 dark:border-ink-700/60 dark:bg-ink-900"
			>
				<Search size={18} class="text-ink-400 dark:text-ink-500 flex-shrink-0" strokeWidth={2} />
				<input
					bind:this={inputRef}
					type="text"
					bind:value={searchTerm}
					onkeydown={handleInputKeydown}
					placeholder="探索知识与灵感..."
					class="w-full bg-transparent text-lg font-serif text-ink-900 dark:text-ink-100 placeholder:text-ink-400 dark:placeholder:text-ink-500 outline-none caret-jade-600 tracking-wide"
				/>
				{#if searchTerm}
					<button
						type="button"
						onclick={clearSearchTerm}
						class="hidden md:flex items-center justify-center rounded-sm p-1 text-ink-400 transition-colors hover:text-ink-700 dark:text-ink-500 dark:hover:text-ink-300"
						aria-label="清空搜索关键词"
						title="清空"
					>
						<X size={14} />
					</button>
				{/if}
				{#if query.isLoading}
					<Loading size="w-3 h-3" />
				{/if}
				<button onclick={onClose} class="md:hidden p-1 text-ink-500 dark:text-ink-400">
					<X size={18} />
				</button>
			</div>

			<!-- Content Body -->
			<div class="p-6 md:p-8 min-h-[300px] max-h-[60vh] overflow-y-auto no-scrollbar">
				{#if !debouncedSearchTerm}
					<!-- Section: Recent -->
					{#if searchHistory.length > 0}
						<div class="mb-8">
							<div class="flex items-center justify-between mb-4">
								<h3
									class="flex items-center gap-2 text-xs font-serif uppercase tracking-widest text-ink-500 dark:text-ink-400"
								>
									<Clock size={12} /> 最近搜索
								</h3>
								<button
									onclick={clearHistory}
									class="text-[10px] text-ink-400 hover:text-jade-600 dark:text-ink-500 transition-colors"
									>清除</button
								>
							</div>
							<div class="flex flex-col gap-2">
								{#each searchHistory as item (item)}
									<button
										onclick={() => {
											searchTerm = item;
										}}
										onmouseenter={() => setActiveItem(`history:${item}`)}
										class="group flex w-[calc(100%+2rem)] items-center justify-between rounded-sm px-4 py-3 -mx-4 text-left transition-colors hover:bg-ink-100 dark:hover:bg-ink-800/50 {isItemActive(
											`history:${item}`
										)
											? 'bg-ink-100 dark:bg-ink-800/50'
											: ''}"
									>
										<span
											class="font-serif text-sm tracking-wide text-ink-700 dark:text-ink-300 group-hover:text-jade-600 transition-colors"
											>{item}</span
										>
										<ArrowRight
											size={14}
											class="text-ink-400 dark:text-ink-500 opacity-0 -translate-x-2 group-hover:opacity-100 group-hover:translate-x-0 transition-all"
										/>
									</button>
								{/each}
							</div>
						</div>
					{:else}
						<div class="py-12 flex flex-col items-center justify-center text-center">
							<Search size={24} class="mb-3 text-ink-300 dark:text-ink-600" />
							<p class="font-serif text-sm text-ink-700 dark:text-ink-200">输入关键词开始搜索</p>
							<p class="mt-1 text-xs text-ink-400 dark:text-ink-500">支持文章、手记、页面、思考</p>
						</div>
					{/if}
				{:else if query.data}
					{#if query.data.articles.length > 0}
						<div class="mb-6">
							<h3
								class="mb-3 flex items-center gap-2 text-xs font-serif uppercase tracking-widest text-ink-500 dark:text-ink-400"
							>
								<FileText size={12} /> 文章
							</h3>
							<div class="flex flex-col gap-1">
								{#each query.data.articles as article (article.id)}
									<button
										onclick={() => handleResultClick(resolveSearchPath('article', article))}
										onmouseenter={() => setActiveItem(`article:${article.id}`)}
										class="text-left py-2 px-3 -mx-3 rounded-sm hover:bg-ink-100 dark:hover:bg-ink-800/50 transition-colors group {isItemActive(
											`article:${article.id}`
										)
											? 'bg-ink-100 dark:bg-ink-800/50'
											: ''}"
									>
										<div
											class="font-serif text-ink-900 dark:text-ink-200 group-hover:text-jade-600 transition-colors"
										>
											<!-- eslint-disable-next-line svelte/no-at-html-tags -->
											{@html highlightKeywords(article.title, query.data.keywords)}
										</div>
										<div class="text-xs text-ink-400 mt-0.5 line-clamp-1">
											<!-- eslint-disable-next-line svelte/no-at-html-tags -->
											{@html highlightKeywords(article.snippet, query.data.keywords)}
										</div>
									</button>
								{/each}
							</div>
						</div>
					{/if}

					{#if query.data.moments.length > 0}
						<div class="mb-6">
							<h3
								class="mb-3 flex items-center gap-2 text-xs font-serif uppercase tracking-widest text-ink-500 dark:text-ink-400"
							>
								<Lightbulb size={12} /> 手记
							</h3>
							<div class="flex flex-col gap-1">
								{#each query.data.moments as moment (moment.id)}
									<button
										onclick={() => handleResultClick(resolveSearchPath('moment', moment))}
										onmouseenter={() => setActiveItem(`moment:${moment.id}`)}
										class="text-left py-2 px-3 -mx-3 rounded-sm hover:bg-ink-100 dark:hover:bg-ink-800/50 transition-colors group {isItemActive(
											`moment:${moment.id}`
										)
											? 'bg-ink-100 dark:bg-ink-800/50'
											: ''}"
									>
										<div
											class="font-serif text-ink-900 dark:text-ink-200 group-hover:text-jade-600 transition-colors"
										>
											<!-- eslint-disable-next-line svelte/no-at-html-tags -->
											{@html highlightKeywords(moment.title || moment.snippet, query.data.keywords)}
										</div>
										{#if moment.title}<div class="text-xs text-ink-400 mt-0.5 line-clamp-1">
												<!-- eslint-disable-next-line svelte/no-at-html-tags -->
												{@html highlightKeywords(moment.snippet, query.data.keywords)}
											</div>{/if}
									</button>
								{/each}
							</div>
						</div>
					{/if}

					{#if query.data.thinkings.length > 0}
						<div class="mb-6">
							<h3
								class="mb-3 flex items-center gap-2 text-xs font-serif uppercase tracking-widest text-ink-500 dark:text-ink-400"
							>
								<Lightbulb size={12} /> 思考
							</h3>
							<div class="flex flex-col gap-1">
								{#each query.data.thinkings as thinking (thinking.id)}
									<button
										onclick={() => handleResultClick(resolveSearchPath('thinking', thinking))}
										onmouseenter={() => setActiveItem(`thinking:${thinking.id}`)}
										class="text-left py-2 px-3 -mx-3 rounded-sm hover:bg-ink-100 dark:hover:bg-ink-800/50 transition-colors group {isItemActive(
											`thinking:${thinking.id}`
										)
											? 'bg-ink-100 dark:bg-ink-800/50'
											: ''}"
									>
										<div
											class="font-serif text-ink-900 dark:text-ink-200 group-hover:text-jade-600 transition-colors"
										>
											<!-- eslint-disable-next-line svelte/no-at-html-tags -->
											{@html highlightKeywords(
												thinking.title || thinking.snippet,
												query.data.keywords
											)}
										</div>
										{#if thinking.title}<div class="text-xs text-ink-400 mt-0.5 line-clamp-1">
												<!-- eslint-disable-next-line svelte/no-at-html-tags -->
												{@html highlightKeywords(thinking.snippet, query.data.keywords)}
											</div>{/if}
									</button>
								{/each}
							</div>
						</div>
					{/if}

					{#if query.data.pages.length > 0}
						<div class="mb-6">
							<h3
								class="mb-3 flex items-center gap-2 text-xs font-serif uppercase tracking-widest text-ink-500 dark:text-ink-400"
							>
								<BookOpen size={12} /> 页面
							</h3>
							<div class="flex flex-col gap-1">
								{#each query.data.pages as page (page.id)}
									<button
										onclick={() => handleResultClick(resolveSearchPath('page', page))}
										onmouseenter={() => setActiveItem(`page:${page.id}`)}
										class="text-left py-2 px-3 -mx-3 rounded-sm hover:bg-ink-100 dark:hover:bg-ink-800/50 transition-colors group {isItemActive(
											`page:${page.id}`
										)
											? 'bg-ink-100 dark:bg-ink-800/50'
											: ''}"
									>
										<div
											class="font-serif text-ink-900 dark:text-ink-200 group-hover:text-jade-600 transition-colors"
										>
											<!-- eslint-disable-next-line svelte/no-at-html-tags -->
											{@html highlightKeywords(page.title, query.data.keywords)}
										</div>
									</button>
								{/each}
							</div>
						</div>
					{/if}

					{#if query.data.articles.length === 0 && query.data.moments.length === 0 && query.data.pages.length === 0 && query.data.thinkings.length === 0}
						<div
							class="py-12 flex flex-col items-center justify-center text-ink-500 dark:text-ink-400"
						>
							<Search size={32} class="mb-2 opacity-60" />
							<span class="font-serif text-sm">未找到相关内容</span>
						</div>
					{/if}
				{:else if query.isError}
					<div
						class="py-12 flex flex-col items-center justify-center text-cinnabar-600 dark:text-cinnabar-400"
					>
						<span class="font-serif text-sm">搜索出错，请稍后重试</span>
					</div>
				{/if}

				<!-- Empty State / Decor -->
				{#if !debouncedSearchTerm}
					<div class="mt-8 flex justify-center opacity-40">
						<div class="w-12 h-1 rounded-full bg-ink-200 dark:bg-ink-700"></div>
					</div>
				{/if}
			</div>

			<!-- Footer -->
			<div
				class="flex items-center justify-between border-t border-ink-200/80 bg-ink-100/60 px-6 py-3 font-sans text-[10px] text-ink-500 dark:border-ink-700/60 dark:bg-ink-900 dark:text-ink-400"
			>
				<div class="flex gap-4">
					<span>输入关键词以搜索</span>
					<span class="hidden md:inline">↑↓ 切换 · Enter 确认</span>
				</div>
				<span class="font-serif italic opacity-50">搜索由 PostgreSQL 提供支持</span>
			</div>
		</div>
	</div>
{/if}
