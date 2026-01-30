<script lang="ts">
	import type { MouseEventHandler } from 'svelte/elements';

	const {
		href = '',
		title = '',
		desc = '',
		className
	} = $props<{
		href?: string;
		title?: string;
		desc?: string;
		className?: string;
	}>();

	const isExternal = $derived(/^https?:\/\//i.test(href) || href.startsWith('//'));
	const target = $derived(isExternal ? '_blank' : '_self');
	const rel = $derived(isExternal ? 'noreferrer' : undefined);

	const cn = (...classes: (string | undefined | null | false)[]) =>
		classes.filter(Boolean).join(' ');

	let mouseX = $state(0);
	let mouseY = $state(0);
	let spotlightOpacity = $state(0);

	const handleMouseMove: MouseEventHandler<HTMLAnchorElement> = (e) => {
		const rect = e.currentTarget.getBoundingClientRect();
		mouseX = e.clientX - rect.left;
		mouseY = e.clientY - rect.top;
		spotlightOpacity = 1;
	};

	const handleMouseLeave: MouseEventHandler<HTMLAnchorElement> = () => {
		spotlightOpacity = 0;
	};
</script>

<a
	class={cn(
		'group relative max-w-96 flex items-center justify-between gap-4 overflow-hidden rounded-default border border-ink-200/80 bg-white/80 px-5 py-4 shadow-subtle transition-all hover:-translate-y-0.5 hover:shadow-float dark:border-ink-800/60 dark:bg-ink-900/40 no-underline',
		className
	)}
	href={href || '#'}
	{target}
	{rel}
	onmousemove={handleMouseMove}
	onmouseleave={handleMouseLeave}
>
	<!-- Spotlight Effect -->
	<div
		class="pointer-events-none absolute inset-0 z-0 transition-opacity duration-300"
		style:opacity={spotlightOpacity}
		style:background={`radial-gradient(600px circle at ${mouseX}px ${mouseY}px, color-mix(in srgb, var(--color-jade-500), transparent 70%) 0%, transparent 40%)`}
		style:mix-blend-mode="soft-light"
	></div>

	<!-- Background tint for dark mode/hover -->
	<div
		class="absolute inset-0 z-0 opacity-0 transition-opacity duration-300 group-hover:opacity-10 dark:group-hover:opacity-20 bg-jade-500/10"
	></div>

	<div class="relative z-10 min-w-0 flex-1">
		<div
			class="truncate text-base font-semibold text-ink-900 transition-colors group-hover:text-jade-700 dark:text-ink-100 dark:group-hover:text-jade-300"
		>
			{title || href}
		</div>
		{#if desc}
			<div class="mt-1 line-clamp-2 text-sm leading-relaxed text-ink-600 dark:text-ink-400">
				{desc}
			</div>
		{/if}
	</div>
	<span
		class="relative z-10 shrink-0 rounded-full border border-ink-200/80 bg-white/70 px-2.5 py-0.5 text-[11px] uppercase tracking-[0.18em] text-ink-500 dark:border-ink-700/60 dark:bg-ink-900/60 dark:text-ink-300"
	>
		Link
	</span>
</a>
