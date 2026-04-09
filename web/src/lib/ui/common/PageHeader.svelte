<script lang="ts">
	import { FadeIn } from '$lib/ui/animation';

	interface Props {
		title: string;
		subtitle?: string;
		description?: string;
		tag?: string;
		className?: string;
	}

	let { title, subtitle, description, tag, className = '' }: Props = $props();
</script>

<header class="relative mb-12 sm:mb-16 {className} px-4">
	<FadeIn y={10} duration={800}>
		<div class="flex flex-col items-center">
			<!-- Title Section -->
			<div class="relative flex flex-col items-center group">
				{#if tag}
					<span
						class="absolute -top-4 -right-8 [writing-mode:vertical-rl] text-[9px] font-serif font-bold text-jade-600/60 dark:text-jade-400/50 tracking-[0.3em] uppercase border-r border-jade-500/20 pr-1 h-10"
					>
						{tag}
					</span>
				{/if}

				<h1
					class="text-2xl md:text-3xl font-serif font-bold text-ink-900 dark:text-ink-50 tracking-[0.1em]"
				>
					{title}
				</h1>

				<!-- Subtle Underline Decor -->
				<div class="flex items-center gap-2 mt-3">
					<div class="w-1 h-1 rounded-full bg-jade-500/40"></div>
					<div class="w-12 h-px bg-jade-500/20"></div>
					<div class="w-1 h-1 rounded-full bg-jade-500/40"></div>
				</div>
			</div>

			{#if subtitle}
				<p
					class="mt-6 font-serif italic text-base text-ink-700 dark:text-ink-200 opacity-90 tracking-wide"
				>
					{subtitle}
				</p>
			{/if}

			{#if description}
				<div class="mt-4 flex items-start gap-3 max-w-lg">
					<span class="text-jade-500/40 font-serif text-lg leading-none">“</span>
					<p
						class="text-[13px] leading-relaxed text-ink-500 dark:text-ink-400 font-serif text-center italic"
					>
						{description}
					</p>
					<span class="text-jade-500/40 font-serif text-lg leading-none self-end">”</span>
				</div>
			{/if}
		</div>
	</FadeIn>
</header>
