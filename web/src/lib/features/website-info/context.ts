import { createModelDataContext } from 'svatoms';

import type { WebsiteInfoMap } from './types';

export const websiteInfoCtx = createModelDataContext<WebsiteInfoMap | null>({
	name: 'websiteInfoCtx',
	initial: null
});
