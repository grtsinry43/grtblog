<script lang="ts">
	import { scrollToAnchor } from '$lib/shared/dom/scroll-to-anchor';
	import type { TOCNode } from '$lib/shared/types/toc';

	type TocTone = 'jade' | 'cinnabar' | 'ink';

	interface Props {
		toc: TOCNode[];
		contentRoot: HTMLElement | null;
		activeAnchor: string | null;
		onAnchorChange?: (anchor: string) => void;
		tone?: TocTone;
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

	let { toc, contentRoot, activeAnchor, onAnchorChange, tone = 'jade' }: Props = $props();

	const toneClasses = $derived(toneClassMap[tone]);

	const handleAnchorClick = (event: MouseEvent, anchor: string) => {
		scrollToAnchor(contentRoot, anchor, event);
		onAnchorChange?.(anchor);
	};
</script>

<ul class="space-y-3 font-sans">
	{#each toc as item}
		<li class="space-y-2">
			<a
				class={`block text-[12px] text-ink-500 transition-all ${toneClasses.parentHover} ${
					activeAnchor === item.anchor ? toneClasses.parentActive : ''
				}`}
				href={'#' + item.anchor}
				onclick={(event) => handleAnchorClick(event, item.anchor)}
			>
				{item.name}
			</a>
			{#if item.children?.length}
				<ul class={`space-y-1.5 border-l pl-3 ${toneClasses.sublistBorder}`}>
					{#each item.children as child}
						<li>
							<a
								class={`block text-[11px] text-ink-400 transition-all ${toneClasses.childHover} ${
									activeAnchor === child.anchor ? toneClasses.childActive : ''
								}`}
								href={'#' + child.anchor}
								onclick={(event) => handleAnchorClick(event, child.anchor)}
							>
								{child.name}
							</a>
						</li>
					{/each}
				</ul>
			{/if}
		</li>
	{/each}
</ul>
