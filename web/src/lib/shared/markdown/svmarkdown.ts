import type { SvmdComponentMap, SvmdParseOptions, SvmdRenderOptions } from 'svmarkdown';
import { markdownComponents as componentDefinitions } from '$lib/shared/markdown/shared/components';
import MarkdownLink from '$lib/ui/markdown/MarkdownLink.svelte';
import MarkdownImage from '$lib/ui/markdown/MarkdownImage.svelte';
import MarkdownCodeBlock from '$lib/ui/markdown/MarkdownCodeBlock.svelte';
import MarkdownFallback from '$lib/ui/markdown/MarkdownFallback.svelte';
import MarkdownHeading from '$lib/ui/markdown/MarkdownHeading.svelte';
import MarkdownParagraph from '$lib/ui/markdown/MarkdownParagraph.svelte';
import MarkdownList from '$lib/ui/markdown/MarkdownList.svelte';
import MarkdownListItem from '$lib/ui/markdown/MarkdownListItem.svelte';
import MarkdownBlockquote from '$lib/ui/markdown/MarkdownBlockquote.svelte';
import MarkdownHr from '$lib/ui/markdown/MarkdownHr.svelte';
import MarkdownTable from '$lib/ui/markdown/MarkdownTable.svelte';
import MarkdownThead from '$lib/ui/markdown/MarkdownThead.svelte';
import MarkdownTbody from '$lib/ui/markdown/MarkdownTbody.svelte';
import MarkdownTr from '$lib/ui/markdown/MarkdownTr.svelte';
import MarkdownTh from '$lib/ui/markdown/MarkdownTh.svelte';
import MarkdownTd from '$lib/ui/markdown/MarkdownTd.svelte';
import YearCard from '$lib/ui/markdown/YearCard.svelte';
import LinkCard from '$lib/ui/markdown/LinkCard.svelte';
import FootnoteLinkCard from '$lib/ui/markdown/FootnoteLinkCard.svelte';

const componentBlocks = Object.fromEntries(
	componentDefinitions.map((component) => [component.name, true])
) satisfies SvmdParseOptions['componentBlocks'];

export const markdownComponents: SvmdComponentMap = {
	h1: MarkdownHeading,
	h2: MarkdownHeading,
	h3: MarkdownHeading,
	h4: MarkdownHeading,
	h5: MarkdownHeading,
	h6: MarkdownHeading,
	p: MarkdownParagraph,
	ul: MarkdownList,
	ol: MarkdownList,
	li: MarkdownListItem,
	blockquote: MarkdownBlockquote,
	hr: MarkdownHr,
	table: MarkdownTable,
	thead: MarkdownThead,
	tbody: MarkdownTbody,
	tr: MarkdownTr,
	th: MarkdownTh,
	td: MarkdownTd,
	a: MarkdownLink,
	img: MarkdownImage,
	code: MarkdownCodeBlock,
	gallery: MarkdownFallback,
	callout: MarkdownFallback,
	timeline: MarkdownFallback,
	'year-card': YearCard,
	'link-card': LinkCard,
	'footnote-link-card': FootnoteLinkCard
};

export const markdownParseOptions: SvmdParseOptions = {
	componentBlocks,
	markdownItPlugins: [],
	markdownItOptions: {
		html: true,
		linkify: true,
		typographer: true
	}
};

export const markdownRenderOptions: SvmdRenderOptions = {
	allowDangerousHtml: true
};
