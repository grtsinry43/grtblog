import { browser } from '$app/environment';
import { login } from '$lib/features/auth/api';
import type { LoginReq, LoginResp } from '$lib/features/auth/types';
import { userStore } from '$lib/shared/stores/userStore';
import { setToken } from '$lib/shared/token';

export type LoginActionOptions = {
	getPayload?: (form: HTMLFormElement) => LoginReq;
	execute?: (payload: LoginReq) => Promise<LoginResp>;
	onStart?: () => void | Promise<void>;
	onSuccess?: (resp: LoginResp) => void | Promise<void>;
	onError?: (error: unknown) => void | Promise<void>;
	onFinally?: () => void | Promise<void>;
};

const defaultGetPayload = (form: HTMLFormElement): LoginReq => {
	const data = new FormData(form);
	return {
		credential: String(data.get('credential') ?? ''),
		password: String(data.get('password') ?? ''),
		turnstileToken: String(data.get('turnstileToken') ?? '') || undefined
	};
};

export const authLogin = (node: HTMLFormElement, options: LoginActionOptions = {}) => {
	if (!browser) return {};

	let currentOptions = options;

	const onSubmit = async (event: SubmitEvent) => {
		event.preventDefault();

		try {
			await currentOptions.onStart?.();
			const payload = (currentOptions.getPayload ?? defaultGetPayload)(node);
			const executor = currentOptions.execute ?? ((body: LoginReq) => login(body));
			const result = await executor(payload);
			setToken(result.token);
			userStore.setUser(result.user);
			await currentOptions.onSuccess?.(result);
		} catch (error) {
			await currentOptions.onError?.(error);
		} finally {
			await currentOptions.onFinally?.();
		}
	};

	node.addEventListener('submit', onSubmit);

	return {
		update(next?: LoginActionOptions) {
			currentOptions = next ?? {};
		},
		destroy() {
			node.removeEventListener('submit', onSubmit);
		}
	};
};
