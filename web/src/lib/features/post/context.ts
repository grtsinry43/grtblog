import { createModelDataContext } from 'svatoms';
import type { NavMenuItem } from '$lib/features/navigation/types';
import type { PostSummary, PostDetail } from '$lib/features/post/types';

// Post list page context
export type PostListPageData = {
	navMenus: NavMenuItem[];
	posts: PostSummary[];
	pagination: {
		total: number;
		page: number;
		size: number;
	};
};

export const postListCtx = createModelDataContext<PostListPageData>({
	name: 'postListCtx',
	initial: null
});

// Post detail page context
export const postDetailCtx = createModelDataContext<PostDetail | null>({
	name: 'postDetailCtx',
	initial: null
});
