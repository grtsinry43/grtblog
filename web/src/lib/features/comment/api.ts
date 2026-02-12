import { getApi } from '$lib/shared/clients/api';
import type {
	CommentCreateResponse,
	CommentNode,
	CommentListResponse,
	CreateCommentLoginPayload,
	CreateCommentVisitorPayload
} from '$lib/features/comment/types';

export const getCommentTree = async (
	fetcher: typeof fetch | undefined,
	areaId: number,
	page = 1,
	size = 10
): Promise<CommentListResponse | null> => {
	const api = getApi(fetcher);
	return api<CommentListResponse>(`/comments/areas/${areaId}?page=${page}&size=${size}`);
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
