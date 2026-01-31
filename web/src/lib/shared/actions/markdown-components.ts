import { imageLazy } from '$lib/shared/actions/image-lazy';
import { websiteInfoCtx } from '$lib/features/website-info/context';
import { getSiteIconUrl, isSiteKey, resolveLinkSite } from '$lib/shared/markdown/link-icons';
import { mountMarkdownComponents, unmountMarkdownComponentsIn } from '$lib/ui/markdown';

type MarkdownComponentsOptions = {
	root?: HTMLElement;
};

export const markdownComponents = (node: HTMLElement, _options: MarkdownComponentsOptions = {}) => {
	const imageCleanups = new WeakMap<HTMLImageElement, () => void>();
	const linkEnhanced = new WeakSet<HTMLAnchorElement>();
	const siteFaviconStore = websiteInfoCtx.selectModelData((data) => data?.favicon || '');
	let siteFavicon = '';
	const unsubscribeFavicon = siteFaviconStore.subscribe((value) => {
		siteFavicon = value || '';
	});

	const enhanceLinks = (root: HTMLElement) => {
		const links = Array.from(root.querySelectorAll<HTMLAnchorElement>('a.md-link')).filter(
			(link) => !linkEnhanced.has(link)
		);

		for (const link of links) {
			const href = link.getAttribute('href') || '';
			const cached = link.dataset.site;
			const site =
				cached && isSiteKey(cached)
					? cached
					: resolveLinkSite(href, window.location.origin);
			if (site) link.dataset.site = site;
			if (!site) continue;

			const icon = link.querySelector<HTMLElement>('.md-link__icon') || (() => {
				const span = document.createElement('span');
				span.className = 'md-link__icon';
				span.setAttribute('aria-hidden', 'true');
				link.appendChild(span);
				return span;
			})();

			const url = getSiteIconUrl(site, siteFavicon);
			if (url) icon.style.setProperty('--md-link-icon-url', `url("${url}")`);
			linkEnhanced.add(link);
		}
	};

	const mountImages = (root: HTMLElement) => {
		const images = Array.from(root.querySelectorAll<HTMLImageElement>('img.md-img')).filter(
			(img) => !imageCleanups.has(img) && img.dataset.mdImageMounted !== 'true'
		);

		for (const img of images) {
			const { destroy } = imageLazy(img, { src: img.currentSrc || img.src });
			img.dataset.mdImageMounted = 'true';
			imageCleanups.set(img, () => {
				destroy?.();
				img.dataset.mdImageMounted = 'false';
			});
		}
	};

	const mountAll = (root: HTMLElement) => {
		const hasPlaceholders =
			root.matches('.md-component-placeholder') ||
			root.querySelector('.md-component-placeholder') !== null;
		const hasImages = root.matches('img.md-img') || root.querySelector('img.md-img') !== null;
		const hasLinks = root.matches('a.md-link') || root.querySelector('a.md-link') !== null;

		if (hasPlaceholders) {
			mountMarkdownComponents(root);
		}
		if (hasImages) {
			mountImages(root);
		}
		if (hasLinks) {
			enhanceLinks(root);
		}
	};

	mountAll(node);

	const observer = new MutationObserver((mutations) => {
		for (const mutation of mutations) {
			for (const added of mutation.addedNodes) {
				if (added instanceof HTMLElement) {
					mountAll(added);
				}
			}
			for (const removed of mutation.removedNodes) {
				if (removed instanceof HTMLElement) {
					unmountMarkdownComponentsIn(removed);
					const images = removed.matches('img.md-img')
						? [removed as HTMLImageElement]
						: Array.from(removed.querySelectorAll<HTMLImageElement>('img.md-img'));
					for (const img of images) {
						const cleanup = imageCleanups.get(img);
						if (cleanup) {
							cleanup();
							imageCleanups.delete(img);
						}
					}
				}
			}
		}
	});
	observer.observe(node, { childList: true, subtree: true });

	return {
		destroy() {
			observer.disconnect();
			unsubscribeFavicon();
			unmountMarkdownComponentsIn(node);
			const images = node.querySelectorAll<HTMLImageElement>('img.md-img');
			for (const img of images) {
				const cleanup = imageCleanups.get(img);
				if (cleanup) cleanup();
			}
		}
	};
};
