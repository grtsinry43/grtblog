<script lang="ts">
	import { ArrowUpRight } from 'lucide-svelte';
	import type { TimelineItemType, UnifiedTimelineItem } from '../types';

	let { item, side = 'right' } = $props<{
		item: UnifiedTimelineItem;
		side?: 'left' | 'right';
	}>();

	const typeLabels: Record<TimelineItemType, string> = {
		post: '文章',
		moment: '手记',
		thinking: '思考',
		yearSummary: '年度记录'
	};

	const isSummary = $derived(item.type === 'yearSummary');
	const isThinking = $derived(item.type === 'thinking');
	const title = $derived(
		item.title ||
			(item.content
				? `${item.content.slice(0, 76)}${item.content.length > 76 ? '…' : ''}`
				: '未命名内容')
	);
	const formattedDate = $derived(
		new Intl.DateTimeFormat('zh-CN', { month: '2-digit', day: '2-digit' })
			.format(item.publishedAt)
			.replace('/', '.')
	);
	const actionLabel = $derived(
		item.type === 'moment' ? '查看手记' : item.type === 'thinking' ? '继续阅读' : '阅读全文'
	);
</script>

<a
	href={item.url}
	class="group relative block overflow-hidden rounded-default transition-transform duration-300 active:scale-[0.985] motion-reduce:transform-none motion-reduce:transition-none {isSummary
		? 'border border-ink-800 bg-ink-900 p-5 text-white shadow-[6px_6px_0_rgba(20,184,166,0.22)] dark:border-ink-200 dark:bg-ink-100 dark:text-ink-900'
		: `border border-ink-200 bg-ink-0 p-4 shadow-[3px_4px_0_rgba(120,113,108,0.12)] dark:border-ink-700 dark:bg-ink-900 dark:shadow-[3px_4px_0_rgba(0,0,0,0.28)] ${
				item.type === 'thinking'
					? 'border-l-2 border-l-jade-500'
					: item.type === 'moment'
						? 'border-t-2 border-t-amber-400'
						: ''
			}`}
	{side === 'left' ? '-rotate-[0.35deg]' : 'rotate-[0.35deg]'}"
>
	{#if item.image && !isThinking}
		<div
			class="mb-4 aspect-[16/9] overflow-hidden border border-black/5 bg-ink-100 dark:border-white/5 dark:bg-ink-800"
		>
			<img
				src={item.image}
				alt=""
				class="h-full w-full object-cover grayscale-[18%] transition duration-500 group-hover:scale-[1.025] group-hover:grayscale-0 motion-reduce:transition-none"
				loading="lazy"
			/>
		</div>
	{/if}

	<div
		class="flex items-baseline justify-between gap-3 border-b pb-2 {isSummary
			? 'border-white/20 dark:border-ink-900/15'
			: 'border-ink-200/80 dark:border-ink-700/70'}"
	>
		<span
			class="font-mono text-[9px] font-bold tracking-[0.18em] {isSummary
				? 'text-jade-300 dark:text-jade-700'
				: 'text-jade-700 dark:text-jade-400'}"
		>
			{typeLabels[item.type as TimelineItemType]}
		</span>
		<time
			datetime={item.publishedAt.toISOString()}
			class="font-mono text-[11px] font-medium tabular-nums {isSummary
				? 'text-white/55 dark:text-ink-500'
				: 'text-ink-400 dark:text-ink-500'}"
		>
			{formattedDate}
		</time>
	</div>

	<div class="relative pt-4">
		{#if isThinking}
			<span
				class="absolute -left-1 -top-1 select-none font-serif text-4xl leading-none text-jade-500/20"
				aria-hidden="true">“</span
			>
		{/if}
		<h3
			class="relative line-clamp-4 font-serif font-semibold leading-[1.55] {isSummary
				? 'text-lg text-white dark:text-ink-900'
				: isThinking
					? 'pl-3 text-[14px] font-medium italic text-ink-700 dark:text-ink-200'
					: 'text-[15px] text-ink-900 dark:text-ink-100'}"
		>
			{title}
		</h3>
	</div>

	<div
		class="mt-5 flex items-center gap-1.5 font-mono text-[9px] tracking-[0.12em] {isSummary
			? 'text-white/65 dark:text-ink-500'
			: 'text-ink-400 transition-colors group-hover:text-jade-700 dark:text-ink-500 dark:group-hover:text-jade-400'}"
	>
		<span>{actionLabel}</span>
		<ArrowUpRight size={11} />
	</div>
</a>
