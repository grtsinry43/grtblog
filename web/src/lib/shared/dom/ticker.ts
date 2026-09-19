import { browser } from '$app/environment';

/**
 * 创建一个 SSR 安全的定时器能力。
 *
 * - 仅在浏览器环境生效，SSR 阶段调用是 no-op。
 * - 返回的 `stop` 必须接入组件 cleanup（如 `$effect` 的返回值）。
 * - 多次调用 `start` 不会产生重复定时器。
 */
export function createTicker(callback: () => void, intervalMs: number) {
	let timer: ReturnType<typeof setInterval> | null = null;

	const start = () => {
		if (!browser || timer !== null) return;
		timer = setInterval(callback, intervalMs);
	};

	const stop = () => {
		if (timer !== null) {
			clearInterval(timer);
			timer = null;
		}
	};

	return { start, stop };
}
