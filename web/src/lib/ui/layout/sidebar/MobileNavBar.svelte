<script lang="ts">
	import type { NavMenuItem } from '$lib/features/navigation/types';
	import DynamicLucideIcon from '$lib/ui/icons/DynamicLucideIcon.svelte';
	import ThemeIcon from './ThemeIcon.svelte';
	import VisitorAvatar from './VisitorAvatar.svelte';
	// TODO: Implement mobile TableOfContents using tocObserver action
	import { Menu, X, ChevronDown, List } from 'lucide-svelte';
	import { page } from '$app/state';
	import { slide, fade } from 'svelte/transition';
	import { cubicOut } from 'svelte/easing';
	import { onMount } from 'svelte';

	let { menuTree = [] } = $props<{ menuTree: NavMenuItem[] }>();

	let isMobileMenuOpen = $state(false);
	let isTocOpen = $state(false);
	let expandedMobileItems = $state<string[]>([]);
	let scrollY = $state(0);
	let isMenuAnimating = $state(false);

	// Interpolation progress: 0 (top/capsule) -> 1 (scrolled/full)
	let navProgress = $derived(Math.max(0, Math.min(scrollY / 50, 1)));

	const isActive = (href: string) =>
		page.url.pathname === href || page.url.pathname.startsWith(href + '/');

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

	$effect(() => {
		isMobileMenuOpen;
		isMenuAnimating = true;
		const timer = setTimeout(() => (isMenuAnimating = false), 500);
		return () => clearTimeout(timer);
	});
</script>

<svelte:window bind:scrollY />

<div
	class="fixed z-50 flex justify-center transition-all ease-[cubic-bezier(0.23,1,0.32,1)] lg:hidden"
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
			class="shadow-glass absolute inset-0 border-white/40 bg-white/90 backdrop-blur-xl transition-all ease-[cubic-bezier(0.23,1,0.32,1)] dark:border-ink-700 dark:bg-ink-900/90"
			class:duration-0={!isMobileMenuOpen && !isMenuAnimating}
			class:duration-500={isMobileMenuOpen || isMenuAnimating}
			style:opacity={1}
			style:border-width="1px"
			style:height={isMobileMenuOpen ? '100vh' : '3rem'}
			style:min-height={isMobileMenuOpen ? '100vh' : '3rem'}
		></div>

		<!-- 1. Collapsed Header -->
		<div class="relative z-10 flex h-12 items-center justify-between px-3">
			<!-- Left: Avatar & Title -->
			<div class="flex items-center gap-3">
				<button
					onclick={(e) => {
						e.stopPropagation();
						isMobileMenuOpen = !isMobileMenuOpen;
					}}
					class="flex h-9 w-9 items-center justify-center rounded-full transition-transform active:scale-90"
				>
					<div
						class="h-8 w-8 shrink-0 overflow-hidden rounded-full border border-ink-100 dark:border-ink-700"
					>
						<img
							src="https://dogeoss.grtsinry43.com/img/author.jpeg"
							alt="Author"
							class="h-full w-full object-cover"
						/>
					</div>
				</button>

				<div
					class="flex flex-col justify-center transition-all duration-300"
					class:opacity-0={isMobileMenuOpen}
				>
					<span
						class="max-w-[200px] truncate font-serif text-sm font-bold leading-none text-ink-900 dark:text-jade-100"
					>
						墨 手记
					</span>
				</div>
			</div>

			<!-- Right: Actions -->
			<div class="flex items-center gap-1">
				<button
					onclick={(e) => {
						e.stopPropagation();
						isTocOpen = true;
					}}
					class="flex h-9 w-9 items-center justify-center rounded-full text-ink-600 transition-colors hover:bg-black/5 dark:text-ink-300 dark:hover:bg-white/10"
				>
					<List size={20} />
				</button>

				<button
					onclick={(e) => {
						e.stopPropagation();
						isMobileMenuOpen = !isMobileMenuOpen;
					}}
					class="flex h-9 w-9 items-center justify-center rounded-full transition-colors hover:bg-black/5 dark:hover:bg-white/10"
				>
					{#if isMobileMenuOpen}
						<X size={20} class="text-ink-600 dark:text-ink-300" />
					{:else}
						<Menu size={20} class="text-ink-600 dark:text-ink-300" />
					{/if}
				</button>
			</div>
		</div>

		<!-- 2. Expanded Content -->
		{#if isMobileMenuOpen}
			<div
				transition:slide={{ duration: 400, axis: 'y' }}
				class="no-scrollbar relative z-10 flex max-h-[75vh] flex-col overflow-y-auto px-2 pb-6 pt-0"
			>
				<!-- Decoration Header -->
				<div
					class="mb-3 flex items-center justify-between border-b border-ink-200/50 bg-transparent px-4 pb-4 pt-2 dark:border-ink-700/50"
				>
					<span class="text-xs font-bold uppercase tracking-widest text-ink-400">Navigation</span>
					<span class="font-mono text-[10px] text-ink-300">MENU</span>
				</div>

				<div class="flex flex-col gap-1">
					{#each menuTree as item}
						{@const active = isParentActive(item)}
						{@const hasChildren = item.children && item.children.length > 0}
						{@const isExpanded = expandedMobileItems.includes(item.name)}

						<div class="flex flex-col">
							<!-- Main Item -->
							<div
								class="relative flex cursor-pointer select-none items-center gap-3 overflow-hidden rounded-xl px-3 py-2 transition-all duration-300
                                {active
									? 'bg-white dark:bg-ink-800'
									: 'hover:bg-white/50 dark:hover:bg-ink-800/50'}"
							>
								{#if active}
									<div
										class="pointer-events-none absolute inset-0 rounded-xl border border-jade-200 dark:border-jade-800"
									></div>
								{/if}

								<button
									type="button"
									onclick={() =>
										!hasChildren ? handleNavigate() : toggleMobileSubmenu(event!, item.name)}
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
								</button>

								<!-- Expand/Collapse Button -->
								{#if hasChildren}
									<button
										type="button"
										onclick={(e) => toggleMobileSubmenu(e, item.name)}
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
									transition:slide={{ duration: 300, easing: cubicOut }}
									class="relative mb-2 mt-1 flex flex-col gap-1"
								>
									<!-- Vertical Line -->
									<div
										class="absolute bottom-4 left-[39px] top-0 w-[1px] bg-ink-200 dark:bg-ink-700"
									></div>

									{#each item.children as sub}
										{@const subActive = isActive(sub.url)}
										<a
											href={sub.url}
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

<!-- TODO: Mobile TableOfContents (placeholder removed, needs proper implementation) -->

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
