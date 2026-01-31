type ImageInteractionHandlers = {
	onClick?: () => void;
	onLoad?: () => void;
};

export const bindImageInteractions = (
	node: HTMLImageElement,
	handlers: ImageInteractionHandlers = {}
) => {
	if (typeof window === 'undefined') return () => {};

	const handleClick = () => handlers.onClick?.();
	const handleLoad = () => handlers.onLoad?.();

	if (handlers.onClick) {
		node.addEventListener('click', handleClick);
	}
	if (handlers.onLoad) {
		node.addEventListener('load', handleLoad);
	}

	return () => {
		if (handlers.onClick) node.removeEventListener('click', handleClick);
		if (handlers.onLoad) node.removeEventListener('load', handleLoad);
	};
};
