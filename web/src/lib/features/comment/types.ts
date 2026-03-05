export type CreateCommentLoginPayload = {
	content: string;
	parentId?: number | null;
	visitorId?: string;
};

export type CreateCommentVisitorPayload = {
	content: string;
	nickName: string;
	email: string;
	website?: string | null;
	parentId?: number | null;
	visitorId?: string;
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
	avatar?: string | null;
	isOwner: boolean;
	isFriend: boolean;
	isAuthor: boolean;
	isViewed: boolean;
	isTop: boolean;
	isMy: boolean;
	isFederated?: boolean;
	federatedProtocol?: string | null;
	federatedActor?: string | null;
	canReply: boolean;
	status: string;
	isEdited: boolean;
	parentId?: number | null;
	createdAt: string;
	updatedAt: string;
	deletedAt?: string | null;
	isDeleted: boolean;
};

export type CommentNode = {
	id: number;
	areaId: number;
	floor?: string;
	content: string | null;
	nickName?: string | null;
	location?: string | null;
	platform?: string | null;
	browser?: string | null;
	website?: string | null;
	avatar?: string | null;
	isOwner: boolean;
	isFriend: boolean;
	isAuthor: boolean;
	isViewed: boolean;
	isTop: boolean;
	isMy: boolean;
	isFederated?: boolean;
	federatedProtocol?: string | null;
	federatedActor?: string | null;
	canReply: boolean;
	status: string;
	isEdited: boolean;
	parentId?: number | null;
	createdAt: string;
	updatedAt: string;
	deletedAt?: string | null;
	isDeleted: boolean;
	children?: CommentNode[];
};

export type CommentListResponse = {
	items: CommentNode[];
	total: number;
	page: number;
	size: number;
	isClosed: boolean;
	requireModeration: boolean;
};
