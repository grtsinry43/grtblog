<script lang="ts">
	import { onMount } from 'svelte';
	import MarkdownView from '$lib/shared/markdown/MarkdownView.svelte';
	import { flattenTOC, type TOCNode } from '$lib/shared/types/toc';
	import { tocObserver } from '$lib/shared/actions/toc-observer';
	import { detailPanelCtx } from '$lib/shared/detail-panel/context';
	import { scrollToAnchor } from '$lib/shared/dom/scroll-to-anchor';

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

	// On mount, scroll to hash anchor if present (shared links / back-navigation).
	// Wait for fonts to finish loading first — font swap causes layout shift that
	// would make the heading position stale if we scroll too early.
	// Use 'instant' to cleanly override the browser's native smooth hash-scroll.
	onMount(async () => {
		const hash = window.location.hash.replace(/^#/, '');
		if (!hash) return;
		await document.fonts.ready;
		requestAnimationFrame(() => {
			scrollToAnchor(contentRoot, hash, undefined, 'instant');
		});
	});

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
