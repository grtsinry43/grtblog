import { getApi } from '$lib/shared/clients/api';
import { browser } from '$app/environment';
import { ofetch, FetchError } from 'ofetch';
import type {
	CommentCreateResponse,
	CommentListResponse,
	CreateCommentLoginPayload,
	CreateCommentVisitorPayload
} from '$lib/features/comment/types';
import type { ApiResponse } from '$lib/shared/clients/types';
import { BusinessError } from '$lib/shared/clients/types';

const isAuthError = (error: unknown): boolean => {
	if (error instanceof BusinessError) {
		return error.code === 401 || error.code === 40101 || error.code === 403;
	}
	if (error instanceof FetchError) {
		const status = error.response?.status ?? 0;
		return status === 401 || status === 403;
	}
	return false;
};

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
	const path = `/comments/areas/${areaId}?${query.toString()}`;
	try {
		return await api<CommentListResponse>(path);
	} catch (error) {
		// Fallback for cases where auth headers are rejected upstream after login.
		if (!browser || fetcher || !isAuthError(error)) {
			throw error;
		}
		const envelope = await ofetch<ApiResponse<CommentListResponse>>(`/api/v2${path}`);
		if (typeof envelope?.code === 'number' && envelope.code === 0) {
			return envelope.data;
		}
		throw error;
	}
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
