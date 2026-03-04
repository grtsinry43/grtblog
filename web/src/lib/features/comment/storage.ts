import { browser } from '$app/environment';

const GUEST_PROFILE_KEY = 'grtblog:comment:guest-profile:v1';
const DRAFT_KEY_PREFIX = 'grtblog:comment:draft:v1';

export type CommentGuestProfile = {
	guestName: string;
	guestEmail: string;
	guestSite: string;
};

const parseJson = <T>(raw: string | null): T | null => {
	if (!raw) return null;
	try {
		return JSON.parse(raw) as T;
	} catch {
		return null;
	}
};

export const readCommentGuestProfile = (): CommentGuestProfile | null => {
	if (!browser) return null;
	const parsed = parseJson<Partial<CommentGuestProfile>>(localStorage.getItem(GUEST_PROFILE_KEY));
	if (!parsed) return null;
	return {
		guestName: typeof parsed.guestName === 'string' ? parsed.guestName : '',
		guestEmail: typeof parsed.guestEmail === 'string' ? parsed.guestEmail : '',
		guestSite: typeof parsed.guestSite === 'string' ? parsed.guestSite : ''
	};
};

export const writeCommentGuestProfile = (profile: Partial<CommentGuestProfile>) => {
	if (!browser) return;
	const next: CommentGuestProfile = {
		guestName: typeof profile.guestName === 'string' ? profile.guestName : '',
		guestEmail: typeof profile.guestEmail === 'string' ? profile.guestEmail : '',
		guestSite: typeof profile.guestSite === 'string' ? profile.guestSite : ''
	};
	if (!next.guestName && !next.guestEmail && !next.guestSite) {
		localStorage.removeItem(GUEST_PROFILE_KEY);
		return;
	}
	localStorage.setItem(GUEST_PROFILE_KEY, JSON.stringify(next));
};

export const buildCommentDraftKey = (areaId: number, parentId?: number): string =>
	`${DRAFT_KEY_PREFIX}:${areaId}:${parentId ?? 0}`;

export const readCommentDraft = (key: string): string => {
	if (!browser || !key) return '';
	const raw = localStorage.getItem(key);
	return typeof raw === 'string' ? raw : '';
};

export const writeCommentDraft = (key: string, value: string) => {
	if (!browser || !key) return;
	if (!value) {
		localStorage.removeItem(key);
		return;
	}
	localStorage.setItem(key, value);
};

export const clearCommentDraft = (key: string) => {
	if (!browser || !key) return;
	localStorage.removeItem(key);
};
