import { createModelDataContext } from 'svatoms';
import type { MomentListResponse } from '$lib/features/moment/types';

export const momentContext = createModelDataContext<MomentListResponse>({
    name: 'momentContext',
    initial: null
});
