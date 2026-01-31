import { writable } from 'svelte/store';
import type { UserInfo } from '$lib/shared/types/user';

export type UserState = {
	isLogin: boolean;
	userInfo: UserInfo | null;
};

const initialState: UserState = {
	isLogin: false,
	userInfo: null
};

function createUserStore() {
	const { subscribe, set, update } = writable<UserState>(initialState);

	return {
		subscribe,
		setUser: (userInfo: UserInfo) => {
			set({ isLogin: true, userInfo });
		},
		setLogin: (isLogin: boolean) => {
			update((state) => ({ ...state, isLogin }));
		},
		updateUser: (patch: Partial<UserInfo>) => {
			update((state) => {
				if (!state.userInfo) return state;
				return { isLogin: true, userInfo: { ...state.userInfo, ...patch } };
			});
		},
		clear: () => {
			set(initialState);
		}
	};
}

export const userStore = createUserStore();
