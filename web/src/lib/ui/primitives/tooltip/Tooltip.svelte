<script lang="ts">
	import { Tooltip as BitsTooltip } from 'bits-ui';
	import type { Snippet } from 'svelte';

	let {
		children,
		content,
		delayDuration = 200,
		side = 'top',
		align = 'center'
	} = $props<{
		children: Snippet;
		content: string | Snippet;
		delayDuration?: number;
		side?: 'top' | 'right' | 'bottom' | 'left';
		align?: 'start' | 'center' | 'end';
	}>();
</script>

<BitsTooltip.Provider {delayDuration}>
	<BitsTooltip.Root>
		<BitsTooltip.Trigger class="inline-flex cursor-default">
			{@render children()}
		</BitsTooltip.Trigger>
		<BitsTooltip.Portal>
			<BitsTooltip.Content
				{side}
				{align}
				sideOffset={4}
				class="z-50 overflow-hidden rounded-default bg-white dark:bg-ink-800 px-3 py-1.5 text-xs text-ink-900 dark:text-jade-100 shadow-lg animate-in fade-in zoom-in-95 data-[side=bottom]:slide-in-from-top-2 data-[side=left]:slide-in-from-right-2 data-[side=right]:slide-in-from-left-2 data-[side=top]:slide-in-from-bottom-2 border border-ink-100 dark:border-ink-700 font-serif"
			>
				{#if typeof content === 'string'}
					{content}
				{:else}
					{@render content()}
				{/if}
				<BitsTooltip.Arrow
					class="fill-white dark:fill-ink-800 stroke-ink-100 dark:stroke-ink-700"
				/>
			</BitsTooltip.Content>
		</BitsTooltip.Portal>
	</BitsTooltip.Root>
</BitsTooltip.Provider>
