import { getApi } from '$lib/shared/clients/api';
import type { FriendLink, FriendApplyForm } from './types';

/**
 * 获取公开友邻列表 - 用于 SvelteKit load 函数 (强 SEO)
 */
export const getFriendLinks = async (fetcher?: typeof fetch): Promise<FriendLink[]> => {
	const api = getApi(fetcher);
	return api<FriendLink[]>('/public/friend-links');
};

/**
 * 提交友链申请 - 客户端异步操作
 */
export const applyFriendLink = async (form: FriendApplyForm): Promise<void> => {
	const api = getApi(); // 客户端 fetcher 留空
	return api<void>('/friend-links/applications', {
		method: 'POST',
		body: form
	});
};
