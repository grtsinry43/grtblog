import { createModelDataContext } from 'svatoms';

export interface ImageExtInfoItem {
	id: string;
	width?: number;
	height?: number;
	color?: string;
}

export interface ContentExtInfo {
	images?: ImageExtInfoItem[];
}

export type ImageExtInfoState = {
	images: ImageExtInfoItem[];
	map: Map<string, ImageExtInfoItem>;
};

export const buildImageExtInfoState = (extInfo?: ContentExtInfo | null): ImageExtInfoState | null => {
	if (!extInfo?.images?.length) return null;
	const images = extInfo.images.filter((item) => item?.id);
	if (!images.length) return null;
	const map = new Map<string, ImageExtInfoItem>();
	for (const item of images) {
		map.set(item.id, item);
	}
	return { images, map };
};

export const imageExtInfoCtx = createModelDataContext<ImageExtInfoState | null>({
	name: 'imageExtInfoCtx',
	initial: null
});
