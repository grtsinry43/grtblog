import { browser } from '$app/environment';

const STORAGE_KEY = 'grtblog:analytics:visitor-id:v1';
const PREFIX = 'v2';

const normalizeVisitorId = (value: string | null | undefined): string | null => {
	if (!value) return null;
	const next = value.trim();
	if (!next) return null;
	return next.slice(0, 255);
};

const randomHex = (bytes: number): string => {
	const array = new Uint8Array(bytes);
	crypto.getRandomValues(array);
	return Array.from(array, (byte) => byte.toString(16).padStart(2, '0')).join('');
};

export const createVisitorId = (): string => {
	if (!browser) return '';
	if (typeof crypto.randomUUID === 'function') {
		return `${PREFIX}_${crypto.randomUUID().replaceAll('-', '')}`;
	}
	return `${PREFIX}_${randomHex(16)}`;
};

export const getVisitorId = (): string | null => {
	if (!browser) return null;
	return normalizeVisitorId(localStorage.getItem(STORAGE_KEY));
};

export const setVisitorId = (visitorId: string | null | undefined): string | null => {
	if (!browser) return null;
	const normalized = normalizeVisitorId(visitorId);
	if (!normalized) {
		localStorage.removeItem(STORAGE_KEY);
		return null;
	}
	localStorage.setItem(STORAGE_KEY, normalized);
	return normalized;
};

export const getOrCreateVisitorId = (): string => {
	if (!browser) return '';
	const cached = getVisitorId();
	if (cached) return cached;
	const next = createVisitorId();
	setVisitorId(next);
	return next;
};

export const syncVisitorId = (visitorId: string | null | undefined): string | null => {
	if (!browser) return null;
	return setVisitorId(visitorId);
};
