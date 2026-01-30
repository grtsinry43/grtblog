export type ImageLazyOptions = {
	blur?: string;
	src?: string;
};

export const imageLazy = (node: HTMLImageElement, options: ImageLazyOptions = {}) => {
	let lastSrc = options.src;

	const setLoaded = (loaded: boolean) => {
		node.dataset.loaded = loaded ? 'true' : 'false';
	};

	const applyBlur = (blur?: string) => {
		if (blur) {
			node.style.setProperty('--md-img-blur', blur);
		} else {
			node.style.removeProperty('--md-img-blur');
		}
	};

	const onLoad = () => setLoaded(true);
	const onError = () => setLoaded(true);

	setLoaded(node.complete && node.naturalWidth > 0);
	applyBlur(options.blur);

	node.addEventListener('load', onLoad);
	node.addEventListener('error', onError);

	return {
		update(next?: ImageLazyOptions) {
			if (next?.src && next.src !== lastSrc) {
				lastSrc = next.src;
				setLoaded(false);
			}
			applyBlur(next?.blur);
		},
		destroy() {
			node.removeEventListener('load', onLoad);
			node.removeEventListener('error', onError);
		}
	};
};
