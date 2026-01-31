<script lang="ts">
    import {browser} from '$app/environment';
    import {createMutation, createQuery} from '@tanstack/svelte-query';
    import {authLogin} from '$lib/shared/actions/auth-login';
    import {preloadTurnstile, turnstileWidget} from '$lib/shared/actions/turnstile';
    import {authModalStore} from '$lib/shared/stores/authModalStore';
    import {websiteInfoCtx} from '$lib/features/website-info/context.js';
    import Button from '$lib/ui/ui/button/Button.svelte';
    import Input from '$lib/ui/ui/input/Input.svelte';
    import {
        authorizeOAuthProvider,
        getProfile,
        getTurnstileState,
        listOAuthProviders,
        login
    } from '$lib/features/auth/api';
    import type {LoginReq, LoginResp, OAuthProvider, TurnstileStateResp} from '$lib/features/auth/types';
    import {getToken} from '$lib/shared/token';
    import {userStore} from '$lib/shared/stores/userStore';
    import type {UserInfo} from "$lib/shared/types/user";

    let error = $state('');
    let loading = $state(false);
    let token = $state(getToken());
    let oauthProviders = $state<OAuthProvider[]>([]);
    let oauthLoadingKey = $state<string | null>(null);
    let oauthError = $state('');
    let showPasswordLogin = $state(false);
    let turnstileEnabled = $state(false);
    let turnstileSiteKey = $state('');
    let turnstileToken = $state('');
    let turnstileError = $state('');
    let turnstileRequested = $state(false);
    let canSubmit = $derived(
        !turnstileEnabled || (turnstileSiteKey.length > 0 && turnstileToken.length > 0)
    );
    let hasOAuthProviders = $derived(oauthProviders.length > 0);

    const websiteName = websiteInfoCtx.selectModelData(
        (data) => data?.website_name || 'grtBlog'
    );

    const handleClose = () => {
        authModalStore.close();
    };

    const loginMutation = createMutation<LoginResp, Error, LoginReq>(() => ({
        mutationFn: async (payload) => login(payload)
    }));

    const loadOAuthProviders = async () => {
        if (!$authModalStore.open) return;
        try {
            const items = await listOAuthProviders();
            oauthProviders = Array.isArray(items) ? items : [];
            showPasswordLogin = oauthProviders.length === 0;
            oauthError = '';
        } catch (err) {
            oauthProviders = [];
            showPasswordLogin = true;
            oauthError = err instanceof Error ? err.message : '获取 OAuth 登录方式失败';
        }
    };

    const loadTurnstileState = async () => {
        if (turnstileRequested) return;
        turnstileRequested = true;
        try {
            const state = await getTurnstileState();
            turnstileEnabled = !!state?.enabled;
            turnstileSiteKey = state?.siteKey ?? '';
            if (!turnstileEnabled) {
                turnstileToken = '';
                turnstileError = '';
            } else if (turnstileSiteKey) {
                preloadTurnstile().catch(() => {
                    turnstileError = '人机验证加载失败，请检查网络或拦截设置';
                });
            }
        } catch (err) {
            turnstileEnabled = false;
            turnstileSiteKey = '';
            turnstileToken = '';
            turnstileError = '';
        }
    };

    createQuery(() => ({
        queryKey: ['auth-profile'],
        enabled: !!token && !$userStore.isLogin,
        retry: false,
        queryFn: () => getProfile(),
        onSuccess: (user: UserInfo) => {
            userStore.setUser(user);
        },
        onError: () => {
            // removeToken();
            // token = null;
            // userStore.clear();
        }
    }));

    const executeLogin = async (payload: LoginReq) => {
        return await loginMutation.mutateAsync(payload);
    };

    const startOAuthLogin = async (provider: OAuthProvider) => {
        if (!browser) return;
        oauthLoadingKey = provider.key;
        oauthError = '';
        try {
            const redirectUri = window.location.href;
            const res = await authorizeOAuthProvider(provider.key, redirectUri);
            window.location.href = res.authUrl;
        } catch (err) {
            oauthLoadingKey = null;
            oauthError = err instanceof Error ? err.message : '获取 OAuth 授权地址失败';
        }
    };

    $effect(() => {
        if (!$authModalStore.open) {
            error = '';
            loading = false;
            oauthLoadingKey = null;
            oauthError = '';
            showPasswordLogin = oauthProviders.length === 0;
            turnstileToken = '';
            turnstileError = '';
            turnstileRequested = false;
        } else {
            // modal opened
        }
    });

    $effect(() => {
        if ($authModalStore.open) {
            void loadTurnstileState();
            void loadOAuthProviders();
        }
    });

    $effect(() => {
        if (!showPasswordLogin) {
            error = '';
            loading = false;
            turnstileToken = '';
            turnstileError = '';
        }
    });
</script>

{#if $authModalStore.open}
    <div class="fixed inset-0 z-50 flex items-center justify-center px-6 py-10">
        <button
                class="absolute inset-0"
                aria-label="关闭登录弹窗"
                onclick={handleClose}
        ></button>
        <div
                class="relative w-full shadow-lg max-w-md rounded-default border border-ink-100 bg-white/90 p-8 backdrop-blur-lg dark:border-ink-800 dark:bg-ink-900/70"
                role="dialog"
                aria-modal="true"
                aria-label="登录"
        >
            <div class="flex items-start justify-between">
                <div>
                    <p class="text-xs font-mono text-ink-500"> 欢迎回来 </p>
                    <h2 class="mt-1 text-2xl font-serif text-ink-900 dark:text-ink-100">
                        登录到 {$websiteName}
                    </h2>
                </div>
                <Button
                        variant="icon"
                        class="h-8 w-8 rounded-default text-ink-400 hover:bg-ink-100 hover:text-ink-900 dark:hover:bg-ink-800 dark:hover:text-ink-100"
                        aria-label="关闭"
                        on:click={handleClose}
                >
                    ✕
                </Button>
            </div>

            <div class="mt-8 space-y-5">
                {#if hasOAuthProviders}
                    <div class="space-y-3">
                        <p class="text-xs font-mono text-ink-500"> 使用 OAuth 登录 </p>
                        <div class="grid gap-2">
                            {#each oauthProviders as provider}
                                <Button
                                        variant="secondary"
                                        fullWidth
                                        loading={oauthLoadingKey === provider.key}
                                        disabled={oauthLoadingKey !== null && oauthLoadingKey !== provider.key}
                                        on:click={() => startOAuthLogin(provider)}
                                >
                                    {provider.displayName}
                                </Button>
                            {/each}
                        </div>
                        {#if oauthError}
                            <p class="text-sm text-cinnabar-600 dark:text-cinnabar-400">{oauthError}</p>
                        {/if}
                    </div>
                {/if}

                {#if hasOAuthProviders}
                    <div class="flex items-center gap-3 text-xs text-ink-500">
                        <span class="h-px flex-1 bg-ink-100 dark:bg-ink-800"></span>
                        <Button
                                variant="ghost"
                                size="sm"
                                on:click={() => (showPasswordLogin = !showPasswordLogin)}
                        >
                            {showPasswordLogin ? '使用 OAuth 登录' : '使用账号密码登录'}
                        </Button>
                        <span class="h-px flex-1 bg-ink-100 dark:bg-ink-800"></span>
                    </div>
                {/if}

                {#if showPasswordLogin || !hasOAuthProviders}
                    <form
                            class="space-y-5"
                            use:authLogin={{
						execute: executeLogin,
						getPayload: (formEl) => {
							const data = new FormData(formEl);
							return {
								credential: String(data.get('credential') ?? ''),
								password: String(data.get('password') ?? ''),
								turnstileToken: turnstileToken || undefined
							};
						},
						onStart: () => {
							error = '';
							loading = true;
							turnstileError = '';
						},
					onSuccess: () => {
						loading = false;
						token = getToken();
						if (loginMutation.data?.user) {
							userStore.setUser(loginMutation.data.user);
						}
						authModalStore.close();
					},
					onError: (err) => {
						loading = false;
						turnstileToken = '';
						if (turnstileEnabled) {
							turnstileError = '人机校验未通过，请重试';
						}
						error = err instanceof Error ? err.message : '登录失败，请稍后重试';
					},
						onFinally: () => {
							loading = false;
						}
					}}
                    >
                        <div class="space-y-2">
                            <label class="text-xs font-mono text-ink-500"> 用户名 / 邮箱 </label>
                            <Input
                                    name="credential"
                                    autocomplete="username"
                                    required
                                    variant="underline"
                                    inputClass="rounded-default border-ink-200 bg-white/70 text-ink-900 placeholder:text-ink-400 focus:border-jade-500 focus:ring-2 dark:border-ink-700 dark:bg-ink-900/40 dark:text-ink-100 dark:placeholder:text-ink-500"
                            />
                        </div>

                        <div class="space-y-2">
                            <label class="text-xs font-mono text-ink-500"> 密码 </label>
                            <Input
                                    type="password"
                                    name="password"
                                    autocomplete="current-password"
                                    required
                                    variant="underline"
                                    inputClass="rounded-default border-ink-200 bg-white/70 text-ink-900 placeholder:text-ink-400 focus:border-jade-500 focus:ring-2 dark:border-ink-700 dark:bg-ink-900/40 dark:text-ink-100 dark:placeholder:text-ink-5000"
                            />
                        </div>

                        {#if turnstileEnabled}
                            <div class="space-y-2">
                                <label class="text-xs font-mono text-ink-500"> 人机验证 </label>
                                {#if turnstileSiteKey}
                                    <div
                                            class="rounded-default border border-ink-100 bg-white/70 p-3 dark:border-ink-800 dark:bg-ink-900/40"
                                            use:turnstileWidget={{
									siteKey: turnstileSiteKey,
									onToken: (token) => {
										turnstileToken = token;
										turnstileError = '';
									},
									onExpired: () => {
										turnstileToken = '';
									},
									onError: () => {
										turnstileToken = '';
										turnstileError = '人机验证失败，请重试';
									}
								}}
                                    ></div>
                                {:else}
                                    <p class="text-xs text-ink-500">Turnstile 未配置，请联系管理员。</p>
                                {/if}
                                {#if turnstileError}
                                    <p class="text-xs text-cinnabar-600 dark:text-cinnabar-400">{turnstileError}</p>
                                {/if}
                            </div>
                        {/if}

                        <input type="hidden" name="turnstileToken" value={turnstileToken} />

                        {#if error}
                            <p class="text-sm text-cinnabar-600 dark:text-cinnabar-400">{error}</p>
                        {/if}

                        <Button
                                class="w-full rounded-default bg-jade-600 text-white hover:bg-jade-700"
                                type="submit"
                                loading={loading}
                                disabled={!canSubmit}
                        >
                            {loading ? '登录中…' : '登录'}
                        </Button>
                    </form>
                {/if}
            </div>
        </div>
    </div>
{/if}

<style lang="postcss">
    @reference "$routes/layout.css";
</style>
