import { getApi } from '$lib/shared/clients/api';
import type { Category, Column } from './types';

export const getCategories = async (fetcher?: typeof fetch): Promise<Category[]> => {
	const api = getApi(fetcher);
	const result = await api<Category[]>('/categories');
	return result ?? [];
};

export const getColumns = async (fetcher?: typeof fetch): Promise<Column[]> => {
	const api = getApi(fetcher);
	const result = await api<Column[]>('/columns');
	return result ?? [];
};
