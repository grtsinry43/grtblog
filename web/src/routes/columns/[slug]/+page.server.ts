import { getMomentListByColumn } from '$lib/features/moment/api';
import { getColumns } from '$lib/features/taxonomy/api';
import { trackISRDeps } from '$lib/server/isr-deps';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async (event) => {
	const { fetch, params, url } = event;
	const slug = params.slug;

	const rawPageSize = Number(url.searchParams.get('pageSize') ?? '20');
	const pageSize = Number.isFinite(rawPageSize) && rawPageSize > 0 ? rawPageSize : 20;

	const [data, columns] = await Promise.all([
		getMomentListByColumn(fetch, slug, { page: 1, pageSize }),
		getColumns(fetch)
	]);

	const column = columns.find((c) => c.shortUrl === slug);
	const columnName = column?.name ?? slug;
	trackISRDeps(event, 'column:list');

	return {
		columnSlug: slug,
		columnName,
		moments: data
	};
};
