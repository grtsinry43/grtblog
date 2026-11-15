<script lang="ts">
	import { FadeIn } from '$lib/ui/animation';

	interface Props {
		title: string;
		subtitle?: string;
		description?: string;
		tag?: string;
		/** Right side of the eyebrow rule, e.g. "128 篇". Prefer real data over decoration. */
		meta?: string;
		className?: string;
	}

	let { title, subtitle, description, tag, meta, className = '' }: Props = $props();

	// Subtitle and description read as one paragraph; close the subtitle sentence if needed.
	const subtitleText = $derived.by(() => {
		if (!subtitle) return '';
		if (!description) return subtitle;
		return /[。！？.!?」”…]$/.test(subtitle) ? subtitle : `${subtitle}。`;
	});
</script>

<header class="relative mb-12 sm:mb-14 {className}">
	<FadeIn y={10} duration={800}>
		{#if tag || meta}
			<div class="flex items-center gap-3 font-mono text-micro">
				{#if tag}
					<span class="tracking-[0.2em] text-jade-700 uppercase dark:text-jade-400">{tag}</span>
				{/if}
				<span
					aria-hidden="true"
					class="h-px flex-1 bg-gradient-to-r from-ink-200 to-ink-200/30 dark:from-ink-800 dark:to-ink-800/30"
				></span>
				{#if meta}
					<span class="tabular-nums text-ink-500 dark:text-ink-400">{meta}</span>
				{/if}
			</div>
		{/if}

		<h1
			class="mt-5 font-serif text-3xl font-bold tracking-wide text-balance text-ink-900 md:text-4xl dark:text-ink-50"
		>
			{title}
		</h1>

		{#if subtitleText || description}
			<p
				class="mt-4 max-w-2xl font-serif text-[15px] leading-[1.85] text-pretty text-ink-500 dark:text-ink-400"
			>
				{#if subtitleText}<span class="text-ink-800 dark:text-ink-200">{subtitleText}</span
					>{/if}{description ?? ''}
			</p>
		{/if}
	</FadeIn>
</header>
