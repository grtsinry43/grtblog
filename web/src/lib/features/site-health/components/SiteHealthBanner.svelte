<script lang="ts">
	import { browser } from '$app/environment';
	import { siteHealthStore } from '../store.svelte';

	let dismissed = $state(false);

	const bannerConfig = $derived.by(() => {
		switch (siteHealthStore.mode) {
			case 'maintenance':
				return {
					text: '站点正在维护中，部分功能可能暂时不可用',
					bg: 'bg-amber-50 dark:bg-amber-950/40 border-amber-200 dark:border-amber-800/60',
					textColor: 'text-amber-800 dark:text-amber-200',
					icon: 'i-ph-wrench-bold',
					iconColor: 'text-amber-500'
				};
			case 'degraded':
				return {
					text: '站点部分服务运行缓慢，我们正在优化中',
					bg: 'bg-amber-50 dark:bg-amber-950/40 border-amber-200 dark:border-amber-800/60',
					textColor: 'text-amber-800 dark:text-amber-200',
					icon: 'i-ph-warning-bold',
					iconColor: 'text-amber-500'
				};
			case 'critical':
			case 'outage':
				return {
					text: '站点部分服务异常，我们正在紧急处理',
					bg: 'bg-red-50 dark:bg-red-950/40 border-red-200 dark:border-red-800/60',
					textColor: 'text-red-800 dark:text-red-200',
					icon: 'i-ph-warning-bold',
					iconColor: 'text-red-500'
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

	// Restore dismissed state from sessionStorage.
	if (browser) {
		try {
			dismissed = sessionStorage.getItem('health-banner-dismissed') === '1';
		} catch {
			// ignore
		}
	}
</script>

{#if siteHealthStore.showBanner && bannerConfig && !dismissed}
	<div
		class="relative border-b px-4 py-2.5 text-center text-sm transition-colors {bannerConfig.bg}"
	>
		<span class="inline-flex items-center gap-2 {bannerConfig.textColor}">
			<span class="iconify {bannerConfig.icon} {bannerConfig.iconColor} text-base"></span>
			{bannerConfig.text}
		</span>
		<button
			class="absolute right-3 top-1/2 -translate-y-1/2 rounded p-1 opacity-50 transition-opacity hover:opacity-100 {bannerConfig.textColor}"
			onclick={dismiss}
			aria-label="关闭提示"
		>
			<span class="iconify i-ph-x-bold text-sm"></span>
		</button>
	</div>
{/if}
