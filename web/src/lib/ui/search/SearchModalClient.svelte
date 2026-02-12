<script lang="ts">
	import { createQuery } from '@tanstack/svelte-query';
	import { searchSite } from './service';
	import { onMount, onDestroy } from 'svelte';
	import {
		X,
		Search,
		ArrowRight,
		Clock,
		Hash,
		Map,
		FileText,
		Lightbulb,
		BookOpen
	} from 'lucide-svelte';
	import { uiState } from '$lib/shared/stores/ui.svelte';
	import DynamicLucideIcon from '$lib/ui/icons/DynamicLucideIcon.svelte';
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

	$effect(() => {
		if (query.data && debouncedSearchTerm) {
		}
	});

	function handleInputKeydown(e: KeyboardEvent) {
		if (e.key === 'Enter' && searchTerm) {
			saveHistory(searchTerm);
		}
	}

	function handleResultClick(path: string | null, term?: string) {
		if (!path) return;
		if (term) saveHistory(term);
		else if (debouncedSearchTerm) saveHistory(debouncedSearchTerm);

		closeWithAnimation();
		goto(path);
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

	// Mock suggestions removed, using history or empty state
	// We could keep some default suggestions if needed, but for now lets rely on history.
	const suggestedTags = ['Golang', 'Svelte', 'Design', 'Architecture'];

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
		class="fixed inset-0 z-[100] flex items-start justify-center pt-[15vh] px-4 transition-all duration-500 ease-[cubic-bezier(0.16,1,0.3,1)]"
		class:opacity-100={isAnimating}
		class:opacity-0={!isAnimating}
		class:backdrop-blur-[3px]={isAnimating}
		class:backdrop-blur-none={!isAnimating}
		class:bg-ink-200-30={isAnimating}
		class:bg-transparent={!isAnimating}
		style:background-color={isAnimating ? 'rgba(231, 229, 228, 0.3)' : 'transparent'}
	>
		<!-- Backdrop click handler -->
		<!-- svelte-ignore a11y_click_events_have_key_events -->
		<!-- svelte-ignore a11y_no_static_element_interactions -->
		<div class="absolute inset-0" onclick={onClose}></div>

		<div
			class="relative w-full max-w-2xl bg-ink-50 dark:bg-ink-900 shadow-glass border border-ink-200 dark:border-ink-200/20 rounded-sm overflow-hidden transition-all duration-500 ease-[cubic-bezier(0.16,1,0.3,1)]"
			class:translate-y-0={isAnimating}
			class:translate-y-8={!isAnimating}
			class:scale-100={isAnimating}
			class:scale-95={!isAnimating}
			class:opacity-100={isAnimating}
			class:opacity-0={!isAnimating}
		>
			<!-- Paper Texture Overlay -->
			<div class="absolute inset-0 opacity-[0.03] pointer-events-none bg-noise"></div>

			<!-- Input Header -->
			<div
				class="relative p-4 border-b border-ink-200 dark:border-ink-200/10 flex items-center gap-3 bg-ink-50 dark:bg-[#232323]"
			>
				<Search size={18} class="text-ink-400 dark:text-ink-600 flex-shrink-0" strokeWidth={2} />
				<input
					bind:this={inputRef}
					type="text"
					bind:value={searchTerm}
					onkeydown={handleInputKeydown}
					placeholder="搜索手记与灵感..."
					class="w-full bg-transparent text-lg font-serif text-ink-900 placeholder:text-ink-300 dark:placeholder:text-ink-700 outline-none caret-jade-600 tracking-wide"
				/>
				<div
					class="hidden md:flex items-center gap-2 text-[10px] text-ink-400 border border-ink-200 dark:border-ink-700 px-1.5 py-0.5 rounded-sm font-mono"
				>
					ESC
				</div>
				{#if query.isLoading}
					<Loading size="w-3 h-3" />
				{/if}
				<button onclick={onClose} class="md:hidden text-ink-400 p-1">
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
									class="text-xs font-serif text-ink-800/40 tracking-widest flex items-center gap-2 uppercase"
								>
									<Clock size={12} /> 最近搜索
								</h3>
								<button
									onclick={clearHistory}
									class="text-[10px] text-ink-400 hover:text-jade-600 transition-colors"
									>清除</button
								>
							</div>
							<div class="flex flex-col gap-2">
								{#each searchHistory as item}
									<button
										onclick={() => {
											searchTerm = item;
										}}
										class="text-left px-4 py-3 -mx-4 hover:bg-ink-100 dark:hover:bg-ink-800/50 rounded-sm group transition-colors flex justify-between items-center w-[calc(100%+2rem)]"
									>
										<span
											class="text-ink-800 dark:text-ink-300 font-serif text-sm group-hover:text-jade-600 transition-colors tracking-wide"
											>{item}</span
										>
										<ArrowRight
											size={14}
											class="opacity-0 -translate-x-2 group-hover:opacity-100 group-hover:translate-x-0 transition-all text-ink-400"
										/>
									</button>
								{/each}
							</div>
						</div>
					{/if}

					<!-- Section: Suggestions -->
					<div>
						<h3
							class="text-xs font-serif text-ink-800/40 tracking-widest mb-4 flex items-center gap-2 uppercase"
						>
							<Hash size={12} /> 猜你想找
						</h3>
						<div class="flex flex-wrap gap-2">
							{#each suggestedTags as tag}
								<button
									onclick={() => {
										searchTerm = tag;
									}}
									class="px-3 py-1.5 bg-ink-100 dark:bg-ink-800/30 hover:bg-ink-200 dark:hover:bg-ink-800 hover:text-jade-600 text-ink-800/70 dark:text-ink-400 text-xs font-serif rounded-sm transition-all duration-300 border border-transparent hover:border-ink-300/50 tracking-wide"
								>
									{tag}
								</button>
							{/each}
						</div>
					</div>
				{:else if query.data}
					{#if query.data.articles.length > 0}
						<div class="mb-6">
							<h3
								class="text-xs font-serif text-ink-800/40 tracking-widest mb-3 flex items-center gap-2 uppercase"
							>
								<FileText size={12} /> 文章
							</h3>
							<div class="flex flex-col gap-1">
								{#each query.data.articles as article}
									<button
										onclick={() => handleResultClick(resolveSearchPath('article', article))}
										class="text-left py-2 px-3 -mx-3 rounded-sm hover:bg-ink-100 dark:hover:bg-ink-800/50 transition-colors group"
									>
										<div
											class="font-serif text-ink-900 dark:text-ink-200 group-hover:text-jade-600 transition-colors"
										>
											{@html highlightKeywords(article.title, query.data.keywords)}
										</div>
										<div class="text-xs text-ink-400 mt-0.5 line-clamp-1">
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
								class="text-xs font-serif text-ink-800/40 tracking-widest mb-3 flex items-center gap-2 uppercase"
							>
								<Lightbulb size={12} /> 手记
							</h3>
							<div class="flex flex-col gap-1">
								{#each query.data.moments as moment}
									<button
										onclick={() => handleResultClick(resolveSearchPath('moment', moment))}
										class="text-left py-2 px-3 -mx-3 rounded-sm hover:bg-ink-100 dark:hover:bg-ink-800/50 transition-colors group"
									>
										<div
											class="font-serif text-ink-900 dark:text-ink-200 group-hover:text-jade-600 transition-colors"
										>
											{@html highlightKeywords(moment.title || moment.snippet, query.data.keywords)}
										</div>
										{#if moment.title}<div class="text-xs text-ink-400 mt-0.5 line-clamp-1">
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
								class="text-xs font-serif text-ink-800/40 tracking-widest mb-3 flex items-center gap-2 uppercase"
							>
								<Lightbulb size={12} /> 思考
							</h3>
							<div class="flex flex-col gap-1">
								{#each query.data.thinkings as thinking}
									<button
										onclick={() => handleResultClick(resolveSearchPath('thinking', thinking))}
										class="text-left py-2 px-3 -mx-3 rounded-sm hover:bg-ink-100 dark:hover:bg-ink-800/50 transition-colors group"
									>
										<div
											class="font-serif text-ink-900 dark:text-ink-200 group-hover:text-jade-600 transition-colors"
										>
											{@html highlightKeywords(
												thinking.title || thinking.snippet,
												query.data.keywords
											)}
										</div>
										{#if thinking.title}<div class="text-xs text-ink-400 mt-0.5 line-clamp-1">
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
								class="text-xs font-serif text-ink-800/40 tracking-widest mb-3 flex items-center gap-2 uppercase"
							>
								<BookOpen size={12} /> 页面
							</h3>
							<div class="flex flex-col gap-1">
								{#each query.data.pages as page}
									<button
										onclick={() => handleResultClick(resolveSearchPath('page', page))}
										class="text-left py-2 px-3 -mx-3 rounded-sm hover:bg-ink-100 dark:hover:bg-ink-800/50 transition-colors group"
									>
										<div
											class="font-serif text-ink-900 dark:text-ink-200 group-hover:text-jade-600 transition-colors"
										>
											{@html highlightKeywords(page.title, query.data.keywords)}
										</div>
									</button>
								{/each}
							</div>
						</div>
					{/if}

					{#if query.data.articles.length === 0 && query.data.moments.length === 0 && query.data.pages.length === 0 && query.data.thinkings.length === 0}
						<div class="py-12 flex flex-col items-center justify-center opacity-40">
							<Search size={32} class="mb-2" />
							<span class="font-serif text-sm">未找到相关内容</span>
						</div>
					{/if}
				{:else if query.isError}
					<div class="py-12 flex flex-col items-center justify-center opacity-40 text-red-500">
						<span class="font-serif text-sm">搜索出错，请稍后重试</span>
					</div>
				{/if}

				<!-- Empty State / Decor -->
				{#if !debouncedSearchTerm}
					<div class="mt-8 flex justify-center opacity-20">
						<div class="w-12 h-1 bg-ink-200 rounded-full"></div>
					</div>
				{/if}
			</div>

			<!-- Footer -->
			<div
				class="bg-ink-100/50 dark:bg-[#1a1a1a] px-6 py-3 border-t border-ink-200 dark:border-ink-200/10 flex items-center justify-between text-[10px] text-ink-400 font-sans"
			>
				<div class="flex gap-4">
					<span>输入关键词以搜索</span>
					<span class="hidden md:inline">Enter 保存记录</span>
				</div>
				<span class="font-serif italic opacity-50">墨 · 索引</span>
			</div>
		</div>
	</div>
{/if}
