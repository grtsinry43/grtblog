import { browser } from '$app/environment';
import { authModalStore } from '$lib/shared/stores/authModalStore';

export type AuthLoginModalOptions = {
	source?: string;
	onOpen?: () => void | Promise<void>;
};

export const authLoginModal = (node: HTMLElement, options: AuthLoginModalOptions = {}) => {
	if (!browser) return {};

	let currentOptions = options;

	const onClick = async (event: MouseEvent) => {
		event.preventDefault();
		authModalStore.open(currentOptions.source);
		await currentOptions.onOpen?.();
	};

	node.addEventListener('click', onClick);

	return {
		update(next?: AuthLoginModalOptions) {
			currentOptions = next ?? {};
		},
		destroy() {
			node.removeEventListener('click', onClick);
		}
	};
};
