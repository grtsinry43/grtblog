import type { UserInfo } from '$lib/shared/types/user';

export type OAuthFlowMode = 'login' | 'bind';

type OAuthFlowMeta = {
	mode: OAuthFlowMode;
	provider: string;
	returnTo: string;
	createdAt: number;
};

export type OAuthPopupResult = {
	type: 'oauth:popup-result';
	provider: string;
	mode: OAuthFlowMode;
	success: boolean;
	error?: string;
	token?: string;
	user?: UserInfo;
	returnTo?: string;
};

const OAUTH_META_PREFIX = 'oauth:flow:';

function buildMetaKey(state: string) {
	return `${OAUTH_META_PREFIX}${state}`;
}

export function saveOAuthFlowMeta(state: string, meta: OAuthFlowMeta) {
	if (!state) return;
	localStorage.setItem(buildMetaKey(state), JSON.stringify(meta));
}

export function consumeOAuthFlowMeta(state: string): OAuthFlowMeta | null {
	if (!state) return null;
	const key = buildMetaKey(state);
	const raw = localStorage.getItem(key);
	localStorage.removeItem(key);
	if (!raw) return null;
	try {
		const parsed = JSON.parse(raw) as OAuthFlowMeta;
		if (!parsed || !parsed.mode || !parsed.provider) return null;
		return parsed;
	} catch {
		return null;
	}
}

export function openOAuthPopup(authUrl: string, provider: string): Window | null {
	const width = 560;
	const height = 720;
	const left = Math.max(0, Math.floor(window.screenX + (window.outerWidth - width) / 2));
	const top = Math.max(0, Math.floor(window.screenY + (window.outerHeight - height) / 2));
	const features = [
		'popup=yes',
		`width=${width}`,
		`height=${height}`,
		`left=${left}`,
		`top=${top}`,
		'resizable=yes',
		'scrollbars=yes'
	].join(',');
	return window.open(authUrl, `oauth-${provider}`, features);
}

export function waitForOAuthPopupResult(params: {
	provider: string;
	mode: OAuthFlowMode;
	popup: Window;
	timeoutMs?: number;
}): Promise<OAuthPopupResult> {
	const { provider, mode, popup, timeoutMs = 180_000 } = params;
	return new Promise((resolve, reject) => {
		let finished = false;

		const done = (fn: () => void) => {
			if (finished) return;
			finished = true;
			window.removeEventListener('message', onMessage);
			clearInterval(closedTimer);
			clearTimeout(timeoutTimer);
			fn();
		};

		const onMessage = (event: MessageEvent<OAuthPopupResult>) => {
			if (event.origin !== window.location.origin) return;
			const payload = event.data;
			if (!payload || payload.type !== 'oauth:popup-result') return;
			if (payload.provider !== provider || payload.mode !== mode) return;
			done(() => resolve(payload));
		};

		const closedTimer = window.setInterval(() => {
			if (popup.closed) {
				done(() => reject(new Error('授权窗口已关闭')));
			}
		}, 250);

		const timeoutTimer = window.setTimeout(() => {
			done(() => reject(new Error('授权超时，请重试')));
		}, timeoutMs);

		window.addEventListener('message', onMessage);
	});
}
