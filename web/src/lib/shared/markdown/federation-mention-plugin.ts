import type MarkdownIt from 'markdown-it';

/**
 * markdown-it inline plugin that intercepts server-generated
 * `<fed-mention user="..." instance="..." status="...">@user@instance</fed-mention>`
 * and converts it into proper open/close token pairs so that svmarkdown
 * renders it via the component map as a Svelte component.
 */

const openRe = /^<fed-mention\s+([^>]*)>/;
const attrRe = /([a-z][\w-]*)\s*=\s*"([^"]*)"/g;

function parseAttrs(raw: string): [string, string][] {
	const attrs: [string, string][] = [];
	let m: RegExpExecArray | null;
	attrRe.lastIndex = 0;
	while ((m = attrRe.exec(raw)) !== null) {
		attrs.push([m[1], m[2]]);
	}
	return attrs;
}

export function federationMentionPlugin(md: MarkdownIt): void {
	md.inline.ruler.before('html_inline', 'fed_mention', (state, silent) => {
		const src = state.src.slice(state.pos);

		const openMatch = openRe.exec(src);
		if (!openMatch) return false;

		// Find the closing tag in the remaining source.
		const afterOpen = state.pos + openMatch[0].length;
		const closeIdx = state.src.indexOf('</fed-mention>', afterOpen);
		if (closeIdx === -1) return false;

		if (silent) return true;

		const attrs = parseAttrs(openMatch[1]);
		const innerText = state.src.slice(afterOpen, closeIdx);

		// Push open token.
		const tokenOpen = state.push('fed_mention_open', 'fed-mention', 1);
		tokenOpen.attrs = attrs;

		// Push inner text.
		const tokenText = state.push('text', '', 0);
		tokenText.content = innerText;

		// Push close token.
		state.push('fed_mention_close', 'fed-mention', -1);

		state.pos = closeIdx + '</fed-mention>'.length;
		return true;
	});
}
