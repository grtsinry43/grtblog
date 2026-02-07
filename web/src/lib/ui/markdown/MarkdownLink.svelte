<script lang="ts">
	import { websiteInfoCtx } from '$lib/features/website-info/context';
	import { getSiteIconUrl, resolveLinkSite } from '$lib/shared/markdown/link-icons';

	const { href = '', title = '', class: className = '' } = $props<{
		href?: string;
		title?: string;
		class?: string;
	}>();

	const siteFavicon = websiteInfoCtx.selectModelData((data) => data?.favicon || '');
	let site = $derived(
		resolveLinkSite(href, typeof window !== 'undefined' ? window.location.origin : undefined)
	);
	const isExternal = $derived(
		!site && (/^https?:\/\//i.test(href) || href.startsWith('//'))
	);
	const rel = $derived(isExternal ? 'noopener noreferrer' : undefined);
	const target = $derived(isExternal ? '_blank' : undefined);
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
</script>

<a
	class={`md-link inline-flex items-center gap-[0.35em] underline decoration-1 underline-offset-2 ${className}`.trim()}
	data-site={site || undefined}
	{href}
	{title}
	{rel}
	{target}
>
	<slot />
	<span
		class={`md-link__icon inline-block rounded opacity-75 bg-center bg-no-repeat bg-cover ${site ? 'h-[0.9em] w-[0.9em]' : 'h-0 w-0'}`.trim()}
		aria-hidden="true"
		style={iconStyle}
	></span>
</a>
