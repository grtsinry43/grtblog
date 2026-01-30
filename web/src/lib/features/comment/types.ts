export type CreateCommentLoginPayload = {
	content: string;
	parentId?: number | null;
};

export type CreateCommentVisitorPayload = {
	content: string;
	nickName: string;
	email: string;
	website?: string | null;
	parentId?: number | null;
};

export type CommentCreateResponse = {
	id: number;
	areaId: number;
	content: string;
	nickName?: string | null;
	location?: string | null;
	platform?: string | null;
	browser?: string | null;
	website?: string | null;
	isOwner: boolean;
	isFriend: boolean;
	isAuthor: boolean;
	isViewed: boolean;
	isTop: boolean;
	parentId?: number | null;
	createdAt: string;
	updatedAt: string;
	deletedAt?: string | null;
};

export type CommentNode = {
	id: number;
	areaId: number;
	content: string;
	nickName?: string | null;
	location?: string | null;
	platform?: string | null;
	browser?: string | null;
	website?: string | null;
	isOwner: boolean;
	isFriend: boolean;
	isAuthor: boolean;
	isViewed: boolean;
	isTop: boolean;
	parentId?: number | null;
	createdAt: string;
	updatedAt: string;
	deletedAt?: string | null;
	children?: CommentNode[];
};
