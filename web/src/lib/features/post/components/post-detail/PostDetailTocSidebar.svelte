<script lang="ts">
	import DetailTocNavList from '$lib/shared/components/detail/DetailTocNavList.svelte';
	import type { TOCNode } from '$lib/shared/types/toc';
	import Tag from '$lib/ui/primitives/tag/Tag.svelte';

	interface Props {
		toc: TOCNode[];
		contentRoot: HTMLElement | null;
		activeAnchor: string | null;
		onAnchorChange: (anchor: string) => void;
	}

	let { toc, contentRoot, activeAnchor, onAnchorChange }: Props = $props();
</script>

<aside class="hidden slide-in-right lg:block">
	<div class="sticky top-20 space-y-10">
		<div class="space-y-5">
			<span
				class="block border-b border-ink-50 pb-2 font-mono text-[8px] font-bold tracking-[0.4em] text-ink-300 uppercase dark:border-ink-800/30"
			>
				本页目录
			</span>
			<DetailTocNavList {toc} {contentRoot} {activeAnchor} {onAnchorChange} tone="jade" />
		</div>

		<div class="space-y-2.5 rounded-lg border border-jade-500/10 bg-jade-500/5 p-5">
			<Tag variant="jade" class="border-none px-0">感悟</Tag>
			<p class="text-[10px] leading-relaxed font-normal text-ink-500 dark:text-ink-400">
				每一篇文章都是漫长探索中的一小步。如果这些文字能引起共鸣，欢迎留下你的思考。
			</p>
		</div>
	</div>
</aside>
