import { getMomentListByColumn } from '$lib/features/moment/api';
import { getColumns } from '$lib/features/taxonomy/api';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async (event) => {
	const { fetch, params, url } = event;
	const slug = params.slug;
	const rawPage = Number(params.page ?? '1');
	const page = Number.isFinite(rawPage) && rawPage > 0 ? rawPage : 1;

	const rawPageSize = Number(url.searchParams.get('pageSize') ?? '20');
	const pageSize = Number.isFinite(rawPageSize) && rawPageSize > 0 ? rawPageSize : 20;

	const [data, columns] = await Promise.all([
		getMomentListByColumn(fetch, slug, { page, pageSize }),
		getColumns(fetch)
	]);

	const column = columns.find((c) => c.shortUrl === slug);
	const columnName = column?.name ?? slug;

	return {
		columnSlug: slug,
		columnName,
		moments: data
	};
};
