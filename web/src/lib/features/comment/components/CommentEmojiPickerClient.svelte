<script lang="ts">
	import { onMount, tick } from 'svelte';
	import { Smile, X } from 'lucide-svelte';
	import { portal } from '$lib/shared/actions/portal';

	type Props = {
		onPick?: (emoji: string) => void;
	};

	let { onPick = () => {} }: Props = $props();

	let open = $state(false);
	let ready = $state(false);
	let pickerEl = $state<HTMLElement | null>(null);
	let wrapperEl = $state<HTMLElement | null>(null);
	let triggerEl = $state<HTMLButtonElement | null>(null);
	let panelEl = $state<HTMLDivElement | null>(null);
	let panelStyle = $state('');

	async function handleToggle() {
		if (!ready) return;
		open = !open;
		if (open) {
			await tick();
			updatePanelPosition();
		}
	}

	function handleClose() {
		open = false;
	}

	function handleOutsidePointerDown(event: PointerEvent) {
		if (!open || !wrapperEl) return;
		const target = event.target;
		if (!(target instanceof Node)) return;
		if (!wrapperEl.contains(target) && !(panelEl && panelEl.contains(target))) {
			open = false;
		}
	}

	function handleEmojiClick(event: Event) {
		const detail = (event as CustomEvent<{ unicode?: string }>).detail;
		const unicode = detail?.unicode;
		if (!unicode) return;
		onPick(unicode);
		open = false;
	}

	function updatePanelPosition() {
		if (!open || !triggerEl || typeof window === 'undefined') return;
		const rect = triggerEl.getBoundingClientRect();
		const margin = 12;
		const gap = 8;
		const panelWidth = 336;
		const panelHeight = panelEl?.offsetHeight ?? 430;

		let left = rect.left;
		left = Math.max(margin, Math.min(left, window.innerWidth - panelWidth - margin));

		let top = rect.top - panelHeight - gap;
		if (top < margin) {
			top = rect.bottom + gap;
		}
		top = Math.max(margin, Math.min(top, window.innerHeight - panelHeight - margin));

		panelStyle = `left:${Math.round(left)}px;top:${Math.round(top)}px;`;
	}

	onMount(async () => {
		await import('emoji-picker-element');
		ready = true;
	});

	$effect(() => {
		if (typeof window === 'undefined') return;
		window.addEventListener('pointerdown', handleOutsidePointerDown, true);
		window.addEventListener('resize', updatePanelPosition);
		window.addEventListener('scroll', updatePanelPosition, true);
		return () => {
			window.removeEventListener('pointerdown', handleOutsidePointerDown, true);
			window.removeEventListener('resize', updatePanelPosition);
			window.removeEventListener('scroll', updatePanelPosition, true);
		};
	});

	$effect(() => {
		if (!pickerEl || !ready) return;
		pickerEl.addEventListener('emoji-click', handleEmojiClick as EventListener);
		return () => {
			pickerEl?.removeEventListener('emoji-click', handleEmojiClick as EventListener);
		};
	});

	$effect(() => {
		if (!open) return;
		updatePanelPosition();
	});
</script>

<div class="relative" bind:this={wrapperEl}>
	<button
		type="button"
		bind:this={triggerEl}
		onclick={handleToggle}
		class="inline-flex items-center gap-1.5 rounded-default border border-ink-200/70 dark:border-ink-700/70 bg-ink-50/70 dark:bg-ink-900/45 px-2.5 py-1 text-[11px] text-ink-500 dark:text-ink-300 hover:text-ink-700 dark:hover:text-ink-100 hover:border-ink-300/80 dark:hover:border-ink-600/80 transition-colors"
	>
		<Smile size={14} />
		<span>表情</span>
	</button>

	{#if open}
		<div
			bind:this={panelEl}
			use:portal
			style={panelStyle}
			class="fixed z-[1200] rounded-default border border-ink-200/70 dark:border-ink-700/70 bg-white/95 dark:bg-ink-900/95 backdrop-blur-md shadow-xl p-2"
		>
			<div class="mb-1 flex items-center justify-between px-1">
				<span class="text-[10px] font-mono text-ink-400 dark:text-ink-500 uppercase">Emoji</span>
				<button
					type="button"
					onclick={handleClose}
					class="inline-flex h-5 w-5 items-center justify-center rounded-full text-ink-400 hover:text-ink-700 dark:hover:text-ink-200 hover:bg-ink-100/80 dark:hover:bg-ink-800/70 transition-colors"
				>
					<X size={12} />
				</button>
			</div>

			{#if ready}
				<emoji-picker bind:this={pickerEl} class="comment-emoji-picker"></emoji-picker>
			{:else}
				<div class="w-[320px] h-[380px] flex items-center justify-center text-xs text-ink-400">
					加载表情面板中...
				</div>
			{/if}
		</div>
	{/if}
</div>

<style lang="postcss">
	@reference "$routes/layout.css";

	:global(.comment-emoji-picker) {
		--border-radius: 10px;
		--num-columns: 8;
		--emoji-size: 1.18rem;
		--emoji-padding: 0.38rem;
		--background: transparent;
		--border-color: transparent;
		--input-border-color: rgb(229 231 235 / 0.85);
		--input-font-color: rgb(71 85 105 / 0.95);
		--button-hover-background: rgb(148 163 184 / 0.16);
		--button-active-background: rgb(100 116 139 / 0.18);
		width: 320px;
		height: 380px;
	}

	:global(.dark .comment-emoji-picker) {
		--input-border-color: rgb(71 85 105 / 0.75);
		--input-font-color: rgb(203 213 225 / 0.95);
		--button-hover-background: rgb(71 85 105 / 0.28);
		--button-active-background: rgb(51 65 85 / 0.35);
	}
</style>
