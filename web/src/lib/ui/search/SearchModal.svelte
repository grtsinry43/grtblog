<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { X, Search, ArrowRight, Clock, Hash } from 'lucide-svelte';
	import { uiState } from '$lib/shared/stores/ui.svelte';
	import DynamicLucideIcon from '$lib/ui/icons/DynamicLucideIcon.svelte';

	let shouldRender = $state(false);
	let isAnimating = $state(false);
	let inputRef: HTMLInputElement | null = $state(null);
	let timer: ReturnType<typeof setTimeout>;

	// Handle open/close transitions
	$effect(() => {
		if (uiState.isSearchOpen) {
			shouldRender = true;
			document.body.style.overflow = 'hidden';
			// Slight delay to allow DOM to mount before starting transition
			timer = setTimeout(() => {
				isAnimating = true;
				// Focus input after animation starts
				setTimeout(() => inputRef?.focus(), 100);
			}, 50);
		} else {
			isAnimating = false;
			document.body.style.overflow = '';
			// Wait for transition to finish before unmounting
			timer = setTimeout(() => {
				shouldRender = false;
			}, 500); // Duration matches CSS transition
		}
		return () => clearTimeout(timer);
	});

	function onClose() {
		uiState.closeSearch();
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') {
			onClose();
		}
	}

	// Mock data
	const recentSearches = ['React 并发模式', '极简主义设计'];
	const suggestedTags = ['摄影', '代码', '生活', '阅读', '咖啡', '夏日'];
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
				class="relative p-6 md:p-8 border-b border-ink-200 dark:border-ink-200/10 flex items-center gap-4 bg-ink-50 dark:bg-[#232323]"
			>
				<Search size={22} class="text-ink-400 dark:text-ink-600 flex-shrink-0" strokeWidth={2} />
				<input
					bind:this={inputRef}
					type="text"
					placeholder="搜索手记与灵感..."
					class="w-full bg-transparent text-xl md:text-2xl font-serif text-ink-900 placeholder:text-ink-300 dark:placeholder:text-ink-700 outline-none caret-cinnabar-600 tracking-wide"
				/>
				<div
					class="hidden md:flex items-center gap-2 text-[10px] text-ink-400 border border-ink-200 dark:border-ink-700 px-2 py-1 rounded-sm font-mono"
				>
					ESC
				</div>
				<button onclick={onClose} class="md:hidden text-ink-400 p-1">
					<X size={20} />
				</button>
			</div>

			<!-- Content Body -->
			<div class="p-6 md:p-8 min-h-[300px] max-h-[60vh] overflow-y-auto no-scrollbar">
				<!-- Section: Recent -->
				<div class="mb-8">
					<h3
						class="text-xs font-serif text-ink-800/40 tracking-widest mb-4 flex items-center gap-2 uppercase"
					>
						<Clock size={12} /> 最近搜索
					</h3>
					<div class="flex flex-col gap-2">
						{#each recentSearches as item}
							<button
								class="text-left px-4 py-3 -mx-4 hover:bg-ink-100 dark:hover:bg-ink-800/50 rounded-sm group transition-colors flex justify-between items-center w-[calc(100%+2rem)]"
							>
								<span
									class="text-ink-800 dark:text-ink-300 font-serif text-sm group-hover:text-cinnabar-600 transition-colors tracking-wide"
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

				<!-- Section: Tags -->
				<div>
					<h3
						class="text-xs font-serif text-ink-800/40 tracking-widest mb-4 flex items-center gap-2 uppercase"
					>
						<Hash size={12} /> 猜你想找
					</h3>
					<div class="flex flex-wrap gap-2">
						{#each suggestedTags as tag}
							<button
								class="px-3 py-1.5 bg-ink-100 dark:bg-ink-800/30 hover:bg-ink-200 dark:hover:bg-ink-800 hover:text-cinnabar-600 text-ink-800/70 dark:text-ink-400 text-xs font-serif rounded-sm transition-all duration-300 border border-transparent hover:border-ink-300/50 tracking-wide"
							>
								{tag}
							</button>
						{/each}
					</div>
				</div>

				<!-- Empty State / Decor -->
				<div class="mt-16 flex justify-center opacity-20">
					<div class="w-12 h-1 bg-ink-200 rounded-full"></div>
				</div>
			</div>

			<!-- Footer -->
			<div
				class="bg-ink-100/50 dark:bg-[#1a1a1a] px-6 py-3 border-t border-ink-200 dark:border-ink-200/10 flex items-center justify-between text-[10px] text-ink-400 font-sans"
			>
				<div class="flex gap-4">
					<span>输入关键词以搜索</span>
					<span class="hidden md:inline">↑↓ 切换选中</span>
				</div>
				<span class="font-serif italic opacity-50">墨 · 索引</span>
			</div>
		</div>
	</div>
{/if}
