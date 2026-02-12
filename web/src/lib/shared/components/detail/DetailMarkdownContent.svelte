<script lang="ts">
	import MarkdownView from '$lib/shared/markdown/MarkdownView.svelte';
	import { flattenTOC, type TOCNode } from '$lib/shared/types/toc';
	import { tocObserver } from '$lib/shared/actions/toc-observer';

	interface Props {
		content: string;
		toc?: TOCNode[] | null;
		className?: string;
		onActiveAnchorChange?: (anchor: string | null) => void;
		onContentRootChange?: (node: HTMLElement | null) => void;
	}

	let {
		content,
		toc = [],
		className = '',
		onActiveAnchorChange,
		onContentRootChange
	}: Props = $props();

	let contentRoot: HTMLElement | null = $state(null);

	$effect(() => {
		onContentRootChange?.(contentRoot);
	});

	const handleActiveAnchorChange = (anchor: string | null) => {
		onActiveAnchorChange?.(anchor);
	};
</script>

<div
	class={`markdown-preview ${className}`.trim()}
	bind:this={contentRoot}
	use:tocObserver={{ onActiveChange: handleActiveAnchorChange }}
>
	<MarkdownView {content} headingAnchors={flattenTOC(toc ?? [])} />
</div>
