import { createModelDataContext } from 'svatoms';
import type { MomentListResponse } from '$lib/features/moment/types';

export const momentListCtx = createModelDataContext<MomentListResponse>({
	name: 'momentListCtx',
	initial: null
});
