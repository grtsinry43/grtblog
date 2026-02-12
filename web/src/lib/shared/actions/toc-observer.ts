import { browser } from '$app/environment';
import { tick } from 'svelte';

export type TocObserverOptions = {
	/** Called when the active heading changes. */
	onActiveChange: (anchor: string | null) => void;
	/** rootMargin for the IntersectionObserver (default: '0px 0px -70% 0px'). */
	rootMargin?: string;
};

/**
 * Svelte action that observes h1–h6 headings inside a container
 * and reports the topmost visible heading's id via `onActiveChange`.
 *
 * Usage:
 * ```svelte
 * <div use:tocObserver={{ onActiveChange: (a) => (activeAnchor = a) }}>
 * ```
 */
export function tocObserver(node: HTMLElement, options: TocObserverOptions) {
	if (!browser) return { update() {}, destroy() {} };

	let observer: IntersectionObserver | null = null;
	let currentOptions = options;

	const setup = () => {
		observer?.disconnect();
		const headings = node.querySelectorAll('h1, h2, h3, h4, h5, h6');
		if (!headings.length) {
			currentOptions.onActiveChange(null);
			return;
		}
		observer = new IntersectionObserver(
			(entries) => {
				const visible = entries.filter((e) => e.isIntersecting);
				if (!visible.length) return;
				visible.sort((a, b) => a.boundingClientRect.top - b.boundingClientRect.top);
				const target = visible[0]?.target as HTMLElement | undefined;
				if (target?.id) currentOptions.onActiveChange(target.id);
			},
			{ rootMargin: currentOptions.rootMargin ?? '0px 0px -70% 0px', threshold: 0 }
		);
		headings.forEach((h) => observer?.observe(h));
	};

	// Defer setup so content rendered via $effect has time to paint.
	tick().then(setup);

	return {
		update(newOptions: TocObserverOptions) {
			currentOptions = newOptions;
			tick().then(setup);
		},
		destroy() {
			observer?.disconnect();
		}
	};
}
