/**
 * April Fools 2026 — click-to-fall physics effect.
 *
 * Encapsulates all browser API access (addEventListener, Web Animations API)
 * for the prank feature. Designed to be called only from onMount (SSR-safe).
 */

const STORAGE_KEY = 'aprilFools2026';

const INTERACTIVE =
	'a, button, input, select, textarea, [role="button"], [tabindex], summary, label';

// ---------------------------------------------------------------------------
// localStorage state
// ---------------------------------------------------------------------------

export type AprilFoolsState = 'fallen' | 'done' | null;

export function getState(): AprilFoolsState {
	try {
		const v = localStorage.getItem(STORAGE_KEY);
		if (v === 'fallen' || v === 'done') return v;
		return null;
	} catch {
		return null;
	}
}

export function setState(s: AprilFoolsState): void {
	try {
		if (s === null) localStorage.removeItem(STORAGE_KEY);
		else localStorage.setItem(STORAGE_KEY, s);
	} catch {
		/* quota / private mode */
	}
}

// ---------------------------------------------------------------------------
// Fall animation
// ---------------------------------------------------------------------------

function findInteractive(el: Element): HTMLElement | null {
	if (el.matches(INTERACTIVE)) return el as HTMLElement;
	return el.closest(INTERACTIVE) as HTMLElement | null;
}

function applyFall(el: HTMLElement): void {
	const rect = el.getBoundingClientRect();
	const dist = window.innerHeight - rect.top + 80;
	const rot = (Math.random() - 0.5) * 50; // -25° ~ +25°

	el.style.pointerEvents = 'none';

	el.animate(
		[
			{
				transform: 'translateY(0) rotate(0deg)',
				easing: 'cubic-bezier(0.4, 0, 1, 1)', // accelerate (gravity)
				offset: 0
			},
			{
				transform: `translateY(${dist}px) rotate(${rot}deg)`,
				easing: 'cubic-bezier(0, 0, 0.2, 1)', // decelerate (bounce up)
				offset: 0.55
			},
			{
				transform: `translateY(${dist - 22}px) rotate(${rot * 0.92}deg)`,
				easing: 'cubic-bezier(0.4, 0, 1, 1)',
				offset: 0.7
			},
			{
				transform: `translateY(${dist}px) rotate(${rot}deg)`,
				easing: 'cubic-bezier(0, 0, 0.2, 1)',
				offset: 0.82
			},
			{
				transform: `translateY(${dist - 6}px) rotate(${rot * 0.98}deg)`,
				easing: 'cubic-bezier(0.4, 0, 1, 1)',
				offset: 0.91
			},
			{
				transform: `translateY(${dist + 50}px) rotate(${rot * 1.2}deg)`,
				opacity: 0,
				offset: 1
			}
		],
		{
			duration: 900 + Math.random() * 300,
			fill: 'forwards'
		}
	);
}

// ---------------------------------------------------------------------------
// Click interceptor
// ---------------------------------------------------------------------------

/**
 * Intercepts clicks on interactive elements in capture phase,
 * prevents default behaviour, and triggers a gravity-fall animation.
 *
 * @param onFirstFall Called once when the very first element falls.
 * @returns Cleanup function to remove the listener.
 */
export function setupFallInterceptor(onFirstFall: () => void): () => void {
	let firstFired = false;

	function onClick(e: MouseEvent) {
		const target = findInteractive(e.target as Element);
		if (!target) return;

		// Don't intercept the dismiss button inside the popup
		if (target.closest('[data-april-fools-popup]')) return;

		e.preventDefault();
		e.stopPropagation();

		if (!firstFired) {
			firstFired = true;
			onFirstFall();
		}

		applyFall(target);
	}

	document.addEventListener('click', onClick, true);

	return () => {
		document.removeEventListener('click', onClick, true);
	};
}
