<script lang="ts">
	import { parseMarkdown, SvmdChildren } from 'svmarkdown';
	import {
		safeMarkdownComponents,
		safeMarkdownParseOptions,
		safeMarkdownRenderOptions
	} from './svmarkdown';

	const { content = '' } = $props<{ content?: string }>();

	const nodes = $derived.by(() => {
		return parseMarkdown(content ?? '', safeMarkdownParseOptions).children;
	});
</script>

<div class="markdown-preview w-full break-words [overflow-wrap:anywhere]">
	<SvmdChildren {nodes} components={safeMarkdownComponents} renderOptions={safeMarkdownRenderOptions} />
</div>
