<script lang="ts">
	import type { NavMenuItem } from '$lib/features/navigation/types';
	import DynamicLucideIcon from '$lib/ui/icons/DynamicLucideIcon.svelte';
	import ThemeIcon from './ThemeIcon.svelte';
	import { page } from '$app/state';
	import { fly } from 'svelte/transition';
	import { cubicOut } from 'svelte/easing';
	import { SearchIcon } from 'lucide-svelte';

	import Button from '$lib/ui/primitives/button/Button.svelte';
	import VisitorAvatar from '$lib/ui/layout/sidebar/VisitorAvatar.svelte';
	import { uiState } from '$lib/shared/stores/ui.svelte';
	import { windowStore } from '$lib/shared/stores/windowStore.svelte';
	import { LayoutIcon } from 'lucide-svelte';

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
</script>

<aside
	class="fixed left-0 top-0 z-50 flex h-full w-24 flex-col items-center border-r border-ink-200 bg-[#FBF9F5] py-6 text-ink-600 dark:border-ink-800 dark:bg-ink-900 dark:text-ink-400"
>
	<div class="relative my-4 flex-none">
		<div class="nav-author-avatar relative z-10">
			<a href="/">
				<img
					src="https://dogeoss.grtsinry43.com/img/author.jpeg"
					alt="Author"
					class="h-10 w-10 rounded-default object-cover shadow-sm ring-1 ring-ink-200 dark:ring-ink-700"
				/>
			</a>
		</div>
	</div>

	<nav class="flex w-full flex-1 flex-col items-center gap-4 px-2 mt-6">
		{#each menuTree as item}
			{@const active = isParentActive(item)}
			{@const hasChildren = item.children && item.children.length > 0}
			{@const isHovered = hoveredName === item.name}

			<div
				class="relative flex w-full justify-center {isHovered ? 'z-50' : 'z-auto'}"
				role="group"
				onmouseenter={() => handleMouseEnter(item.name)}
				onmouseleave={handleMouseLeave}
			>
				<a
					href={item.url}
					class="relative z-20 flex h-10 w-10 items-center justify-center rounded-default transition-all duration-200
                    {active
						? 'bg-ink-900 text-white shadow-sm dark:bg-ink-100 dark:text-ink-950'
						: 'hover:bg-ink-200 hover:text-ink-900 dark:hover:bg-ink-800 dark:hover:text-ink-100'}"
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
									{#each item.children as child}
										<li>
											<a
												href={child.url}
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
			<Button
				variant="icon"
				onclick={() => windowStore.restore()}
				class="h-10 w-10 rounded-default bg-jade-500 text-white shadow-lg animate-bounce duration-[2000ms] transition-all"
				title="恢复窗口"
			>
				<LayoutIcon class="h-5 w-5" />
			</Button>
		{/if}
		<Button
			variant="icon"
			onclick={() => uiState.openSearch()}
			class="h-10 w-10 rounded-default text-ink-400 hover:bg-ink-100 hover:text-ink-900 dark:hover:bg-ink-800 dark:hover:text-ink-100"
		>
			<SearchIcon class="h-5 w-5" />
		</Button>
		<ThemeIcon />
		<VisitorAvatar />
	</div>
</aside>

<style lang="postcss">
	@reference "$routes/layout.css";

	.nav-author-avatar::before {
		content: '';
		@apply absolute inset-0 -z-10 translate-x-1 translate-y-1 rounded-default border border-ink-300 dark:border-ink-900/30;
	}

	.nav-author-avatar:hover::before {
		content: '';
		@apply absolute inset-0 -z-10 translate-x-0.5 translate-y-0.5 rounded-default border border-ink-300 transition dark:border-ink-900/30;
	}
</style>
