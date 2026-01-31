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
