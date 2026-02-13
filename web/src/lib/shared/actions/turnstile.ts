import { browser } from '$app/environment';

type TurnstileRenderOptions = {
	siteKey?: string;
	theme?: 'light' | 'dark' | 'auto';
	size?: 'normal' | 'compact';
	action?: string;
	onToken?: (token: string) => void;
	onExpired?: () => void;
	onError?: (error: unknown) => void;
};

type TurnstileApi = {
	render: (container: HTMLElement, options: Record<string, unknown>) => string;
	remove: (widgetId: string) => void;
};

declare global {
	interface Window {
		turnstile?: TurnstileApi;
	}
}

let turnstileScriptPromise: Promise<void> | null = null;

const loadTurnstileScript = () => {
	if (!browser) return Promise.resolve();
	if (window.turnstile && typeof window.turnstile.render === 'function') return Promise.resolve();
	if (turnstileScriptPromise) return turnstileScriptPromise;

	turnstileScriptPromise = new Promise((resolve, reject) => {
		const script = document.createElement('script');
		script.src = 'https://challenges.cloudflare.com/turnstile/v0/api.js?render=explicit';
		script.async = true;
		script.defer = true;
		script.onload = () => resolve();
		script.onerror = () => reject(new Error('Turnstile 脚本加载失败'));
		document.head.appendChild(script);
	});

	return turnstileScriptPromise;
};

export const preloadTurnstile = () => loadTurnstileScript();

const waitForTurnstile = (timeoutMs = 3000) =>
	new Promise<void>((resolve, reject) => {
		if (window.turnstile && typeof window.turnstile.render === 'function') {
			resolve();
			return;
		}
		const start = Date.now();
		const timer = window.setInterval(() => {
			if (window.turnstile && typeof window.turnstile.render === 'function') {
				window.clearInterval(timer);
				resolve();
				return;
			}
			if (Date.now() - start > timeoutMs) {
				window.clearInterval(timer);
				reject(new Error('Turnstile API 未就绪'));
			}
		}, 50);
	});

export const turnstileWidget = (node: HTMLElement, options: TurnstileRenderOptions = {}) => {
	if (!browser) return {};

	let widgetId: string | null = null;
	let currentOptions = options;
	let destroyed = false;
	let renderVersion = 0;

	const cleanup = () => {
		if (widgetId && window.turnstile) {
			window.turnstile.remove(widgetId);
		}
		widgetId = null;
		node.innerHTML = '';
	};

	const render = async () => {
		const version = ++renderVersion;
		if (!currentOptions.siteKey) {
			cleanup();
			return;
		}
		try {
			await loadTurnstileScript();
			await waitForTurnstile();
			if (destroyed || !window.turnstile || typeof window.turnstile.render !== 'function') {
				currentOptions.onError?.(new Error('Turnstile API 未就绪'));
				return;
			}

			if (destroyed || version !== renderVersion || !window.turnstile) return;
			cleanup();
			widgetId = window.turnstile.render(node, {
				sitekey: currentOptions.siteKey,
				theme: currentOptions.theme ?? 'auto',
				size: currentOptions.size ?? 'normal',
				action: currentOptions.action,
				callback: (token: string) => currentOptions.onToken?.(token),
				'expired-callback': () => currentOptions.onExpired?.(),
				'error-callback': () => currentOptions.onError?.(new Error('Turnstile 验证失败'))
			});
		} catch (error) {
			currentOptions.onError?.(error);
		}
	};

	void render();

	return {
		update(next?: TurnstileRenderOptions) {
			currentOptions = next ?? {};
			void render();
		},
		destroy() {
			destroyed = true;
			cleanup();
		}
	};
};
