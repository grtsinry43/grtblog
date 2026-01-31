<script lang="ts">
	import { websiteInfoCtx } from '$lib/features/website-info/context';
	import { getSiteIconUrl, resolveLinkSite } from '$lib/shared/markdown/link-icons';

	const { href = '', title = '', contentHtml = '' } = $props<{
		href?: string;
		title?: string;
		contentHtml?: string;
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
</script>

<a class="md-link" data-site={site || undefined} {href} {title} {rel} {target}>
	{@html contentHtml}
	<span
		class="md-link__icon"
		aria-hidden="true"
		style={iconUrl ? `--md-link-icon-url: url("${iconUrl}")` : undefined}
	></span>
</a>
