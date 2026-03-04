import { createModelDataContext } from 'svatoms';
import type { CommentNode } from '$lib/features/comment/types';

export type CommentAreaModel = {
	areaId: number;
	comments: CommentNode[];
	isLoading: boolean;
	isError: boolean;
	replyingTo: CommentNode | null;
	editingComment: CommentNode | null;
	isLoggedIn: boolean;
	guestName: string;
	guestEmail: string;
	guestSite: string;
	commentsCount: number;
	total: number;
	page: number;
	size: number;
	isClosed: boolean;
	requireModeration: boolean;
};

export const commentAreaCtx = createModelDataContext<CommentAreaModel>({
	name: 'commentAreaCtx',
	initial: {
		areaId: 0,
		comments: [],
		isLoading: false,
		isError: false,
		replyingTo: null,
		editingComment: null,
		isLoggedIn: false,
		guestName: '',
		guestEmail: '',
		guestSite: '',
		commentsCount: 0,
		total: 0,
		page: 1,
		size: 10,
		isClosed: false,
		requireModeration: false
	}
});
