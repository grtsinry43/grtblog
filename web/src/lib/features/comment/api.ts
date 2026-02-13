import { getApi } from '$lib/shared/clients/api';
import type {
	CommentCreateResponse,
	CommentListResponse,
	CreateCommentLoginPayload,
	CreateCommentVisitorPayload
} from '$lib/features/comment/types';

export const getCommentTree = async (
	fetcher: typeof fetch | undefined,
	areaId: number,
	page = 1,
	size = 10,
	visitorId?: string
): Promise<CommentListResponse | null> => {
	const api = getApi(fetcher);
	const query = new URLSearchParams({
		page: String(page),
		size: String(size)
	});
	if (visitorId?.trim()) {
		query.set('visitorId', visitorId.trim());
	}
	return api<CommentListResponse>(`/comments/areas/${areaId}?${query.toString()}`);
};

export const createCommentLogin = async (
	fetcher: typeof fetch | undefined,
	areaId: number,
	payload: CreateCommentLoginPayload
): Promise<CommentCreateResponse | null> => {
	const api = getApi(fetcher);
	const result = await api<CommentCreateResponse>(`/comments/areas/${areaId}`, {
		method: 'POST',
		body: payload
	});
	return result ?? null;
};

export const createCommentVisitor = async (
	fetcher: typeof fetch | undefined,
	areaId: number,
	payload: CreateCommentVisitorPayload
): Promise<CommentCreateResponse | null> => {
	const api = getApi(fetcher);
	const result = await api<CommentCreateResponse>(`/comments/areas/${areaId}/visitor`, {
		method: 'POST',
		body: payload
	});
	return result ?? null;
};
