import { browser } from '$app/environment';

type WindowMessageOptions = {
	handler: (event: MessageEvent) => void;
};

export const windowMessage = (_node: HTMLElement, options: WindowMessageOptions) => {
	if (!browser) {
		return {
			update() {},
			destroy() {}
		};
	}

	let currentHandler = options?.handler;
	const listener = (event: MessageEvent) => {
		currentHandler?.(event);
	};

	window.addEventListener('message', listener);

	return {
		update(next: WindowMessageOptions) {
			currentHandler = next?.handler;
		},
		destroy() {
			window.removeEventListener('message', listener);
		}
	};
};
