import { createModelDataContext } from 'svatoms';
import type { ThinkingListResponse } from './types';

export const thinkingListCtx = createModelDataContext<ThinkingListResponse>({
    name: 'thinkingListCtx',
    initial: { items: [], total: 0 }
});
