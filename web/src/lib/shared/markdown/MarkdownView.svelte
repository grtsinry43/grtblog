<script lang="ts">
	import { parseMarkdown, SvmdChildren } from 'svmarkdown';
	import type { SvmdComponentMap, SvmdNode, SvmdParseOptions, SvmdRenderOptions } from 'svmarkdown';
	import { markdownComponents, markdownParseOptions, markdownRenderOptions } from './svmarkdown';

	const {
		content = '',
		headingAnchors = [],
		components = markdownComponents,
		parseOptions = markdownParseOptions,
		renderOptions = markdownRenderOptions
	} = $props<{
		content?: string;
		headingAnchors?: string[];
		components?: SvmdComponentMap;
		parseOptions?: SvmdParseOptions;
		renderOptions?: SvmdRenderOptions;
	}>();

	const applyHeadingAnchors = (nodes: SvmdNode[], anchors: string[]) => {
		if (!anchors.length) return nodes;
		let index = 0;
		const walk = (items: SvmdNode[]) => {
			for (const node of items) {
				if (node.kind === 'element' && /^h[1-6]$/.test(node.name)) {
					const anchor = anchors[index];
					if (anchor) {
						node.attrs = { ...node.attrs, id: node.attrs?.id ?? anchor };
					}
					index += 1;
				}
				if ('children' in node && node.children?.length) {
					walk(node.children);
				}
			}
		};
		walk(nodes);
		return nodes;
	};

	const nodes = $derived.by(() => {
		const ast = parseMarkdown(content ?? '', parseOptions);
		return applyHeadingAnchors(ast.children, headingAnchors);
	});
</script>

<div class="markdown-preview w-full break-words [overflow-wrap:anywhere]">
	<SvmdChildren {nodes} {components} {renderOptions} />
</div>
