<script lang="ts">
    import {createMutation, createQuery} from '@tanstack/svelte-query';
    import {authLogin} from '$lib/shared/actions/auth-login';
    import {authModalStore} from '$lib/shared/stores/authModalStore';
    import {websiteInfoCtx} from '$lib/features/website-info/context.js';
    import Button from '$lib/ui/ui/button/Button.svelte';
    import Input from '$lib/ui/ui/input/Input.svelte';
    import {getProfile, login} from '$lib/features/auth/api';
    import type {LoginReq, LoginResp} from '$lib/features/auth/types';
    import {getToken, removeToken} from '$lib/shared/token';
    import {userStore} from '$lib/shared/stores/userStore';
    import type {UserInfo} from "$lib/shared/types/user";

    let error = $state('');
    let loading = $state(false);
    let token = $state(getToken());

    const websiteName = websiteInfoCtx.selectModelData(
        (data) => data?.website_name || 'grtBlog'
    );

    const handleClose = () => {
        authModalStore.close();
    };

    const loginMutation = createMutation<LoginResp, Error, LoginReq>(() => ({
        mutationFn: async (payload) => login(payload)
    }));

    createQuery(() => ({
        queryKey: ['auth-profile'],
        enabled: !!token && !$userStore.isLogin,
        retry: false,
        queryFn: () => getProfile(),
        onSuccess: (user: UserInfo) => {
            userStore.setUser(user);
        },
        onError: () => {
            removeToken();
            token = null;
            userStore.clear();
        }
    }));

    const executeLogin = async (payload: LoginReq) => {
        return await loginMutation.mutateAsync(payload);
    };

    $effect(() => {
        if (!$authModalStore.open) {
            error = '';
            loading = false;
        }
    });

</script>

{#if $authModalStore.open}
    <div class="fixed inset-0 z-50 flex items-center justify-center px-6 py-10">
        <button
                class="absolute inset-0 bg-ink-900/40 backdrop-blur-sm"
                aria-label="关闭登录弹窗"
                onclick={handleClose}
        ></button>
        <div
                class="relative w-full max-w-md rounded-default border border-ink-100 bg-white/90 p-8 shadow-[var(--shadow-glass)] backdrop-blur dark:border-ink-800 dark:bg-ink-900/70"
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

            <form
                    class="mt-8 space-y-5"
                    use:authLogin={{
					execute: executeLogin,
					onStart: () => {
						error = '';
						loading = true;
					},
					onSuccess: () => {
						loading = false;
						token = getToken();
						authModalStore.close();
					},
					onError: (err) => {
						loading = false;
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

                {#if error}
                    <p class="text-sm text-cinnabar-600 dark:text-cinnabar-400">{error}</p>
                {/if}

                <Button
                        class="w-full rounded-default bg-jade-600 text-white hover:bg-jade-700"
                        type="submit"
                        loading={loading}
                >
                    {loading ? '登录中…' : '登录'}
                </Button>
            </form>
        </div>
    </div>
{/if}

<style lang="postcss">
    @reference "$routes/layout.css";
</style>
