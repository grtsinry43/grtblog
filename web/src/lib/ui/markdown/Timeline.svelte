<script lang="ts">
	import type { SvmdComponentNode } from 'svmarkdown';
	import { extractPlainTextFromNodes } from '$lib/shared/markdown/component-body';

	let { node } = $props<{
		node?: SvmdComponentNode;
	}>();

	const title = $derived(node?.props?.title || '');
	const sub = $derived(node?.props?.sub || 'CHRONICLE');
	const bodyText = $derived(extractPlainTextFromNodes(node?.children));

	const items = $derived(
		String(bodyText)
			.split('\n')
			.map((line) => {
				const parts = line.split('|');
				if (parts.length < 2) return null;
				return {
					time: parts[0].trim(),
					title: parts[1].trim(),
					desc: parts[2]?.trim() || ''
				};
			})
			.filter(Boolean)
	);
</script>

<div class="timeline-wrapper not-prose my-10 pl-6">
	<!-- 顶部标题：简洁排版 -->
	<div class="mb-10">
		<span class="text-[10px] font-bold tracking-[0.3em] text-jade-600 uppercase dark:text-jade-400"
			>{sub}</span
		>
		{#if title}
			<h3
				class="mt-1 font-serif text-[22px] font-bold tracking-tight text-ink-900 dark:text-ink-50"
			>
				{title}
			</h3>
		{/if}
	</div>

	<!-- 列表容器 -->
	<div class="relative flex flex-col gap-10">
		<!-- 极细轴线 -->
		<div
			class="absolute left-[-24px] top-2 h-[calc(100%-8px)] w-[1px] bg-ink-200 dark:bg-ink-800"
		></div>

		{#each items as item, index (`${index}-${item!.time}-${item!.title}`)}
			<div class="relative">
				<!-- 极简圆点：仅在 hover 时略微变色 -->
				<div
					class="absolute -left-[27.5px] top-[7px] h-2 w-2 rounded-full border border-jade-500 bg-white ring-4 ring-transparent transition-all duration-300 dark:bg-ink-950"
				></div>

				<div class="flex flex-col">
					<span class="mb-1 text-[11px] font-bold tracking-wider text-ink-400 uppercase"
						>{item!.time}</span
					>
					<h4 class="text-base font-bold text-ink-900 dark:text-ink-50">
						{item!.title}
					</h4>
					{#if item!.desc}
						<p class="mt-2 text-[13.5px] leading-relaxed text-ink-500 dark:text-ink-400">
							{item!.desc}
						</p>
					{/if}
				</div>
			</div>
		{:else}
			<div class="py-2 text-[12px] italic text-ink-400">No events found.</div>
		{/each}
	</div>
</div>

<style>
	/* 仅保留一个极简的交互：选中时圆点实心 */
	.relative:hover div:first-child {
		background-color: var(--color-jade-500, #10b981);
		transform: scale(1.1);
	}
</style>
