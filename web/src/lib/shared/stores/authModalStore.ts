import { writable } from 'svelte/store';

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
		open: (source?: string) => set({ open: true, source }),
		close: () => set(initialState)
	};
}

export const authModalStore = createAuthModalStore();
