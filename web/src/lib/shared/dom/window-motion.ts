/**
 * Web Animations for the floating window. Each function returns the Animation so callers can
 * cancel it when the next state change arrives. All respect prefers-reduced-motion.
 */

const EASE_OUT_SOFT = 'cubic-bezier(0.16, 1, 0.3, 1)';
/** Accelerating curve: the window "gets sucked in" toward the dock. */
const EASE_SUCK = 'cubic-bezier(0.55, 0, 0.8, 0.35)';

const prefersReducedMotion = () =>
	typeof window !== 'undefined' && window.matchMedia('(prefers-reduced-motion: reduce)').matches;

const duration = (ms: number) => (prefersReducedMotion() ? 1 : ms);

export function popIn(el: HTMLElement): Animation {
	return el.animate(
		[
			{ opacity: 0, transform: 'translateY(8px) scale(0.96)', filter: 'blur(4px)' },
			{ opacity: 1, transform: 'none', filter: 'blur(0)' }
		],
		{ duration: duration(320), easing: EASE_OUT_SOFT }
	);
}

export function popOut(el: HTMLElement): Animation {
	return el.animate(
		[
			{ opacity: 1, transform: 'none', filter: 'blur(0)' },
			{ opacity: 0, transform: 'translateY(4px) scale(0.97)', filter: 'blur(2px)' }
		],
		{ duration: duration(180), easing: 'cubic-bezier(0.4, 0, 1, 1)', fill: 'forwards' }
	);
}

/**
 * Genie-style minimize/restore between the window and a dock target. A clip-path funnel
 * pinches the edge facing the dock while the window shrinks into it, approximating the
 * macOS mesh warp with compositor-friendly properties.
 */
export function genie(el: HTMLElement, dock: DOMRect, direction: 'out' | 'in'): Animation {
	const win = el.getBoundingClientRect();
	const dx = dock.left + dock.width / 2 - (win.left + win.width / 2);
	const dy = dock.top + dock.height / 2 - (win.top + win.height / 2);
	const sx = Math.max(0.04, dock.width / win.width);
	const sy = Math.max(0.04, dock.height / win.height);

	// Pinch the edge that faces the dock.
	const horizontal = Math.abs(dx) >= Math.abs(dy);
	const full = 'polygon(0% 0%, 100% 0%, 100% 100%, 0% 100%)';
	const pinch = (amount: number) => {
		const a = `${amount}%`;
		const b = `${100 - amount}%`;
		if (horizontal) {
			return dx < 0
				? `polygon(0% ${a}, 100% 0%, 100% 100%, 0% ${b})`
				: `polygon(0% 0%, 100% ${a}, 100% ${b}, 0% 100%)`;
		}
		return dy > 0
			? `polygon(0% 0%, 100% 0%, ${b} 100%, ${a} 100%)`
			: `polygon(${a} 0%, ${b} 0%, 100% 100%, 0% 100%)`;
	};

	const frames: Keyframe[] = [
		{ offset: 0, transform: 'none', clipPath: full, opacity: 1, filter: 'blur(0)' },
		{
			offset: 0.4,
			transform: `translate(${dx * 0.25}px, ${dy * 0.25}px) scale(${horizontal ? 0.72 : 0.88}, ${horizontal ? 0.88 : 0.72})`,
			clipPath: pinch(28),
			opacity: 1,
			filter: 'blur(0)'
		},
		{
			offset: 1,
			transform: `translate(${dx}px, ${dy}px) scale(${sx}, ${sy})`,
			clipPath: pinch(42),
			opacity: 0,
			filter: 'blur(2px)'
		}
	];

	const isOut = direction === 'out';
	const keyframes = isOut
		? frames
		: [...frames].reverse().map((f) => ({ ...f, offset: 1 - (f.offset as number) }));
	return el.animate(keyframes, {
		duration: duration(isOut ? 520 : 460),
		easing: isOut ? EASE_SUCK : EASE_OUT_SOFT,
		fill: isOut ? 'forwards' : 'none'
	});
}

/** Smoothly morph between two layout rects (maximize/restore). Animates layout properties
 *  rather than transform so content reflows naturally instead of being stretched. */
export function morphRect(el: HTMLElement, from: DOMRect, to: DOMRect): Animation | null {
	const unchanged = (['left', 'top', 'width', 'height'] as const).every(
		(k) => Math.abs(from[k] - to[k]) < 1
	);
	if (unchanged) return null;
	const rect = (r: DOMRect) => ({
		left: `${r.left}px`,
		top: `${r.top}px`,
		width: `${r.width}px`,
		height: `${r.height}px`
	});
	return el.animate([rect(from), rect(to)], { duration: duration(380), easing: EASE_OUT_SOFT });
}

/** Ease the window's top edge when auto-recentering after its content height changes. */
export function slideTop(el: HTMLElement, fromTop: number, toTop: number): Animation | null {
	if (Math.abs(fromTop - toTop) < 2) return null;
	return el.animate([{ top: `${fromTop}px` }, { top: `${toTop}px` }], {
		duration: duration(240),
		easing: EASE_OUT_SOFT
	});
}
