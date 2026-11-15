<script lang="ts">
	import { browser } from '$app/environment';
	import { onMount } from 'svelte';
	import { Wrench, TriangleAlert, X } from 'lucide-svelte';
	import { siteHealthStore } from '../store.svelte';

	let dismissed = $state(false);

	const bannerConfig = $derived.by(() => {
		switch (siteHealthStore.mode) {
			case 'maintenance':
				return {
					text: '站点正在维护中，部分功能可能暂时不可用',
					bg: 'bg-amber-50 dark:bg-amber-950/40 border-amber-200 dark:border-amber-800/60',
					textColor: 'text-amber-800 dark:text-amber-200',
					icon: Wrench,
					iconColor: 'text-amber-500'
				};
			case 'degraded':
				return {
					text: '站点部分服务运行缓慢，我们正在优化中',
					bg: 'bg-amber-50 dark:bg-amber-950/40 border-amber-200 dark:border-amber-800/60',
					textColor: 'text-amber-800 dark:text-amber-200',
					icon: TriangleAlert,
					iconColor: 'text-amber-500'
				};
			case 'critical':
			case 'outage':
				return {
					text: '站点部分服务异常，我们正在紧急处理',
					bg: 'bg-cinnabar-50 dark:bg-cinnabar-950/40 border-cinnabar-200 dark:border-cinnabar-800/60',
					textColor: 'text-cinnabar-800 dark:text-cinnabar-200',
					icon: TriangleAlert,
					iconColor: 'text-cinnabar-500'
				};
			default:
				return null;
		}
	});

	function dismiss() {
		dismissed = true;
		if (browser) {
			try {
				sessionStorage.setItem('health-banner-dismissed', '1');
			} catch {
				// ignore
			}
		}
	}

	// Restore dismissed state from sessionStorage (after mount to avoid hydration mismatch).
	onMount(() => {
		try {
			dismissed = sessionStorage.getItem('health-banner-dismissed') === '1';
		} catch {
			// ignore
		}
	});
</script>

{#if siteHealthStore.showBanner && bannerConfig && !dismissed}
	<div
		role="status"
		class="relative border-b px-4 py-2.5 text-center text-sm transition-colors {bannerConfig.bg}"
	>
		<span class="inline-flex items-center gap-2 {bannerConfig.textColor}">
			<bannerConfig.icon size={16} class="shrink-0 {bannerConfig.iconColor}" aria-hidden="true" />
			{bannerConfig.text}
		</span>
		<button
			class="absolute right-3 top-1/2 -translate-y-1/2 rounded-default p-1.5 opacity-50 transition-opacity hover:opacity-100 {bannerConfig.textColor}"
			onclick={dismiss}
			aria-label="关闭提示"
		>
			<X size={14} />
		</button>
	</div>
{/if}
