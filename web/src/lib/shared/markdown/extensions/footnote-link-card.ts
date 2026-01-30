import type MarkdownIt from 'markdown-it';
import type Token from 'markdown-it/lib/token.mjs';

import type { MarkdownExtension } from '../types';

type FootnoteLinkInfo = {
	href: string;
	title?: string;
	desc?: string;
};

const renderComponentPlaceholder = (
	md: MarkdownIt,
	component: string,
	props: Record<string, string>
) => {
	const propsJson = JSON.stringify(props);
	const propsAttr = propsJson !== '{}' ? ` data-props="${md.utils.escapeHtml(propsJson)}"` : '';
	return `<div class="md-component-placeholder" data-component="${md.utils.escapeHtml(component)}"${propsAttr}></div>`;
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

const resolveHref = (href: string, origin?: string) => {
	if (href.startsWith('/')) return href;
	if (!origin) return href;
	try {
		return new URL(href, origin).href;
	} catch {
		return href;
	}
};

export const footnoteLinkCardExtension: MarkdownExtension = (md) => {
	md.core.ruler.after('footnote_tail', 'footnote_linkcard_map', (state) => {
		const env = state.env as any;
		if (!env.footnotes?.list?.length) return;
		const map: Record<string, FootnoteLinkInfo> = {};

		const list = env.footnotes.list as Array<{ tokens?: Token[]; content?: string }>;
		for (let i = 0; i < list.length; i += 1) {
			const entry = list[i];
			const tokens = entry?.tokens || [];
			const link = findLinkInTokens(tokens);
			if (!link?.href) continue;
			const title = link.text?.trim() || deriveTitleFromHref(link.href);
			const desc = tokens.length ? collectText(tokens) : (entry?.content || '').trim();
			map[String(i)] = { href: link.href, title, desc };
		}

		env.footnoteLinkMap = map;
	});

	const defaultFootnoteRef =
		md.renderer.rules.footnote_ref ||
		((tokens, idx, options, env, self) => self.renderToken(tokens, idx, options));

	md.renderer.rules.footnote_ref = (tokens, idx, options, env, self) => {
		const base = defaultFootnoteRef(tokens, idx, options, env, self);
		const map = (env as any)?.footnoteLinkMap as Record<string, FootnoteLinkInfo> | undefined;
		const id = tokens[idx].meta?.id;
		if (!map || typeof id !== 'number') return base;
		const info = map[String(id)];
		if (!info?.href) return base;

		const origin = (env as any)?.origin as string | undefined;
		const href = resolveHref(info.href, origin);

		const props = {
			href,
			title: info.title || deriveTitleFromHref(href),
			desc: info.desc || ''
		};
		return `${base}${renderComponentPlaceholder(md, 'footnote-link-card', props)}`;
	};
};
