import { createModelDataContext } from 'svatoms';
import type { FriendTimelineListResponse } from './types';

export const friendTimelineListCtx = createModelDataContext<FriendTimelineListResponse>({
	name: 'friendTimelineListCtx',
	initial: null
});
