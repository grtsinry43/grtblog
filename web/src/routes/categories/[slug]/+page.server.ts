import { getPostListByCategory } from '$lib/features/post/api';
import { getCategories } from '$lib/features/taxonomy/api';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async (event) => {
	const { fetch, params, url } = event;
	const slug = params.slug;

	const rawPageSize = Number(url.searchParams.get('pageSize') ?? '10');
	const pageSize = Number.isFinite(rawPageSize) && rawPageSize > 0 ? rawPageSize : 10;

	const [data, categories] = await Promise.all([
		getPostListByCategory(fetch, slug, { page: 1, pageSize }),
		getCategories(fetch)
	]);

	const category = categories.find((c) => c.shortUrl === slug);
	const categoryName = category?.name ?? slug;

	return {
		categorySlug: slug,
		categoryName,
		posts: data.items,
		pagination: { total: data.total, page: data.page, size: data.size }
	};
};
