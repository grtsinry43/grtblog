<script lang="ts">
	import { Separator } from 'bits-ui';

	interface Props {
		vertical?: boolean;
		class?: string;
		label?: string;
	}

	let { vertical = false, class: className = '', label = '' }: Props = $props();

	const cx = (...parts: Array<string | false | null | undefined>) =>
		parts.filter(Boolean).join(' ');

	let containerClasses = $derived(
		cx('flex items-center gap-4', vertical ? 'h-full flex-col' : 'w-full', className)
	);
	let lineClasses = $derived(
		cx(
			'bg-ink-100 transition-colors dark:bg-ink-800/50 shrink-0',
			vertical ? 'w-px flex-1' : 'h-px flex-1'
		)
	);
</script>

<div class={containerClasses}>
	<Separator.Root
		orientation={vertical ? 'vertical' : 'horizontal'}
		class={lineClasses}
		decorative
	/>
	{#if label}
		<span class="whitespace-nowrap font-mono text-[9px] tracking-[0.3em] text-ink-400 uppercase">
			{label}
		</span>
		<Separator.Root
			orientation={vertical ? 'vertical' : 'horizontal'}
			class={lineClasses}
			decorative
		/>
	{/if}
</div>
