import { mountMarkdownComponents } from '$lib/ui/markdown';

type MarkdownComponentsOptions = {
	root?: HTMLElement;
};

export const markdownComponents = (node: HTMLElement, _options: MarkdownComponentsOptions = {}) => {
	let cleanup = mountMarkdownComponents(node);
	let raf = 0;

	const refresh = () => {
		cancelAnimationFrame(raf);
		raf = requestAnimationFrame(() => {
			cleanup?.();
			cleanup = mountMarkdownComponents(node);
		});
	};

	const observer = new MutationObserver(refresh);
	observer.observe(node, { childList: true, subtree: true });

	return {
		destroy() {
			observer.disconnect();
			cleanup?.();
			cancelAnimationFrame(raf);
		}
	};
};
