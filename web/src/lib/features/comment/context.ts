import { createModelDataContext } from 'svatoms';
import type { CommentNode } from '$lib/features/comment/types';

export type CommentAreaModel = {
	areaId: number;
	comments: CommentNode[];
	isLoading: boolean;
	isError: boolean;
	replyingTo: CommentNode | null;
	isLoggedIn: boolean;
	guestName: string;
	guestEmail: string;
	guestSite: string;
	commentsCount: number;
};

export const commentAreaCtx = createModelDataContext<CommentAreaModel>({
	name: 'commentAreaCtx',
	initial: {
		areaId: 0,
		comments: [],
		isLoading: false,
		isError: false,
		replyingTo: null,
		isLoggedIn: false,
		guestName: '',
		guestEmail: '',
		guestSite: '',
		commentsCount: 0
	}
});
