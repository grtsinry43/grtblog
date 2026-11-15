import { browser } from '$app/environment';

export type ReadingSize = 'sm' | 'md' | 'lg';
export type ReadingFont = 'sans' | 'serif';

export type ReadingPrefs = {
	size: ReadingSize;
	font: ReadingFont;
};

/** Shared with the inline boot script in app.html — keep the key in sync. */
export const READING_PREFS_KEY = 'reading-prefs';

const DEFAULT_PREFS: ReadingPrefs = { size: 'md', font: 'sans' };

function readPrefs(): ReadingPrefs {
	if (!browser) return DEFAULT_PREFS;
	try {
		const raw = JSON.parse(localStorage.getItem(READING_PREFS_KEY) ?? '{}');
		return {
			size: ['sm', 'md', 'lg'].includes(raw.size) ? raw.size : DEFAULT_PREFS.size,
			font: ['sans', 'serif'].includes(raw.font) ? raw.font : DEFAULT_PREFS.font
		};
	} catch {
		return DEFAULT_PREFS;
	}
}

function applyPrefs(prefs: ReadingPrefs) {
	const root = document.documentElement;
	root.dataset.readingSize = prefs.size;
	root.dataset.readingFont = prefs.font;
}

function createReadingPrefs() {
	let prefs = $state<ReadingPrefs>(readPrefs());

	function update(patch: Partial<ReadingPrefs>) {
		prefs = { ...prefs, ...patch };
		if (!browser) return;
		applyPrefs(prefs);
		try {
			localStorage.setItem(READING_PREFS_KEY, JSON.stringify(prefs));
		} catch {
			// storage unavailable (private mode) — prefs still apply for this session
		}
	}

	return {
		get size() {
			return prefs.size;
		},
		get font() {
			return prefs.font;
		},
		update
	};
}

export const readingPrefs = createReadingPrefs();
