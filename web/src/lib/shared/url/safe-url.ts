/**
 * URL scheme guards for externally sourced links (federation citations,
 * friend links, remote timeline posts). Blocks `javascript:` and other
 * non-http(s) schemes while keeping relative and protocol-relative URLs.
 */
const hasScheme = (value: string) => /^[a-zA-Z][a-zA-Z\d+\-.]*:/.test(value);

export const isSafeHttpUrl = (value: string | null | undefined): boolean => {
	const raw = (value ?? '').trim();
	if (!raw) return false;
	if (raw.startsWith('//')) return true;
	if (hasScheme(raw)) {
		return /^https?:/i.test(raw);
	}
	return true;
};

/** Returns the trimmed URL when safe, otherwise the fallback. */
export const safeHttpUrl = (value: string | null | undefined, fallback = ''): string => {
	const raw = (value ?? '').trim();
	return isSafeHttpUrl(raw) ? raw : fallback;
};
