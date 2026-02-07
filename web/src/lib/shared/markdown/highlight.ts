import { createHighlighterCore } from 'shiki/core';
import { createOnigurumaEngine } from 'shiki/engine/oniguruma';
import { getWasmInstance } from 'shiki/wasm';
import langBash from '@shikijs/langs/bash';
import langCss from '@shikijs/langs/css';
import langDiff from '@shikijs/langs/diff';
import langGo from '@shikijs/langs/go';
import langHtml from '@shikijs/langs/html';
import langJava from '@shikijs/langs/java';
import langJavaScript from '@shikijs/langs/javascript';
import langJson from '@shikijs/langs/json';
import langKotlin from '@shikijs/langs/kotlin';
import langMarkdown from '@shikijs/langs/markdown';
import langPhp from '@shikijs/langs/php';
import langPython from '@shikijs/langs/python';
import langSvelte from '@shikijs/langs/svelte';
import langToml from '@shikijs/langs/toml';
import langTsx from '@shikijs/langs/tsx';
import langTypeScript from '@shikijs/langs/typescript';
import langYaml from '@shikijs/langs/yaml';
import themeGithubDark from '@shikijs/themes/github-dark';
import themeGithubLight from '@shikijs/themes/github-light';

const langAlias: Record<string, string> = {
	js: 'javascript',
	ts: 'typescript'
};

const highlighter = await createHighlighterCore({
	themes: [themeGithubLight, themeGithubDark],
	langs: [
		langBash,
		langGo,
		langCss,
		langDiff,
		langHtml,
		langJavaScript,
		langJava,
		langPython,
		langPhp,
		langKotlin,
		langJson,
		langMarkdown,
		langSvelte,
		langTsx,
		langTypeScript,
		langYaml,
		langToml
	],
	engine: createOnigurumaEngine(getWasmInstance),
	langAlias
});

export const highlightCode = (code: string, lang?: string) => {
	const rawLang = (lang || 'plaintext').trim();
	const normalizedLang = rawLang === 'text' ? 'plaintext' : rawLang;
	const resolvedLang = highlighter.getLoadedLanguages().includes(normalizedLang)
		? normalizedLang
		: 'plaintext';

	return highlighter.codeToHtml(code ?? '', {
		lang: resolvedLang,
		themes: {
			light: 'github-light',
			dark: 'github-dark'
		}
	});
};
