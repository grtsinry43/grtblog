<script lang="ts">
	import MarkdownView from '$lib/shared/markdown/MarkdownView.svelte';
	import { flattenTOC, type TOCNode } from '$lib/shared/types/toc';
	import { tocObserver } from '$lib/shared/actions/toc-observer';
	import { detailPanelCtx } from '$lib/shared/detail-panel/context';

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
	const { updateModelData } = detailPanelCtx.useModelActions();

	$effect(() => {
		onContentRootChange?.(contentRoot);
		updateModelData((prev) => {
			if (!prev || prev.contentRoot === contentRoot) return prev;
			return { ...prev, contentRoot };
		});
	});

	const handleActiveAnchorChange = (anchor: string | null) => {
		onActiveAnchorChange?.(anchor);
		updateModelData((prev) => {
			if (!prev || prev.activeAnchor === anchor) return prev;
			return { ...prev, activeAnchor: anchor };
		});
	};
</script>

<div
	class={`markdown-preview ${className}`.trim()}
	bind:this={contentRoot}
	use:tocObserver={{ onActiveChange: handleActiveAnchorChange }}
>
	<MarkdownView {content} headingAnchors={flattenTOC(toc ?? [])} />
</div>
