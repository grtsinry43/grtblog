// --- Absolute formatters ---

/** Format as "2024年3月15日" (Chinese locale date). */
export const formatDateCN = (value?: string): string => {
	if (!value) return '';
	const d = new Date(value);
	if (Number.isNaN(d.getTime())) return value;
	return `${d.getFullYear()}年${d.getMonth() + 1}月${d.getDate()}日`;
};

/** Format as "2024.03.15" (dotted date). */
export const formatDateDotted = (value?: string): string => {
	if (!value) return '';
	const d = new Date(value);
	if (Number.isNaN(d.getTime())) return value;
	return `${d.getFullYear()}.${String(d.getMonth() + 1).padStart(2, '0')}.${String(d.getDate()).padStart(2, '0')}`;
};

/** Format as "0315" (MMDD compact). */
export const formatDateCompact = (value?: string): string => {
	if (!value) return '';
	const d = new Date(value);
	if (Number.isNaN(d.getTime())) return '';
	return `${String(d.getMonth() + 1).padStart(2, '0')}${String(d.getDate()).padStart(2, '0')}`;
};

/** Get Chinese season name from a date string. */
export const getSeason = (value?: string): string => {
	if (!value) return '';
	const month = new Date(value).getMonth() + 1;
	if (month >= 3 && month <= 5) return '春';
	if (month >= 6 && month <= 8) return '夏';
	if (month >= 9 && month <= 11) return '秋';
	return '冬';
};

/** Check if two date strings represent different days. */
export const isDifferentDay = (a?: string, b?: string): boolean => {
	if (!a || !b) return false;
	const da = new Date(a);
	const db = new Date(b);
	if (Number.isNaN(da.getTime()) || Number.isNaN(db.getTime())) return false;
	return da.getFullYear() !== db.getFullYear() || da.getMonth() !== db.getMonth() || da.getDate() !== db.getDate();
};

// --- Relative formatters ---

export const formatRelativeTime = (dateStr: string): string => {
	const date = new Date(dateStr);
	const now = new Date();
	const diff = now.getTime() - date.getTime();

	const seconds = Math.floor(diff / 1000);
	const minutes = Math.floor(seconds / 60);
	const hours = Math.floor(minutes / 60);
	const days = Math.floor(hours / 24);

	if (days < 1) {
		if (hours < 1) {
			if (minutes < 1) return '刚刚';
			return `${minutes} 分钟前`;
		}
		return `${hours} 小时前`;
	}

	if (days < 7) return `${days} 天前`;
	if (days < 30) return `大约 ${Math.ceil(days / 7)} 周前`;
	if (days < 365) return `大约 ${Math.floor(days / 30)} 个月前`;

	return `${date.getFullYear()}年`;
};

export const formatRelativeTimeWithSeconds = (dateStr: string, now = new Date()): string => {
	const date = new Date(dateStr);
	const diffMs = now.getTime() - date.getTime();

	const clampedMs = Math.max(diffMs, 0);
	const seconds = Math.floor(clampedMs / 1000);
	const minutes = Math.floor(seconds / 60);
	const hours = Math.floor(minutes / 60);
	const days = Math.floor(hours / 24);

	if (days < 1) {
		if (hours < 1) {
			if (minutes < 1) return seconds <= 0 ? '刚刚' : `${seconds} 秒前`;
			return `${minutes} 分钟前`;
		}
		return `${hours} 小时前`;
	}

	if (days < 7) return `${days} 天前`;
	if (days < 30) return `大约 ${Math.ceil(days / 7)} 周前`;
	if (days < 365) return `大约 ${Math.floor(days / 30)} 个月前`;

	return `${date.getFullYear()}年`;
};

const getNextDelay = (diffMs: number): number | null => {
	if (diffMs < 60_000) return 1_000;
	if (diffMs < 3_600_000) return 60_000;
	return null;
};

export const createRelativeTimeTicker = (
	dateStr: string,
	onTick: (value: string) => void
): (() => void) => {
	if (typeof window === 'undefined') return () => {};

	let timeoutId: ReturnType<typeof setTimeout> | null = null;

	const tick = () => {
		const now = new Date();
		const diffMs = now.getTime() - new Date(dateStr).getTime();
		onTick(formatRelativeTimeWithSeconds(dateStr, now));
		const delay = getNextDelay(Math.max(diffMs, 0));
		if (delay !== null) {
			timeoutId = setTimeout(tick, delay);
		}
	};

	tick();

	return () => {
		if (timeoutId !== null) {
			clearTimeout(timeoutId);
		}
	};
};
