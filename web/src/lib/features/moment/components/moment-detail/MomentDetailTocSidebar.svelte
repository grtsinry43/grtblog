<script lang="ts">
	import DetailTocNavList from '$lib/shared/components/detail/DetailTocNavList.svelte';
	import type { TOCNode } from '$lib/shared/types/toc';

	interface Props {
		toc: TOCNode[];
		contentRoot: HTMLElement | null;
		activeAnchor: string | null;
		onAnchorChange: (anchor: string) => void;
	}

	let { toc, contentRoot, activeAnchor, onAnchorChange }: Props = $props();
</script>

<aside class="hidden lg:block pt-24 h-full relative slide-in-right">
	<div class="sticky top-24 space-y-10">
		<div class="space-y-5">
			<span
				class="block border-b border-ink-800/10 pb-2 font-mono text-[8px] font-bold tracking-[0.4em] text-ink-400 uppercase"
			>
				目录
			</span>
			<DetailTocNavList {toc} {contentRoot} {activeAnchor} {onAnchorChange} tone="cinnabar" />
		</div>
	</div>
</aside>
