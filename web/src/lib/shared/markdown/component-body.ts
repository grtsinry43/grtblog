import type { SvmdNode } from 'svmarkdown';

const urlPattern = /^https?:\/\/\S+$/i;

const collectText = (nodes: SvmdNode[], buffer: string[]) => {
	for (const node of nodes) {
		if (node.kind === 'text') {
			buffer.push(node.value);
			continue;
		}

		if (node.kind === 'break') {
			buffer.push('\n');
			continue;
		}

		if (node.kind === 'code') {
			buffer.push(node.text);
			if (!node.inline) {
				buffer.push('\n');
			}
			continue;
		}

		if (node.kind === 'element' || node.kind === 'component') {
			collectText(node.children, buffer);
			if (node.kind === 'component' || node.block) {
				buffer.push('\n');
			}
		}
	}
};

export const extractPlainTextFromNodes = (nodes?: SvmdNode[]) => {
	if (!nodes?.length) return '';
	const buffer: string[] = [];
	collectText(nodes, buffer);
	return buffer
		.join('')
		.replace(/\n{3,}/g, '\n\n')
		.trim();
};

export const extractImageUrlsFromNodes = (nodes?: SvmdNode[]) => {
	if (!nodes?.length) return [];
	const urls: string[] = [];
	const seen = new Set<string>();

	const walk = (items: SvmdNode[]) => {
		for (const node of items) {
			if (node.kind === 'element' && node.name === 'img') {
				const src = node.attrs?.src;
				if (src && !seen.has(src)) {
					seen.add(src);
					urls.push(src);
				}
			}
			if (node.kind === 'element' || node.kind === 'component') {
				walk(node.children);
			}
		}
	};

	walk(nodes);
	return urls;
};

export const extractUrlsFromBodyText = (bodyText: string) => {
	if (!bodyText) return [];
	const urls: string[] = [];
	const seen = new Set<string>();

	const pushUrl = (raw: string) => {
		const value = raw.trim().replace(/^<|>$/g, '');
		if (!value || !urlPattern.test(value) || seen.has(value)) return;
		seen.add(value);
		urls.push(value);
	};

	for (const rawLine of bodyText.split('\n')) {
		const line = rawLine.trim().replace(/^[-*+]\s+/, '');
		if (!line) continue;

		let hasMarkdownImage = false;
		for (const match of line.matchAll(/!\[[^\]]*]\(([^)\s]+)(?:\s+"[^"]*")?\)/g)) {
			hasMarkdownImage = true;
			if (match[1]) {
				pushUrl(match[1]);
			}
		}

		if (hasMarkdownImage) continue;

		for (const segment of line.split(',')) {
			pushUrl(segment);
		}
	}

	return urls;
};
