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

const renderComponentPlaceholder = (
	md: MarkdownIt,
	component: string,
	props: Record<string, string>,
	contentHtml = ''
) => {
	const propsJson = JSON.stringify(props);
	const propsAttr = propsJson !== '{}' ? ` data-props="${md.utils.escapeHtml(propsJson)}"` : '';
	return `<span class="md-component-placeholder" data-component="${md.utils.escapeHtml(component)}"${propsAttr}>${contentHtml}</span>`;
};

const renderBlockComponentPlaceholder = (
	md: MarkdownIt,
	component: string,
	props: Record<string, string>,
	contentHtml = ''
) => {
	const propsJson = JSON.stringify(props);
	const propsAttr = propsJson !== '{}' ? ` data-props="${md.utils.escapeHtml(propsJson)}"` : '';
	return `<div class="md-component-placeholder" data-component="${md.utils.escapeHtml(component)}"${propsAttr}>${contentHtml}</div>`;
};

export const markdownElementsExtension: MarkdownExtension = (md) => {
	md.renderer.rules.link_open = (tokens, idx) => {
		const token = tokens[idx];
		const href = token.attrGet('href') ?? '';
		const title = token.attrGet('title') ?? '';
		const props = { href, title };
		return renderComponentPlaceholder(md, 'md-link', props);
	};

	md.renderer.rules.link_close = () => '</span>';

	md.renderer.rules.image = (tokens, idx) => {
		const token = tokens[idx];
		const src = token.attrGet('src') ?? '';
		const alt = token.content || token.attrGet('alt') || '';
		const title = token.attrGet('title') ?? '';
		return renderBlockComponentPlaceholder(md, 'md-image', { src, alt, title });
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
		return renderBlockComponentPlaceholder(md, 'md-codeblock', { lang }, codeHtml);
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
		return renderBlockComponentPlaceholder(md, 'md-codeblock', { lang: 'plaintext' }, codeHtml);
	};
};
