import type { PostDetail, PostRelatedMoment } from '$lib/features/post/types';
import type { TOCNode } from '$lib/shared/types/toc';

type Metrics = PostDetail['metrics'] | null | undefined;

export const sameMetrics = (a: Metrics, b: Metrics): boolean => {
	if (!a || !b) return a === b;
	return a.views === b.views && a.likes === b.likes && a.comments === b.comments;
};

const sameTocNode = (a: TOCNode, b: TOCNode): boolean => {
	if (a.anchor !== b.anchor || a.name !== b.name) return false;
	return sameToc(a.children, b.children);
};

export const sameToc = (
	a: TOCNode[] | null | undefined,
	b: TOCNode[] | null | undefined
): boolean => {
	if (a === b) return true;
	if (!a?.length && !b?.length) return true;
	if (!a || !b || a.length !== b.length) return false;

	for (let i = 0; i < a.length; i += 1) {
		if (!sameTocNode(a[i], b[i])) return false;
	}

	return true;
};

const sameStringArray = (a: string[] | undefined, b: string[] | undefined): boolean => {
	if (a === b) return true;
	if (!a?.length && !b?.length) return true;
	if (!a || !b || a.length !== b.length) return false;

	for (let i = 0; i < a.length; i += 1) {
		if (a[i] !== b[i]) return false;
	}

	return true;
};

const sameRelatedMoment = (a: PostRelatedMoment, b: PostRelatedMoment): boolean => {
	return (
		a.id === b.id &&
		a.title === b.title &&
		a.shortUrl === b.shortUrl &&
		a.summary === b.summary &&
		a.createdAt === b.createdAt &&
		sameStringArray(a.image, b.image)
	);
};

export const samePostRelatedMoments = (
	a: PostRelatedMoment[] | null | undefined,
	b: PostRelatedMoment[] | null | undefined
): boolean => {
	if (a === b) return true;
	if (!a?.length && !b?.length) return true;
	if (!a || !b || a.length !== b.length) return false;

	for (let i = 0; i < a.length; i += 1) {
		if (!sameRelatedMoment(a[i], b[i])) return false;
	}

	return true;
};
