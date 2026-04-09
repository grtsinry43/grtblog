<script lang="ts">
	import type { Snippet } from 'svelte';

	type CalloutType = 'info' | 'warning' | 'error' | 'success' | 'quote' | 'idea';

	let {
		children,
		type = 'info',
		title = ''
	} = $props<{
		type?: CalloutType;
		title?: string;
		children?: Snippet;
	}>();

	const configs = {
		info: {
			icon: 'M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z',
			colorClass: 'text-jade-500',
			bgClass: 'bg-jade-50/40 dark:bg-jade-950/20',
			borderClass: 'border-jade-200/60 dark:border-jade-800/40',
			barClass: 'bg-jade-500'
		},
		warning: {
			icon: 'M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z',
			colorClass: 'text-amber-500',
			bgClass: 'bg-amber-50/40 dark:bg-amber-950/20',
			borderClass: 'border-amber-200/60 dark:border-amber-800/40',
			barClass: 'bg-amber-500'
		},
		error: {
			icon: 'M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z',
			colorClass: 'text-cinnabar-500',
			bgClass: 'bg-cinnabar-50/40 dark:bg-cinnabar-950/20',
			borderClass: 'border-cinnabar-200/60 dark:border-cinnabar-800/40',
			barClass: 'bg-cinnabar-500'
		},
		success: {
			icon: 'M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z',
			colorClass: 'text-bamboo-500',
			bgClass: 'bg-bamboo-50/40 dark:bg-bamboo-950/20',
			borderClass: 'border-bamboo-200/60 dark:border-bamboo-800/40',
			barClass: 'bg-bamboo-500'
		},
		quote: {
			icon: 'M11 19l-7-7 7-7m8 14l-7-7 7-7',
			colorClass: 'text-ink-400',
			bgClass: 'bg-ink-50/40 dark:bg-ink-900/20',
			borderClass: 'border-ink-200/60 dark:border-ink-800/40',
			barClass: 'bg-ink-400'
		},
		idea: {
			icon: 'M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z',
			colorClass: 'text-jade-400',
			bgClass: 'bg-jade-50/40 dark:bg-jade-950/20',
			borderClass: 'border-jade-200/60 dark:border-jade-800/40',
			barClass: 'bg-jade-400'
		}
	} satisfies Record<
		CalloutType,
		{
			icon: string;
			colorClass: string;
			bgClass: string;
			borderClass: string;
			barClass: string;
		}
	>;

	const config = $derived(configs[type as CalloutType] ?? configs.info);
</script>

<div
	class="not-prose relative my-5 flex gap-3 overflow-hidden rounded-default border {config.borderClass} {config.bgClass} p-3.5 transition-colors duration-300"
>
	<!-- 左侧极细装饰线 -->
	<div class="absolute bottom-0 left-0 top-0 w-0.5 {config.barClass} opacity-70"></div>

	<!-- 微型图标 -->
	<div class="mt-0.5 shrink-0 {config.colorClass}">
		<svg class="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor">
			<path d={config.icon} stroke-width="2" stroke-linecap="round" stroke-linejoin="round" />
		</svg>
	</div>

	<!-- 文字内容 -->
	<div class="flex-1 space-y-0.5">
		{#if title}
			<div class="text-[13px] font-semibold tracking-wide {config.colorClass}">
				{title}
			</div>
		{/if}
		<div class="text-[13.5px] leading-relaxed text-ink-700 dark:text-ink-300">
			{#if children}
				{@render children()}
			{/if}
		</div>
	</div>
</div>
