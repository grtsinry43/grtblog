<script lang="ts">
	import Button from '$lib/ui/primitives/button/Button.svelte';
	import { ArrowLeft, ArrowRight } from 'lucide-svelte';

	interface Props {
		current: number;
		total: number;
		onPageChange: (page: number) => void;
		class?: string;
	}

	let { current, total, onPageChange, class: className = '' }: Props = $props();

	const pages = $derived(Array.from({ length: total }, (_, i) => i + 1));
</script>

{#snippet prevContent()}
	<ArrowLeft size={12} class="mr-1.5" />
	上一页
{/snippet}

{#snippet nextContent()}
	下一页
	<ArrowRight size={12} class="ml-1.5" />
{/snippet}

<nav class={`flex items-center justify-center gap-3 ${className}`}>
	<Button
		variant="ghost"
		disabled={current <= 1}
		onclick={() => onPageChange(current - 1)}
		class="h-8 !bg-transparent px-2 font-mono text-[10px] tracking-widest uppercase hover:!text-jade-600"
		content={prevContent}
	/>

	<div class="flex items-center gap-1.5">
		{#each pages as page}
			{#if page === current}
				<span
					class="flex h-7 w-7 items-center justify-center rounded-default bg-jade-800 font-mono text-[11px] font-bold text-ink-200 dark:text-ink-800 shadow-sm"
				>
					{page}
				</span>
			{:else if page === 1 || page === total || (page >= current - 1 && page <= current + 1)}
				<Button
					variant="ghost"
					onclick={() => onPageChange(page)}
					class="h-7 w-7 rounded-default font-mono text-[11px] text-ink-800 hover:text-ink-900"
					>{page}</Button
				>
			{:else if (page === current - 2 && page > 1) || (page === current + 2 && page < total)}
				<span class="select-none px-0.5 font-mono text-[9px] text-ink-300">...</span>
			{/if}
		{/each}
	</div>

	<Button
		variant="ghost"
		disabled={current >= total}
		onclick={() => onPageChange(current + 1)}
		class="h-8 !bg-transparent px-2 font-mono text-[10px] tracking-widest uppercase hover:!text-jade-600"
		content={nextContent}
	/>
</nav>
