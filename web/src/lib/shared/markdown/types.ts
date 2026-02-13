import type MarkdownIt from 'markdown-it';
import type { Options } from 'markdown-it';

export type MarkdownExtension = (md: MarkdownIt, options?: unknown) => void;

export type MarkdownConfig = {
	options?: Options;
	extensions?: MarkdownExtension[];
};
