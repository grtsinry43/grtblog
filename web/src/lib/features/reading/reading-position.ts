import { browser } from '$app/environment';

const STORAGE_PREFIX = 'reading-pos:';
const MAX_AGE_MS = 90 * 24 * 3600 * 1000;
/** Below this the reader has barely started; above it they've essentially finished. */
const MIN_RATIO = 0.08;
const FINISHED_RATIO = 0.92;
/** Reading line: the point in the viewport treated as "where the eye is". */
const READING_LINE = 0.3;

type SavedPosition = { ratio: number; savedAt: number };

export type ReadingPositionOptions = {
	/** Stable per-document key, e.g. `article:42`. */
	key: string;
	/** Called once when a previous position is worth offering to the reader. */
	onResumeAvailable?: (ratio: number, resume: () => void) => void;
};

function load(key: string): SavedPosition | null {
	try {
		const raw = JSON.parse(localStorage.getItem(STORAGE_PREFIX + key) ?? 'null');
		if (!raw || typeof raw.ratio !== 'number' || typeof raw.savedAt !== 'number') return null;
		return raw;
	} catch {
		return null;
	}
}

function save(key: string, ratio: number) {
	try {
		if (ratio >= FINISHED_RATIO) localStorage.removeItem(STORAGE_PREFIX + key);
		else if (ratio >= MIN_RATIO)
			localStorage.setItem(STORAGE_PREFIX + key, JSON.stringify({ ratio, savedAt: Date.now() }));
	} catch {
		// storage unavailable — resume just won't be offered
	}
}

/**
 * Remembers how far into `node` (the article body) the reader got, as a ratio of its height
 * so it survives late-loading images, and offers to resume on the next visit.
 */
export function readingPosition(node: HTMLElement, options: ReadingPositionOptions) {
	if (!browser) return { destroy() {} };

	const { key, onResumeAvailable } = options;

	const nodeTop = () => node.getBoundingClientRect().top + window.scrollY;
	const currentRatio = () => {
		const line = window.scrollY + window.innerHeight * READING_LINE;
		const height = node.offsetHeight || 1;
		return Math.min(1, Math.max(0, (line - nodeTop()) / height));
	};

	let saveTimer = 0;
	const onScroll = () => {
		clearTimeout(saveTimer);
		saveTimer = window.setTimeout(() => save(key, currentRatio()), 400);
	};

	// Let layout settle (fonts, first images) before measuring or offering.
	const offerTimer = window.setTimeout(() => {
		const saved = load(key);
		if (!saved || Date.now() - saved.savedAt > MAX_AGE_MS) return;
		if (saved.ratio < MIN_RATIO || saved.ratio >= FINISHED_RATIO) return;
		// Respect deep links and readers who already started scrolling.
		if (window.location.hash || currentRatio() > MIN_RATIO) return;
		onResumeAvailable?.(saved.ratio, () => {
			const top = nodeTop() + saved.ratio * node.offsetHeight - window.innerHeight * READING_LINE;
			const reduced = window.matchMedia('(prefers-reduced-motion: reduce)').matches;
			window.scrollTo({ top: Math.max(0, top), behavior: reduced ? 'auto' : 'smooth' });
		});
	}, 600);

	window.addEventListener('scroll', onScroll, { passive: true });

	return {
		destroy() {
			clearTimeout(offerTimer);
			clearTimeout(saveTimer);
			window.removeEventListener('scroll', onScroll);
		}
	};
}
