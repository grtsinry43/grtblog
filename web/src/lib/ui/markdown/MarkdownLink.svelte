<script lang="ts">
	/* eslint-disable svelte/no-navigation-without-resolve */
	import { resolve } from '$app/paths';
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

	const standaloneFavicon = $derived.by(() => {
		if (iconUrl) return iconUrl;
		try {
			const url = new URL(href);
			return `https://www.google.com/s2/favicons?domain=${url.hostname}&sz=32`;
		} catch {
			return '';
		}
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
		class={`group relative mx-auto my-4 flex max-w-[400px] items-center justify-between gap-3 overflow-hidden rounded-default border border-ink-200/80 bg-white/80 px-4 py-3 shadow-subtle transition-all hover:-translate-y-0.5 hover:shadow-float dark:border-ink-800/60 dark:bg-ink-900/40 no-underline ${className}`.trim()}
		data-site={site || undefined}
		href={href && !/^(https?:|mailto:|tel:|#|\/\/)/i.test(href) ? resolve(href) : href}
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

		<span class="relative z-10 min-w-0 flex-1 overflow-hidden">
			<span
				class="block truncate text-[13px] font-semibold text-ink-900 transition-colors group-hover:text-jade-700 dark:text-ink-100 dark:group-hover:text-jade-300"
			>
				{@render children?.()}
			</span>
			<span class="mt-0.5 block truncate text-[11px] text-ink-500 dark:text-ink-400">
				{href}
			</span>
		</span>
		{#if standaloneFavicon}
			<img
				class="relative z-10 h-5 w-5 shrink-0 rounded opacity-70 transition-opacity group-hover:opacity-100"
				src={standaloneFavicon}
				alt=""
				loading="lazy"
			/>
		{:else}
			<svg class="relative z-10 h-4 w-4 shrink-0 text-ink-400 dark:text-ink-500" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
				<path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71" />
				<path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71" />
			</svg>
		{/if}
	</a>
{:else}
	<a
		class={`md-link relative inline-flex max-w-full flex-wrap items-center gap-x-[0.35em] gap-y-[0.15em] break-words [overflow-wrap:anywhere] no-underline ${className}`.trim()}
		data-site={site || undefined}
		href={href && !/^(https?:|mailto:|tel:|#|\/\/)/i.test(href) ? resolve(href) : href}
		{title}
		{rel}
		{target}
	>
		<span class="relative z-[1] min-w-0 break-words [overflow-wrap:anywhere]">{@render children?.()}</span>
		{#if site}
			<span
				class="md-link__icon relative z-[1] inline-block h-[0.9em] w-[0.9em] shrink-0 rounded bg-cover bg-center bg-no-repeat opacity-75"
				aria-hidden="true"
				style={iconStyle}
			></span>
		{/if}
	</a>
{/if}

<style>
	.md-link {
		color: var(--color-ink-900);
		transition: color 200ms cubic-bezier(0, 0.8, 0.13, 1),
			transform 150ms ease;
		cursor: pointer;
	}

	:global(.dark) .md-link {
		color: var(--color-ink-100);
	}

	.md-link::after {
		content: '';
		position: absolute;
		left: 0;
		right: 0;
		bottom: 0;
		top: 70%;
		z-index: 0;
		background: color-mix(in srgb, var(--color-jade-300) 45%, transparent);
		border-radius: 2px;
		transition: top 200ms cubic-bezier(0, 0.8, 0.13, 1);
	}

	:global(.dark) .md-link::after {
		background: color-mix(in srgb, var(--color-jade-700) 40%, transparent);
	}

	.md-link:hover {
		color: var(--color-jade-900);
		transform: translateY(-0.5px);
	}

	:global(.dark) .md-link:hover {
		color: var(--color-jade-100);
	}

	.md-link:hover::after {
		top: 0%;
	}
</style>
