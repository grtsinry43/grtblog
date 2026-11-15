import { browser } from '$app/environment';

export type TextSelectionInfo = {
	text: string;
	/** Viewport-relative rect of the selection, refreshed on scroll/resize. */
	rect: DOMRect;
	/** Pointer that produced the selection; touch UIs place their own menu above it. */
	pointerType: string;
};

export type TextSelectionOptions = {
	onChange: (selection: TextSelectionInfo | null) => void;
	/** Ignore selections shorter than this (after trimming). */
	minLength?: number;
};

/**
 * Reports non-empty text selections made inside `node`. Fires `null` when the selection
 * collapses or leaves the node. Reports only after the pointer is released, so a toolbar
 * doesn't chase the cursor while the user is still dragging.
 */
export function textSelection(node: HTMLElement, options: TextSelectionOptions) {
	if (!browser) return { update() {}, destroy() {} };

	let { onChange, minLength = 2 } = options;
	let pointerDown = false;
	let pointerType = 'mouse';
	let raf = 0;
	let current: TextSelectionInfo | null = null;

	const emit = (next: TextSelectionInfo | null) => {
		if (!next && !current) return;
		current = next;
		onChange(next);
	};

	const read = () => {
		const sel = document.getSelection();
		if (!sel || sel.isCollapsed || sel.rangeCount === 0) return emit(null);
		const range = sel.getRangeAt(0);
		if (!node.contains(range.commonAncestorContainer)) return emit(null);
		const text = sel.toString().trim();
		if (Array.from(text).length < minLength) return emit(null);
		const rect = range.getBoundingClientRect();
		if (rect.width === 0 && rect.height === 0) return emit(null);
		emit({ text, rect, pointerType });
	};

	const schedule = () => {
		cancelAnimationFrame(raf);
		raf = requestAnimationFrame(read);
	};

	const onSelectionChange = () => {
		if (pointerDown) return;
		schedule();
	};
	const onPointerDown = (e: PointerEvent) => {
		pointerDown = true;
		pointerType = e.pointerType || 'mouse';
	};
	const onPointerUp = () => {
		if (!pointerDown) return;
		pointerDown = false;
		schedule();
	};
	const onViewportChange = () => {
		if (current) schedule();
	};

	document.addEventListener('selectionchange', onSelectionChange);
	node.addEventListener('pointerdown', onPointerDown);
	document.addEventListener('pointerup', onPointerUp);
	document.addEventListener('pointercancel', onPointerUp);
	window.addEventListener('scroll', onViewportChange, { passive: true });
	window.addEventListener('resize', onViewportChange, { passive: true });

	return {
		update(next: TextSelectionOptions) {
			({ onChange, minLength = 2 } = next);
		},
		destroy() {
			cancelAnimationFrame(raf);
			document.removeEventListener('selectionchange', onSelectionChange);
			node.removeEventListener('pointerdown', onPointerDown);
			document.removeEventListener('pointerup', onPointerUp);
			document.removeEventListener('pointercancel', onPointerUp);
			window.removeEventListener('scroll', onViewportChange);
			window.removeEventListener('resize', onViewportChange);
		}
	};
}
