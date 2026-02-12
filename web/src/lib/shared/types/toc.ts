export type TOCNode = {
	name: string;
	anchor: string;
	children?: TOCNode[];
};

/** Flatten a TOCNode tree into a flat list of anchor strings. */
export const flattenTOC = (nodes?: TOCNode[]): string[] => {
	if (!nodes?.length) return [];
	const anchors: string[] = [];
	const walk = (items: TOCNode[]) => {
		for (const item of items) {
			anchors.push(item.anchor);
			if (item.children?.length) walk(item.children);
		}
	};
	walk(nodes);
	return anchors;
};
