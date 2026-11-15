/**
 * True when the element was server-rendered into the initial viewport of a hard load
 * (no client-side navigation has happened yet). Entrance animations should skip such
 * elements: the user has already seen them, so snapping them to hidden causes a flash.
 */
export function isVisibleOnFirstPaint(node: Element | undefined | null): boolean {
	if (!node || typeof window === 'undefined') return false;
	if (document.documentElement.dataset.navType) return false;
	const rect = node.getBoundingClientRect();
	return rect.top < window.innerHeight && rect.bottom > 0;
}
