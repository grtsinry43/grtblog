import { ofetch } from 'ofetch';
import type { FetchOptions } from 'ofetch';
import { browser } from '$app/environment';
import { getToken } from '$lib/shared/token';
import { type ApiResponse, BusinessError } from '$lib/shared/clients/types';
import { authModalStore } from '$lib/shared/stores/authModalStore';

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

export const createServerApi = (svelteFetch: typeof fetch) => {
	return ofetch.create({
		...defaults,
		// eslint-disable-next-line @typescript-eslint/ban-ts-comment
		// @ts-expect-error
		// eslint-disable-next-line
		fetch: svelteFetch as any, // 替换底层 fetch 为 SvelteKit 的特供版
		// 如果需要服务端走内网 DNS，可以在这里覆盖 baseURL
		baseURL: 'http://localhost:8080/api/v2'
	});
};

export const getApi = (svelteFetch?: typeof fetch) => {
	return svelteFetch ? createServerApi(svelteFetch) : api;
};
