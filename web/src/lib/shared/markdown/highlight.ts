import hljs from 'highlight.js/lib/core';
import bash from 'highlight.js/lib/languages/bash';
import css from 'highlight.js/lib/languages/css';
import diff from 'highlight.js/lib/languages/diff';
import go from 'highlight.js/lib/languages/go';
import java from 'highlight.js/lib/languages/java';
import javascript from 'highlight.js/lib/languages/javascript';
import json from 'highlight.js/lib/languages/json';
import kotlin from 'highlight.js/lib/languages/kotlin';
import markdown from 'highlight.js/lib/languages/markdown';
import php from 'highlight.js/lib/languages/php';
import python from 'highlight.js/lib/languages/python';
import typescript from 'highlight.js/lib/languages/typescript';
import yaml from 'highlight.js/lib/languages/yaml';
import xml from 'highlight.js/lib/languages/xml';

hljs.registerLanguage('bash', bash);
hljs.registerLanguage('css', css);
hljs.registerLanguage('diff', diff);
hljs.registerLanguage('go', go);
hljs.registerLanguage('java', java);
hljs.registerLanguage('javascript', javascript);
hljs.registerLanguage('json', json);
hljs.registerLanguage('kotlin', kotlin);
hljs.registerLanguage('markdown', markdown);
hljs.registerLanguage('php', php);
hljs.registerLanguage('python', python);
hljs.registerLanguage('typescript', typescript);
hljs.registerLanguage('yaml', yaml);
hljs.registerLanguage('xml', xml);

const langAlias: Record<string, string> = {
	js: 'javascript',
	ts: 'typescript',
	tsx: 'typescript',
	html: 'xml',
	xhtml: 'xml',
	svelte: 'xml',
	plaintext: 'markdown'
};

const resolveLanguage = (lang?: string) => {
	const raw = (lang || '').trim().toLowerCase();
	const normalized = raw === 'text' ? 'plaintext' : raw;
	const resolved = langAlias[normalized] ?? normalized;
	return hljs.getLanguage(resolved) ? resolved : 'markdown';
};

export const highlightCode = (code: string, lang?: string) => {
	const language = resolveLanguage(lang);
	const result = hljs.highlight(code ?? '', { language, ignoreIllegals: true });
	const html = result.value || hljs.escapeHTML(code ?? '');
	return `<pre><code class="hljs language-${language}">${html}</code></pre>`;
};
