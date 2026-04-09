import { writable } from 'svelte/store';
import { windowStore } from './windowStore.svelte';

export type AuthModalState = {
	open: boolean;
	source?: string;
};

const initialState: AuthModalState = {
	open: false
};

function createAuthModalStore() {
	const { subscribe, set } = writable<AuthModalState>(initialState);

	return {
		subscribe,
		open: (source?: string) => {
			set({ open: true, source });
			windowStore.open('登录', null, 'login');
		},
		close: () => {
			set(initialState);
			if (windowStore.kind === 'login') {
				windowStore.close();
			}
		},
		/** Reset internal state without touching windowStore (used for external close sync). */
		_reset: () => set(initialState)
	};
}

export const authModalStore = createAuthModalStore();
