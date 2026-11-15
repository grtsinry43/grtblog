<script lang="ts">
	import { fade } from 'svelte/transition';
	import { Copy, TextQuote } from 'lucide-svelte';
	import { toast } from 'svelte-sonner';
	import { portal } from '$lib/shared/actions/portal';
	import { textSelection, type TextSelectionInfo } from '$lib/shared/actions/text-selection';
	import { commentQuote } from '$lib/features/comment/quote.svelte';

	interface Props {
		contentRoot: HTMLElement | null;
		/** Used in the "copy with source" attribution. */
		title: string;
		/** Show "引用评论" only when the page has a comment area. */
		allowQuote?: boolean;
	}

	let { contentRoot, title, allowQuote = true }: Props = $props();

	const TOOLBAR_GAP = 10;
	const VIEWPORT_PAD = 8;

	let selection = $state<TextSelectionInfo | null>(null);
	let toolbarWidth = $state(0);
	let toolbarHeight = $state(0);

	$effect(() => {
		if (!contentRoot) return;
		const handle = textSelection(contentRoot, {
			onChange: (next) => (selection = next)
		});
		return () => handle.destroy();
	});

	// Mouse: above the selection. Touch: below it, since the OS selection menu sits above.
	const position = $derived.by(() => {
		if (!selection) return null;
		const { rect, pointerType } = selection;
		const half = toolbarWidth / 2;
		const centerX = Math.min(
			Math.max(rect.left + rect.width / 2, VIEWPORT_PAD + half),
			window.innerWidth - VIEWPORT_PAD - half
		);
		const preferBelow = pointerType === 'touch';
		const aboveTop = rect.top - TOOLBAR_GAP - toolbarHeight;
		const below = preferBelow || aboveTop < 64;
		return {
			left: centerX,
			top: below ? rect.bottom + TOOLBAR_GAP : aboveTop,
			below,
			// Keep the arrow pointing at the selection even when the toolbar is clamped.
			arrowX: rect.left + rect.width / 2 - (centerX - half)
		};
	});

	const clearSelection = () => {
		document.getSelection()?.removeAllRanges();
		selection = null;
	};

	function quoteToComment() {
		if (!selection) return;
		commentQuote.request(selection.text);
		clearSelection();
		document
			.querySelector('[data-comment-area]')
			?.scrollIntoView({ behavior: 'smooth', block: 'start' });
	}

	async function copyWithSource() {
		if (!selection) return;
		const url = window.location.href.split('#')[0];
		const text = `「${selection.text}」\n—— 《${title}》 ${url}`;
		try {
			await navigator.clipboard.writeText(text);
			toast.success('已复制，附带出处');
		} catch {
			toast.error('复制失败，请手动复制');
		}
		clearSelection();
	}

	const actionClass =
		'inline-flex h-8 items-center gap-1.5 rounded-[2px] px-2.5 text-[12px] text-ink-100 transition-colors hover:bg-white/10 hover:text-white dark:text-ink-800 dark:hover:bg-ink-900/10 dark:hover:text-ink-950';
</script>

{#if selection && position}
	<div
		use:portal
		bind:clientWidth={toolbarWidth}
		bind:clientHeight={toolbarHeight}
		role="toolbar"
		tabindex="-1"
		aria-label="选中文字操作"
		class="fixed z-(--z-popover) -translate-x-1/2"
		style:left="{position.left}px"
		style:top="{position.top}px"
		style:visibility={toolbarWidth ? 'visible' : 'hidden'}
		transition:fade={{ duration: 120 }}
		onpointerdown={(e) => e.preventDefault()}
	>
		<div
			class="relative flex items-center gap-0.5 rounded-default bg-ink-900/95 p-0.5 font-sans shadow-deep backdrop-blur-md dark:bg-ink-100/95"
		>
			{#if allowQuote}
				<button type="button" class={actionClass} onclick={quoteToComment}>
					<TextQuote size={14} />
					引用评论
				</button>
				<span aria-hidden="true" class="h-4 w-px bg-white/15 dark:bg-ink-900/15"></span>
			{/if}
			<button type="button" class={actionClass} onclick={copyWithSource}>
				<Copy size={13} />
				复制出处
			</button>
			<span
				aria-hidden="true"
				class="absolute h-2 w-2 -translate-x-1/2 rotate-45 bg-ink-900/95 dark:bg-ink-100/95 {position.below
					? '-top-1'
					: '-bottom-1'}"
				style:left="{Math.min(Math.max(position.arrowX, 12), toolbarWidth - 12)}px"
			></span>
		</div>
	</div>
{/if}
