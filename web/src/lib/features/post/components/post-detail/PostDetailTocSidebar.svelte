<script lang="ts">
	import DetailTocNavList from '$lib/ui/detail/DetailTocNavList.svelte';
	import type { TOCNode } from '$lib/shared/types/toc';
	import PostDetailRelatedMoments from './PostDetailRelatedMoments.svelte';

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

		<PostDetailRelatedMoments />
	</div>
</aside>
