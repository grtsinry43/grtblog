import { browser } from '$app/environment';

export type IntersectOptions = {
	onEnter?: (entry: IntersectionObserverEntry) => void;
	onLeave?: (entry: IntersectionObserverEntry) => void;
	once?: boolean;
	threshold?: number;
	rootMargin?: string;
};

export function intersect(node: HTMLElement, options: IntersectOptions = {}) {
	if (!browser) return { update() {}, destroy() {} };

	const { onEnter, onLeave, once = true, threshold = 0.1, rootMargin } = options;

	const observer = new IntersectionObserver(
		(entries) => {
			for (const entry of entries) {
				if (entry.isIntersecting) {
					onEnter?.(entry);
					if (once) observer.disconnect();
				} else {
					onLeave?.(entry);
				}
			}
		},
		{ threshold, rootMargin }
	);

	// Defer observe to next frame so $effect-driven hidden state
	// gets at least one paint before the observer can fire onEnter.
	const raf = requestAnimationFrame(() => {
		observer.observe(node);
	});

	return {
		destroy() {
			cancelAnimationFrame(raf);
			observer.disconnect();
		}
	};
}
