import { createModelDataContext } from 'svatoms';
import type { MomentDetail, MomentListResponse } from '$lib/features/moment/types';

export const momentListCtx = createModelDataContext<MomentListResponse>({
	name: 'momentListCtx',
	initial: null
});

export const momentDetailCtx = createModelDataContext<MomentDetail | null>({
	name: 'momentDetailCtx',
	initial: null
});
