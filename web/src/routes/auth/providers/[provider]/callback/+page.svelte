<script lang="ts">
	import { browser } from '$app/environment';
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { bindOAuth } from '$lib/features/user-center/api';
	import { callbackOAuthProvider } from '$lib/features/auth/api';
	import { consumeOAuthFlowMeta, type OAuthPopupResult } from '$lib/features/auth/oauth-flow';
	import { setToken } from '$lib/shared/token';
	import { userStore } from '$lib/shared/stores/userStore';

	let status = $state('正在完成 OAuth 授权...');
	let errorText = $state('');
	let returnTo = $state('/');

	const postResult = (payload: OAuthPopupResult) => {
		if (window.opener && !window.opener.closed) {
			window.opener.postMessage(payload, window.location.origin);
			window.close();
			return true;
		}
		return false;
	};

	onMount(() => {
		if (!browser) return;
		void (async () => {
			const provider = page.params.provider ?? '';
			const code = page.url.searchParams.get('code') ?? '';
			const state = page.url.searchParams.get('state') ?? '';

			if (!provider || !code || !state) {
				status = '授权参数缺失';
				errorText = '缺少 provider/code/state';
				return;
			}

			const meta = consumeOAuthFlowMeta(state);
			const mode = meta?.mode ?? 'login';
			returnTo = meta?.returnTo || '/';
			const redirectUri = `${window.location.origin}/auth/providers/${provider}/callback/`;

			try {
				if (mode === 'bind') {
					await bindOAuth(provider, { code, state, redirectUri });
					status = '绑定成功，正在返回...';
					const handled = postResult({
						type: 'oauth:popup-result',
						success: true,
						mode,
						provider,
						returnTo
					});
					if (!handled) window.location.replace(returnTo);
					return;
				}

				const result = await callbackOAuthProvider(provider, { code, state, redirectUri });
				setToken(result.token);
				userStore.setUser(result.user);
				status = '登录成功，正在返回...';

				const handled = postResult({
					type: 'oauth:popup-result',
					success: true,
					mode,
					provider,
					token: result.token,
					user: result.user,
					returnTo
				});
				if (!handled) window.location.replace(returnTo);
			} catch (err) {
				const message = err instanceof Error ? err.message : '授权失败';
				status = '授权失败';
				errorText = message;
				const handled = postResult({
					type: 'oauth:popup-result',
					success: false,
					mode,
					provider,
					error: message,
					returnTo
				});
				if (!handled) {
					setTimeout(() => window.location.replace(returnTo), 1200);
				}
			}
		})();
	});
</script>

<section class="mx-auto max-w-lg px-4 py-20">
	<div
		class="rounded-default border border-ink-200 bg-white/90 p-6 text-center shadow-sm dark:border-ink-700 dark:bg-ink-900/80"
	>
		<h1 class="text-lg font-serif text-ink-900 dark:text-ink-100">{status}</h1>
		{#if errorText}
			<p class="mt-3 text-sm text-cinnabar-600 dark:text-cinnabar-400">{errorText}</p>
		{/if}
		<p class="mt-3 text-xs text-ink-500 dark:text-ink-400">
			如果页面没有自动关闭，请返回原页面继续操作。
		</p>
	</div>
</section>
