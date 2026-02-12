<script lang="ts">
	import DetailMarkdownContent from '$lib/shared/components/detail/DetailMarkdownContent.svelte';
	import type { TOCNode } from '$lib/shared/types/toc';

	interface Props {
		content: string;
		toc: TOCNode[];
		onActiveAnchorChange: (anchor: string | null) => void;
		onContentRootChange: (node: HTMLElement | null) => void;
	}

	let { content, toc, onActiveAnchorChange, onContentRootChange }: Props = $props();
</script>

<DetailMarkdownContent
	{content}
	{toc}
	className="markdown-body max-w-none text-[15px] leading-[1.8] font-normal text-ink-800 md:text-base dark:text-ink-200"
	{onContentRootChange}
	{onActiveAnchorChange}
/>
