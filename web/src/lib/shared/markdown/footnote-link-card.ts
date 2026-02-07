import type MarkdownIt from 'markdown-it';
import type Token from 'markdown-it/lib/token.mjs';

type FootnoteLinkInfo = {
	href: string;
	title?: string;
	desc?: string;
};

const urlRegex = /(https?:\/\/[^\s)]+)|(^\/[^\s)]+)/i;

const deriveTitleFromHref = (href: string) => {
	try {
		if (href.startsWith('http')) {
			const url = new URL(href);
			const last = url.pathname.split('/').filter(Boolean).pop();
			return last ? decodeURIComponent(last).replace(/[-_]/g, ' ') : url.hostname;
		}
		const last = href.split('/').filter(Boolean).pop();
		return last ? decodeURIComponent(last).replace(/[-_]/g, ' ') : href;
	} catch {
		return href;
	}
};

const findLinkInChildren = (children?: Token[]) => {
	if (!children?.length) return null;
	let inLink = false;
	let href = '';
	let text = '';

	for (const child of children) {
		if (child.type === 'link_open') {
			href = child.attrGet('href') ?? '';
			inLink = true;
			continue;
		}
		if (child.type === 'link_close' && inLink) {
			return href ? { href, text: text.trim() } : null;
		}
		if (inLink && (child.type === 'text' || child.type === 'code_inline')) {
			text += child.content;
			continue;
		}
		if (child.children?.length) {
			const nested = findLinkInChildren(child.children);
			if (nested) return nested;
		}
	}

	return null;
};

const findLinkInTokens = (tokens: Token[]) => {
	for (const token of tokens) {
		if (token.type === 'text' || token.type === 'code_inline') {
			const match = token.content.match(urlRegex);
			if (match?.[0]) {
				return { href: match[0], text: match[0] };
			}
		}
		if (token.type === 'inline' && token.children?.length) {
			const link = findLinkInChildren(token.children);
			if (link?.href) return link;

			const match = token.content.match(urlRegex);
			if (match?.[0]) {
				return { href: match[0], text: match[0] };
			}
		}
	}
	return null;
};

const collectText = (tokens: Token[]) => {
	let text = '';
	for (const token of tokens) {
		if (token.type === 'inline' && token.children?.length) {
			text += collectText(token.children);
			continue;
		}
		if (token.type === 'text' || token.type === 'code_inline') {
			text += token.content;
			continue;
		}
		if (token.children?.length) {
			text += collectText(token.children);
		}
	}
	return text.replace(/\s+/g, ' ').trim();
};

const buildFootnoteLinkMap = (tokens: Array<{ tokens?: Token[]; content?: string }>) => {
	const map: Record<string, FootnoteLinkInfo> = {};
	for (let i = 0; i < tokens.length; i += 1) {
		const entry = tokens[i];
		const entryTokens = entry?.tokens || [];
		const link = findLinkInTokens(entryTokens);
		if (!link?.href) continue;
		const title = link.text?.trim() || deriveTitleFromHref(link.href);
		const desc = entryTokens.length ? collectText(entryTokens) : (entry?.content || '').trim();
		map[String(i)] = { href: link.href, title, desc };
	}
	return map;
};

const injectLinkCards = (state: any, tokens: Token[], map: Record<string, FootnoteLinkInfo>) => {
	for (let i = 0; i < tokens.length; i += 1) {
		const token = tokens[i];
		if (token.type === 'footnote_ref') {
			const id = token.meta?.id;
			const info = typeof id === 'number' ? map[String(id)] : null;
			if (info?.href) {
				const open = new state.Token('footnote_link_card_open', 'footnote-link-card', 1);
				open.attrs = [
					['href', info.href],
					['title', info.title || deriveTitleFromHref(info.href)],
					['desc', info.desc || '']
				];
				const close = new state.Token('footnote_link_card_close', 'footnote-link-card', -1);
				tokens.splice(i + 1, 0, open, close);
				i += 2;
			}
		}
		if (token.children?.length) {
			injectLinkCards(state, token.children, map);
		}
	}
};

export const footnoteLinkCardPlugin = (md: MarkdownIt) => {
	md.core.ruler.after('footnote_tail', 'footnote_link_card', (state) => {
		const env = state.env as any;
		if (!env.footnotes?.list?.length) return;
		const list = env.footnotes.list as Array<{ tokens?: Token[]; content?: string }>;
		const map = buildFootnoteLinkMap(list);
		if (!Object.keys(map).length) return;
		injectLinkCards(state, state.tokens, map);
	});
};
