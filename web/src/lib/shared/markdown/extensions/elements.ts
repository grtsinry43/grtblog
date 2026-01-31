import { createHighlighter } from 'shiki';
import type MarkdownIt from 'markdown-it';
import type { MarkdownExtension } from '../types';

const highlighter = await createHighlighter({
	themes: ['github-light', 'github-dark'],
	langs: [
		'plaintext',
		'bash',
		'go',
		'css',
		'diff',
		'html',
		'javascript',
		'java',
		'python',
		'php',
		'kotlin',
		'js',
		'ts',
		'json',
		'markdown',
		'svelte',
		'tsx',
		'typescript',
		'yaml',
		'toml'
	]
});

const escapeAttr = (md: MarkdownIt, value: string) => md.utils.escapeHtml(value);

const renderImage = (md: MarkdownIt, src: string, alt: string, title: string) => {
	const safeSrc = escapeAttr(md, src);
	const safeAlt = escapeAttr(md, alt);
	const safeTitle = escapeAttr(md, title);
	const titleAttr = title ? ` title="${safeTitle}"` : '';
	const caption = title ? `<figcaption class="md-caption">${safeTitle}</figcaption>` : '';
	const propsJson = JSON.stringify({ src, alt, title });
	const propsAttr = propsJson !== '{}' ? ` data-props="${escapeAttr(md, propsJson)}"` : '';

	return `<figure class="md-figure md-component-placeholder" data-component="md-image"${propsAttr} data-mount-target=".md-image__mount"><img class="md-img" src="${safeSrc}" alt="${safeAlt}"${titleAttr} loading="lazy" decoding="async" data-loaded="false" />${caption}<span class="md-image__mount" aria-hidden="true" style="display:none"></span></figure>`;
};

const renderCodeBlock = (md: MarkdownIt, lang: string, codeHtml: string) => {
	const safeLang = escapeAttr(md, lang);
	return `<div class="md-codeblock" data-lang="${safeLang}"><div class="md-codeblock__header"><span class="md-codeblock__lang">${safeLang || 'text'}</span></div><div class="md-codeblock__body">${codeHtml}</div></div>`;
};

const resolveLinkSite = (href: string, origin?: string) => {
	if (!href) return null;
	if (href.startsWith('#') || href.startsWith('/')) return 'internal';

	try {
		const base = origin || (typeof window !== 'undefined' ? window.location.origin : undefined);
		const url = base ? new URL(href, base) : new URL(href);
		if (base && url.origin === base) return 'internal';
		const host = url.hostname.toLowerCase();
		if (host === 'github.com' || host.endsWith('.github.com')) return 'github';
		if (host === 'bilibili.com' || host.endsWith('.bilibili.com') || host === 'b23.tv')
			return 'bilibili';
		if (host === 'leetcode.com' || host.endsWith('.leetcode.com') || host === 'leetcode.cn')
			return 'leetcode';
	} catch {
		return null;
	}

	return null;
};

export const markdownElementsExtension: MarkdownExtension = (md) => {
	md.renderer.rules.link_open = (tokens, idx, options, env, self) => {
		const token = tokens[idx];
		const href = token.attrGet('href') ?? '';
		const site = resolveLinkSite(href, (env as any)?.origin);
		const isExternal =
			!site && (/^https?:\/\//i.test(href) || href.startsWith('//'));
		const existingClass = token.attrGet('class');
		token.attrSet('class', existingClass ? `${existingClass} md-link` : 'md-link');
		if (site) token.attrSet('data-site', site);
		if (isExternal) {
			token.attrSet('rel', 'noopener noreferrer');
			token.attrSet('target', '_blank');
		}
		return self.renderToken(tokens, idx, options);
	};

	md.renderer.rules.link_close = () => '<span class="md-link__icon" aria-hidden="true"></span></a>';

	md.renderer.rules.image = (tokens, idx) => {
		const token = tokens[idx];
		const src = token.attrGet('src') ?? '';
		const alt = token.content || token.attrGet('alt') || '';
		const title = token.attrGet('title') ?? '';
		return renderImage(md, src, alt, title);
	};

	md.renderer.rules.fence = (tokens, idx) => {
		const token = tokens[idx];
		const info = (token.info || '').trim();
		const rawLang = info.split(/\s+/)[0] || 'plaintext';
		const lang = rawLang === 'text' ? 'plaintext' : rawLang;
		const code = token.content ?? '';
		const resolvedLang = highlighter.getLoadedLanguages().includes(lang) ? lang : 'plaintext';
		const codeHtml = highlighter.codeToHtml(code, {
			lang: resolvedLang,
			themes: {
				light: 'github-light',
				dark: 'github-dark'
			}
		});
		return renderCodeBlock(md, lang, codeHtml);
	};

	md.renderer.rules.code_block = (tokens, idx) => {
		const token = tokens[idx];
		const code = token.content ?? '';
		const codeHtml = highlighter.codeToHtml(code, {
			lang: 'plaintext',
			themes: {
				light: 'github-light',
				dark: 'github-dark'
			}
		});
		return renderCodeBlock(md, 'plaintext', codeHtml);
	};
};
