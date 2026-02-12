<script lang="ts">
	import type { Snippet } from 'svelte';
	import { websiteInfoCtx } from '$lib/features/website-info/context';
	import { getSiteIconUrl, resolveLinkSite } from '$lib/shared/markdown/link-icons';

	const {
		href = '',
		title = '',
		children,
		class: className = '',
		linkLayout = 'inline',
		linkStandalone
	} = $props<{
		href?: string;
		title?: string;
		children?: Snippet;
		class?: string;
		linkLayout?: 'inline' | 'standalone';
		linkStandalone?: boolean;
	}>();

	const siteFavicon = websiteInfoCtx.selectModelData((data) => data?.favicon || '');
	let site = $derived(
		resolveLinkSite(href, typeof window !== 'undefined' ? window.location.origin : undefined)
	);
	const isExternal = $derived(!site && (/^https?:\/\//i.test(href) || href.startsWith('//')));
	const rel = $derived(isExternal ? 'noopener noreferrer' : undefined);
	const target = $derived(isExternal ? '_blank' : undefined);
	const standalone = $derived(linkStandalone ?? linkLayout === 'standalone');
	let iconUrl = $derived(getSiteIconUrl(site, $siteFavicon));
	const iconStyle = $derived.by(() => {
		if (!site || !iconUrl) return undefined;
		if (site === 'internal') {
			return `background-image: url("${iconUrl}")`;
		}
		return [
			`background-color: currentColor`,
			`mask-image: url("${iconUrl}")`,
			`mask-size: cover`,
			`mask-position: center`,
			`-webkit-mask-image: url("${iconUrl}")`,
			`-webkit-mask-size: cover`,
			`-webkit-mask-position: center`
		].join('; ');
	});

	let mouseX = $state(0);
	let mouseY = $state(0);
	let spotlightOpacity = $state(0);

	const handleMouseMove = (e: MouseEvent) => {
		const rect = (e.currentTarget as HTMLElement).getBoundingClientRect();
		mouseX = e.clientX - rect.left;
		mouseY = e.clientY - rect.top;
		spotlightOpacity = 1;
	};

	const handleMouseLeave = () => {
		spotlightOpacity = 0;
	};
</script>

{#if standalone}
	<a
		class={`group relative my-4 flex items-center justify-between gap-4 overflow-hidden rounded-default border border-ink-200/80 bg-white/80 px-5 py-4 shadow-subtle transition-all hover:-translate-y-0.5 hover:shadow-float dark:border-ink-800/60 dark:bg-ink-900/40 no-underline ${className}`.trim()}
		data-site={site || undefined}
		{href}
		{title}
		{rel}
		{target}
		onmousemove={handleMouseMove}
		onmouseleave={handleMouseLeave}
	>
		<span
			class="pointer-events-none absolute inset-0 z-0 transition-opacity duration-300"
			style:opacity={spotlightOpacity}
			style:background={`radial-gradient(600px circle at ${mouseX}px ${mouseY}px, color-mix(in srgb, var(--color-jade-500), transparent 70%) 0%, transparent 40%)`}
			style:mix-blend-mode="soft-light"
		></span>

		<span
			class="absolute inset-0 z-0 opacity-0 transition-opacity duration-300 group-hover:opacity-10 dark:group-hover:opacity-20 bg-jade-500/10"
		></span>

		<span class="relative z-10 min-w-0 flex-1">
			<span class="block truncate text-base font-semibold text-ink-900 transition-colors group-hover:text-jade-700 dark:text-ink-100 dark:group-hover:text-jade-300">
				{@render children?.()}
			</span>
			<span class="mt-1 block line-clamp-2 text-sm leading-relaxed text-ink-600 dark:text-ink-400">
				{href}
			</span>
		</span>
		<span
			class="relative z-10 shrink-0 rounded-full border border-ink-200/80 bg-white/70 px-2.5 py-0.5 text-[11px] uppercase tracking-[0.18em] text-ink-500 dark:border-ink-700/60 dark:bg-ink-900/60 dark:text-ink-300"
		>
			Link
		</span>
	</a>
{:else}
	<a
		class={`md-link inline-flex items-center gap-[0.35em] underline decoration-1 underline-offset-2 ${className}`.trim()}
		data-site={site || undefined}
		{href}
		{title}
		{rel}
		{target}
	>
		<span>{@render children?.()}</span>
		<span
			class={`md-link__icon inline-block rounded opacity-75 bg-center bg-no-repeat bg-cover ${site ? 'h-[0.9em] w-[0.9em]' : 'h-0 w-0'}`.trim()}
			aria-hidden="true"
			style={iconStyle}
		></span>
	</a>
{/if}
