<script lang="ts">
	import { ownerStatusStore } from '$lib/features/owner-status/store.svelte';
	import { websiteInfoCtx } from '$lib/features/website-info/context';
	import { resolveHomeThemeConfig } from '$lib/features/home/theme';
	import OwnerStatusCard from './OwnerStatusCard.svelte';

	const websiteInfoStore = websiteInfoCtx.selectModelData((m) => m ?? null);
	const siteAvatar = $derived(resolveHomeThemeConfig($websiteInfoStore).hero?.avatarUrl || '');

	const isOnline = $derived(ownerStatusStore.status.ok === 1);

	let showDetails = $state(false);
</script>

<div
	class="relative my-4 flex-none"
	role="presentation"
	onmouseenter={() => (showDetails = true)}
	onmouseleave={() => (showDetails = false)}
>
	<div class="nav-author-avatar relative z-10">
		<a href="/" aria-label="返回首页" class="relative block">
			{#if siteAvatar}
				<img
					src={siteAvatar}
					alt="Author"
					class="h-10 w-10 rounded-default object-cover shadow-sm ring-1 ring-ink-200 dark:ring-ink-700"
				/>
			{:else}
				<div
					class="h-10 w-10 rounded-default bg-ink-200 dark:bg-ink-700 ring-1 ring-ink-200 dark:ring-ink-700 shadow-sm"
				></div>
			{/if}
			<span class="absolute -bottom-0.5 -right-0.5 flex h-2.5 w-2.5">
				<span
					class="absolute inline-flex h-full w-full rounded-full opacity-75 {isOnline
						? 'bg-jade-400'
						: 'bg-ink-300 dark:bg-ink-600'}"
				></span>
				<span
					class="relative inline-flex h-2.5 w-2.5 rounded-full border border-white dark:border-ink-900 {isOnline
						? 'bg-jade-500'
						: 'bg-ink-400 dark:bg-ink-500'}"
				></span>
			</span>
		</a>
	</div>

	{#if showDetails}
		<div
			class="absolute left-[calc(100%+0.75rem)] top-0 z-60 w-80 rounded-default border border-ink-200 bg-white/95 p-3 shadow-xl backdrop-blur-sm dark:border-ink-700 dark:bg-ink-900/95"
		>
			<OwnerStatusCard variant="popover" />
		</div>
	{/if}
</div>

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
