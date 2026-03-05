import { ofetch, FetchError } from 'ofetch';
import type { FetchOptions } from 'ofetch';
import { browser } from '$app/environment';
import { getToken } from '$lib/shared/token';
import { type ApiResponse, BusinessError } from '$lib/shared/clients/types';
import { authModalStore } from '$lib/shared/stores/authModalStore';

/**
 * 执行 API 调用，当后端返回 404 时返回 null 而非抛出异常。
 * 同时处理 BusinessError（onResponse 抛出）和 FetchError（ofetch 原生错误）两种情况。
 */
export async function fetchOrNull<T>(fn: () => Promise<T>): Promise<T | null> {
	try {
		return await fn();
	} catch (e) {
		if (e instanceof BusinessError && e.code === 404) {
			return null;
		}
		if (e instanceof FetchError && e.response?.status === 404) {
			return null;
		}
		throw e;
	}
}

const defaults: FetchOptions = {
	baseURL: '/api/v2',
	headers: {
		'Content-Type': 'application/json'
	},
	// 响应拦截：统一处理错误
	async onResponseError({ response }) {
		if (response.status === 401 && browser) {
			// 客户端收到 401，打开登录弹窗
			authModalStore.open('unauthorized');
		}

		if (browser && response.status >= 500) {
			console.error('服务器炸了:', response._data);
			// toast.error('服务器开小差了');
		}
	},
	async onResponse({ response }) {
		const res = response._data as ApiResponse<never>;

		if (typeof res?.code !== 'number') {
			return;
		}

		if (res.code === 0) {
			response._data = res.data;
		}

		// 业务错误分支 (code != 0)
		else {
			throw new BusinessError(
				res.code,
				res.msg || '未知错误',
				res.bizErr || '' // 业务调试信息
			);
		}
	},
	async onRequest({ options }) {
		const token = browser ? getToken() : null;
		if (token) {
			options.headers = {
				...options.headers,
				// eslint-disable-next-line @typescript-eslint/ban-ts-comment
				// @ts-expect-error
				Authorization: `Bearer ${token}`
			};
		}
	}
};

export const api = ofetch.create(defaults);

const defaultInternalApiBaseURL = 'http://localhost:8080/api/v2';

function resolveInternalApiBaseURL(): string {
	if (typeof process === 'undefined' || !process.env) {
		return defaultInternalApiBaseURL;
	}
	const raw = (process.env.INTERNAL_API_BASE_URL || '').trim();
	if (!raw) {
		return defaultInternalApiBaseURL;
	}
	if (raw.endsWith('/api/v2')) {
		return raw;
	}
	return `${raw.replace(/\/+$/, '')}/api/v2`;
}

export const createServerApi = (svelteFetch: typeof fetch) => {
	return ofetch.create({
		...defaults,
		// eslint-disable-next-line @typescript-eslint/ban-ts-comment
		// @ts-expect-error
		// eslint-disable-next-line
		fetch: svelteFetch as any, // 替换底层 fetch 为 SvelteKit 的特供版
		// 服务端优先走容器内网地址；默认回退到 localhost 便于本地开发
		baseURL: resolveInternalApiBaseURL()
	});
};

export const getApi = (svelteFetch?: typeof fetch) => {
	return svelteFetch ? createServerApi(svelteFetch) : api;
};
