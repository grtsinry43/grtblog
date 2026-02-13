<script lang="ts">
	import { scrollToAnchor } from '$lib/shared/dom/scroll-to-anchor';
	import type { TOCNode } from '$lib/shared/types/toc';
	import { fly } from 'svelte/transition';

	type TocTone = 'jade' | 'cinnabar' | 'ink';
	type TocSize = 'sm' | 'md';

	interface Props {
		toc: TOCNode[];
		contentRoot: HTMLElement | null;
		activeAnchor: string | null;
		onAnchorChange?: (anchor: string) => void;
		tone?: TocTone;
		size?: TocSize;
	}

	type ToneConfig = {
		parentHover: string;
		parentActive: string;
		childHover: string;
		childActive: string;
		sublistBorder: string;
	};

	const toneClassMap: Record<TocTone, ToneConfig> = {
		jade: {
			parentHover: 'hover:translate-x-0.5 hover:text-jade-600 dark:hover:text-jade-400',
			parentActive: 'font-bold text-jade-700 dark:text-jade-400',
			childHover: 'hover:translate-x-0.5 hover:text-jade-500',
			childActive: 'font-bold text-jade-600 dark:text-jade-300',
			sublistBorder: 'border-ink-50 dark:border-ink-800/30'
		},
		cinnabar: {
			parentHover: 'hover:translate-x-0.5 hover:text-cinnabar-600 dark:hover:text-cinnabar-400',
			parentActive: 'font-bold text-cinnabar-700 dark:text-cinnabar-400',
			childHover: 'hover:translate-x-0.5 hover:text-cinnabar-500',
			childActive: 'font-bold text-cinnabar-600 dark:text-cinnabar-300',
			sublistBorder: 'border-ink-200 dark:border-ink-800/30'
		},
		ink: {
			parentHover: 'hover:translate-x-0.5 hover:text-ink-800 dark:hover:text-ink-200',
			parentActive: 'font-bold text-ink-900 dark:text-ink-100',
			childHover: 'hover:translate-x-0.5 hover:text-ink-700 dark:hover:text-ink-300',
			childActive: 'font-bold text-ink-800 dark:text-ink-200',
			sublistBorder: 'border-ink-100 dark:border-ink-800/30'
		}
	};

	let {
		toc,
		contentRoot,
		activeAnchor,
		onAnchorChange,
		tone = 'jade',
		size = 'sm'
	}: Props = $props();

	const toneClasses = $derived(toneClassMap[tone]);

	const handleAnchorClick = (event: MouseEvent, anchor: string) => {
		scrollToAnchor(contentRoot, anchor, event);
		onAnchorChange?.(anchor);
	};

	function isNodeOrDescendantActive(node: TOCNode): boolean {
		if (!activeAnchor) return false;
		if (node.anchor === activeAnchor) return true;
		if (node.children) {
			return node.children.some((child) => isNodeOrDescendantActive(child));
		}
		return false;
	}
</script>

{#snippet tocList(nodes: TOCNode[], depth: number)}
	<ul
		class={depth === 0
			? 'space-y-3 font-sans'
			: `mt-2 space-y-1.5 border-l pl-3 ${toneClasses.sublistBorder}`}
	>
		{#each nodes as item (item.anchor)}
			{@const isActive = activeAnchor === item.anchor}
			{@const isExpanded = isNodeOrDescendantActive(item)}
			<li class={depth === 0 ? 'space-y-2' : ''}>
				<a
					class={`block transition-all ${
						depth === 0
							? size === 'md'
								? 'text-[14px] leading-6'
								: 'text-[12px]'
							: size === 'md'
								? 'text-[13px] leading-6'
								: 'text-[11px]'
					} ${depth === 0 ? toneClasses.parentHover : toneClasses.childHover} ${
						isActive
							? depth === 0
								? toneClasses.parentActive
								: toneClasses.childActive
							: depth === 0
								? 'text-ink-500'
								: 'text-ink-400'
					}`}
					href={'#' + item.anchor}
					onclick={(event) => handleAnchorClick(event, item.anchor)}
				>
					{item.name}
				</a>
				{#if item.children?.length && isExpanded}
					<div transition:fly={{ y: -4, duration: 220 }}>
						{@render tocList(item.children, depth + 1)}
					</div>
				{/if}
			</li>
		{/each}
	</ul>
{/snippet}

<div class="custom-scrollbar max-h-[calc(100vh-20rem)] overflow-y-auto pr-2 scroll-smooth">
	{@render tocList(toc, 0)}
</div>

<style>
	.custom-scrollbar::-webkit-scrollbar {
		width: 4px;
	}
	.custom-scrollbar::-webkit-scrollbar-track {
		background: transparent;
	}
	.custom-scrollbar::-webkit-scrollbar-thumb {
		background: var(--color-ink-200);
		border-radius: 10px;
	}
	:global(.dark) .custom-scrollbar::-webkit-scrollbar-thumb {
		background: var(--color-ink-800);
	}
</style>
