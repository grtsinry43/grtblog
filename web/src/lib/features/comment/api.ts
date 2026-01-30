import { getApi } from '$lib/shared/clients/api';
import type {
	CommentCreateResponse,
	CommentNode,
	CreateCommentLoginPayload,
	CreateCommentVisitorPayload
} from '$lib/features/comment/types';

export const getCommentTree = async (
	fetcher: typeof fetch | undefined,
	areaId: number
): Promise<CommentNode[]> => {
	const api = getApi(fetcher);
	const result = await api<CommentNode[]>(`/comments/areas/${areaId}`);
	return result ?? [];
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
