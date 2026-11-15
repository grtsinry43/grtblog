<script lang="ts">
	import type { NavMenuItem } from '$lib/features/navigation/types';
	import OwnerStatusAvatar from '$lib/features/owner-status/components/OwnerStatusAvatar.svelte';
	import DynamicLucideIcon from '$lib/ui/icons/DynamicLucideIcon.svelte';
	import ThemeIcon from './ThemeIcon.svelte';
	import { page } from '$app/state';
	import { fly, scale } from 'svelte/transition';
	import { backOut, cubicOut } from 'svelte/easing';
	import { Spring } from 'svelte/motion';
	import { SearchIcon } from 'lucide-svelte';

	import VisitorAvatar from '$lib/ui/layout/sidebar/VisitorAvatar.svelte';
	import { uiState } from '$lib/shared/stores/ui.svelte';
	import { windowStore } from '$lib/shared/stores/windowStore.svelte';
	import { AppWindowMac } from 'lucide-svelte';
	import { resolveHref } from '$lib/shared/utils/resolve-path';

	let { menuTree = [] } = $props<{ menuTree: NavMenuItem[] }>();

	const isActive = (href: string) =>
		page.url.pathname === href || page.url.pathname.startsWith(href + '/');

	const isParentActive = (item: NavMenuItem) => {
		if (isActive(item.url)) return true;
		return item.children?.some((child) => isActive(child.url));
	};

	let hoveredName = $state<string | null>(null);
	let hoverTimeout: ReturnType<typeof setTimeout>;

	function handleMouseEnter(name: string) {
		clearTimeout(hoverTimeout);
		hoveredName = name;
	}

	function handleMouseLeave() {
		hoverTimeout = setTimeout(() => {
			hoveredName = null;
		}, 100);
	}

	// Sliding jade indicator on the rail edge: rests on the active item, previews the hovered one.
	const INDICATOR_ACTIVE_H = 24;
	const INDICATOR_HOVER_H = 14;
	let itemEls = $state<Record<string, HTMLElement>>({});
	const activeName = $derived(
		menuTree.find((item: NavMenuItem) => isParentActive(item))?.name ?? null
	);
	const indicatorTarget = $derived(hoveredName ?? activeName);
	const indicatorIsPreview = $derived(hoveredName !== null && hoveredName !== activeName);
	const indicatorY = new Spring(0, { stiffness: 0.18, damping: 0.72 });
	const indicatorH = new Spring(0, { stiffness: 0.2, damping: 0.8 });
	let indicatorReady = false;

	$effect(() => {
		const el = indicatorTarget ? itemEls[indicatorTarget] : undefined;
		if (!el) {
			indicatorH.target = 0;
			return;
		}
		const h = indicatorIsPreview ? INDICATOR_HOVER_H : INDICATOR_ACTIVE_H;
		const y = el.offsetTop + (el.offsetHeight - h) / 2;
		const reduced = window.matchMedia('(prefers-reduced-motion: reduce)').matches;
		// First placement snaps so the indicator doesn't fly in from the top on page load.
		const instant = !indicatorReady || reduced;
		indicatorY.set(y, { instant });
		indicatorH.set(h, { instant });
		indicatorReady = true;
	});
</script>

<aside
	class="fixed left-0 top-0 z-50 flex h-full w-24 flex-col items-center border-r border-ink-200 bg-paper py-6 text-ink-600 dark:border-ink-800 dark:bg-ink-900 dark:text-ink-400"
>
	<OwnerStatusAvatar />

	<nav class="relative flex w-full flex-1 flex-col items-center gap-4 px-2 mt-6">
		<span
			aria-hidden="true"
			class="pointer-events-none absolute left-0 top-0 w-[3px] rounded-r-full bg-jade-500 transition-opacity duration-200 dark:bg-jade-400"
			style:transform="translateY({indicatorY.current}px)"
			style:height="{Math.max(0, indicatorH.current)}px"
			style:opacity={indicatorTarget ? (indicatorIsPreview ? 0.45 : 1) : 0}
		></span>
		{#each menuTree as item (item.url)}
			{@const active = isParentActive(item)}
			{@const hasChildren = item.children && item.children.length > 0}
			{@const isHovered = hoveredName === item.name}

			<div
				bind:this={itemEls[item.name]}
				class="relative flex w-full justify-center {isHovered ? 'z-50' : 'z-auto'}"
				role="group"
				onmouseenter={() => handleMouseEnter(item.name)}
				onmouseleave={handleMouseLeave}
			>
				<a
					href={/^(https?:|\/\/)/i.test(item.url) ? item.url : resolveHref(item.url)}
					aria-label={item.name}
					aria-current={active ? 'page' : undefined}
					class="relative z-20 flex h-10 w-10 items-center justify-center rounded-default transition-colors duration-200
                    {active
						? 'bg-jade-500/10 text-jade-700 dark:bg-jade-400/10 dark:text-jade-300'
						: 'hover:bg-ink-100 hover:text-ink-900 dark:hover:bg-ink-800 dark:hover:text-ink-100'}"
				>
					{#if item.icon}
						<DynamicLucideIcon name={item.icon} className="w-5 h-5" />
					{/if}
				</a>

				{#if isHovered}
					{#if hasChildren}
						<div
							class="absolute left-[calc(100%+0.5rem)] top-0 w-48 origin-top-left"
							transition:fly={{ x: -20, duration: 300, easing: cubicOut, opacity: 0 }}
						>
							<div
								class="rounded-default border border-ink-200 bg-white/95 p-1 shadow-xl backdrop-blur-sm dark:border-ink-700 dark:bg-ink-900/95"
							>
								<div
									class="border-b border-ink-100 px-3 py-2 text-xs font-semibold text-ink-400 dark:border-ink-800 dark:text-ink-500"
								>
									{item.name}
								</div>
								<ul class="flex flex-col gap-0.5 py-1">
									{#each item.children as child (child.url)}
										<li>
											<a
												href={/^(https?:|\/\/)/i.test(child.url)
													? child.url
													: resolveHref(child.url)}
												class="flex items-center gap-2 rounded-default px-3 py-2 text-sm transition-colors
                                                {isActive(child.url)
													? 'bg-ink-100 text-ink-900 font-medium dark:bg-ink-800 dark:text-ink-100'
													: 'text-ink-600 hover:bg-ink-50 hover:text-ink-900 dark:text-ink-400 dark:hover:bg-ink-800/50 dark:hover:text-ink-200'}"
											>
												{#if child.icon}
													<DynamicLucideIcon name={child.icon} className="w-4 h-4 opacity-70" />
												{/if}
												<span>{child.name}</span>
											</a>
										</li>
									{/each}
								</ul>
							</div>
						</div>
					{:else}
						<div
							class="absolute left-[calc(100%+0.5rem)] top-1/2 -translate-y-1/2 whitespace-nowrap"
							transition:fly={{ x: 10, duration: 300, easing: cubicOut, opacity: 0 }}
						>
							<div
								class="relative rounded-default border border-ink-200 dark:border-ink-700 bg-ink-50 px-3 py-1.5 text-xs font-serif font-medium text-ink-950 dark:bg-ink-900 dark:text-ink-0"
							>
								{item.name}
								<div
									class="absolute border-b border-l border-ink-200 dark:border-ink-700 -left-1 top-1/2 h-2 w-2 -translate-y-1/2 rotate-45 bg-ink-50 dark:bg-ink-900"
								></div>
							</div>
						</div>
					{/if}
				{/if}
			</div>
		{/each}
	</nav>

	<div class="flex flex-none flex-col items-center gap-6 pb-6 pt-6">
		{#if windowStore.isOpen && windowStore.isMinimized}
			<!-- Dock for the minimized window. Pops in as the genie animation lands
			     (FloatingWindow aims at [data-window-dock]). -->
			<div class="group/dock relative">
				<button
					type="button"
					data-window-dock
					onclick={() => windowStore.restore()}
					aria-label={`恢复窗口：${windowStore.title}`}
					class="relative flex h-10 w-10 items-center justify-center rounded-default border border-ink-200 bg-white text-ink-600 shadow-subtle transition-[color,border-color,box-shadow,translate] duration-200 hover:-translate-y-0.5 hover:border-jade-500/40 hover:text-jade-700 hover:shadow-float dark:border-ink-700 dark:bg-ink-800 dark:text-ink-300 dark:hover:border-jade-400/40 dark:hover:text-jade-300"
					in:scale={{ start: 0.4, duration: 380, delay: 420, easing: backOut }}
					out:scale={{ start: 0.6, duration: 160 }}
				>
					<AppWindowMac class="h-[18px] w-[18px]" strokeWidth={1.75} />
					<span class="absolute -top-1 -right-1 flex h-2.5 w-2.5" aria-hidden="true">
						<span class="dock-ping absolute inset-0 rounded-full bg-jade-400"></span>
						<span
							class="relative h-2.5 w-2.5 rounded-full bg-jade-500 ring-2 ring-paper dark:ring-ink-900"
						></span>
					</span>
				</button>
				<div
					class="pointer-events-none absolute top-1/2 left-[calc(100%+0.5rem)] -translate-x-1 -translate-y-1/2 whitespace-nowrap opacity-0 transition-[opacity,translate] duration-200 group-hover/dock:translate-x-0 group-hover/dock:opacity-100"
					aria-hidden="true"
				>
					<div
						class="relative rounded-default border border-ink-200 bg-ink-50 px-3 py-1.5 font-serif text-xs text-ink-950 dark:border-ink-700 dark:bg-ink-900 dark:text-ink-0"
					>
						恢复「{windowStore.title}」
						<div
							class="absolute top-1/2 -left-1 h-2 w-2 -translate-y-1/2 rotate-45 border-b border-l border-ink-200 bg-ink-50 dark:border-ink-700 dark:bg-ink-900"
						></div>
					</div>
				</div>
			</div>
		{/if}
		<button
			type="button"
			onclick={() => uiState.openSearch()}
			aria-label="搜索"
			class="h-10 w-10 rounded-default text-ink-400 hover:bg-ink-100 hover:text-ink-900 dark:hover:bg-ink-800 dark:hover:text-ink-100 flex items-center justify-center"
		>
			<SearchIcon class="h-5 w-5" />
		</button>
		<ThemeIcon />
		<VisitorAvatar />
	</div>
</aside>

<style lang="postcss">
	@reference "$routes/layout.css";

	/* Two soft rings when the window lands in the dock, then rest — no endless bouncing. */
	.dock-ping {
		animation: dock-ping 1.1s cubic-bezier(0, 0, 0.2, 1) 0.75s 2 both;
	}

	@keyframes dock-ping {
		0% {
			transform: scale(1);
			opacity: 0.7;
		}
		100% {
			transform: scale(2.4);
			opacity: 0;
		}
	}
</style>
