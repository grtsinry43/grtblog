<script lang="ts">
    import {browser} from '$app/environment';
    import {goto} from '$app/navigation';
    import {resolvePath} from '$lib/shared/utils/resolve-path';
    import {createMutation, createQuery} from '@tanstack/svelte-query';
    import {listOAuthProviders, authorizeOAuthProvider} from '$lib/features/auth/api';
    import {openOAuthPopup, saveOAuthFlowMeta, waitForOAuthPopupResult} from '$lib/features/auth/oauth-flow';
    import {userStore} from '$lib/shared/stores/userStore';
    import {getToken, removeToken} from '$lib/shared/token';
    import {
        getUserProfile,
        listOAuthBindings,
        unbindOAuth,
        updateUserProfile
    } from '$lib/features/user-center/api';
    import type {OAuthProvider} from '$lib/features/auth/types';
    import {toast} from 'svelte-sonner';
    import {windowStore} from '$lib/shared/stores/windowStore.svelte';
    import {ExternalLink, Unlink, Link as LinkIcon, LogOut, Check} from 'lucide-svelte';

    let profileForm = $state({
        nickname: '',
    });
    let bindingLoadingProvider = $state<string | null>(null);

    const meStore = $derived($userStore.userInfo);
    const isLogin = $derived($userStore.isLogin && !!$userStore.userInfo);

    const profileQuery = createQuery(() => ({
        queryKey: ['user-center', 'profile'],
        enabled: browser && !!getToken(),
        queryFn: () => getUserProfile(),
        retry: false,
        staleTime: 30_000
    }));

    const bindingQuery = createQuery(() => ({
        queryKey: ['user-center', 'oauth-bindings'],
        enabled: isLogin,
        queryFn: () => listOAuthBindings(),
        retry: false
    }));

    const providerQuery = createQuery(() => ({
        queryKey: ['user-center', 'oauth-providers'],
        enabled: isLogin,
        queryFn: () => listOAuthProviders(),
        retry: false
    }));

    const updateProfileMutation = createMutation(() => ({
        mutationFn: (payload: { nickname: string; email: string; avatar: string }) =>
            updateUserProfile(payload)
    }));

    const unbindMutation = createMutation(() => ({
        mutationFn: (provider: string) => unbindOAuth(provider)
    }));

    const boundProviderSet = $derived.by(() => {
        const keys = (bindingQuery.data ?? []).map((item) => item.providerKey);
        return new Set(keys);
    });

    const startBindOAuth = async (provider: OAuthProvider) => {
        if (!browser) return;
        bindingLoadingProvider = provider.key;
        try {
            const redirectUri = `${window.location.origin}/auth/providers/${provider.key}/callback/`;
            const res = await authorizeOAuthProvider(provider.key, redirectUri);
            saveOAuthFlowMeta(res.state, {
                mode: 'bind',
                provider: provider.key,
                returnTo: window.location.href,
                createdAt: Date.now()
            });
            const popup = openOAuthPopup(res.authUrl, provider.key);
            if (!popup) {
                window.location.href = res.authUrl;
                return;
            }
            const result = await waitForOAuthPopupResult({
                provider: provider.key,
                mode: 'bind',
                popup
            });
            if (!result.success) {
                toast.error(result.error || '账号绑定失败');
                return;
            }
            await bindingQuery.refetch();
            toast.success('账号绑定成功');
        } catch (err) {
            toast.error(err instanceof Error ? err.message : '获取 OAuth 授权地址失败');
        } finally {
            bindingLoadingProvider = null;
        }
    };

    const handleUnbind = async (provider: string) => {
        if (browser && !window.confirm('确认解绑该账号吗？')) return;
        try {
            await unbindMutation.mutateAsync(provider);
            await bindingQuery.refetch();
            toast.success('解绑成功');
        } catch (err) {
            toast.error(err instanceof Error ? err.message : '解绑失败');
        }
    };

    const saveProfile = async () => {
        if (!profileForm.nickname.trim()) {
            toast.error('昵称不能为空');
            return;
        }
        try {
            const user = await updateProfileMutation.mutateAsync({
                nickname: profileForm.nickname.trim(),
                email: meStore?.email ?? '',
                avatar: meStore?.avatar ?? ''
            });
            userStore.setUser(user);
            toast.success('昵称已更新');
        } catch (err) {
            toast.error(err instanceof Error ? err.message : '更新失败');
        }
    };

    const signOut = async () => {
        if (browser && !window.confirm('确认退出登录吗？')) return;
        removeToken();
        userStore.clear();
        windowStore.close();
        await goto(resolvePath('/'), {replaceState: true});
        toast.success('已退出登录');
    };

    $effect(() => {
        const profile = profileQuery.data;
        if (profile) {
            userStore.setUser(profile);
        }
        if (profileQuery.isError) {
            removeToken();
            userStore.clear();
            windowStore.close();
        }
    });

    $effect(() => {
        const me = meStore;
        if (!me) return;
        if (profileForm.nickname === '' && me.nickname) {
            profileForm.nickname = me.nickname;
        }
    });

</script>

<div class="space-y-6">
    {#if !isLogin}
        <div class="text-center py-8 text-ink-500">
            <p> 未登录状态无法访问用户中心 </p>
        </div>
    {:else}
        <!-- Profile Section -->
        <section class="space-y-3">
            <h3 class="text-xs font-bold text-ink-400 uppercase tracking-wider"> 个人资料 </h3>
            <label class="block space-y-1.5">
                <span class="text-xs text-ink-500 font-medium"> 邮箱 </span>
                <p class="text-sm text-ink-700 dark:text-ink-200">{meStore?.email || '未设置邮箱'}</p>
            </label>
            <div class="flex items-end gap-2">
                <label class="block flex-1 space-y-1.5">
                    <span class="text-xs text-ink-500 font-medium"> 昵称 </span>
                    <input
                            class="w-full rounded-default border border-ink-200 bg-white px-3 py-1.5 text-sm dark:border-ink-700 dark:bg-ink-800 transition-colors focus:border-jade-500 focus:ring-1 focus:ring-jade-500 outline-none"
                            bind:value={profileForm.nickname}
                            placeholder="请输入昵称"
                    />
                </label>
                <button
                        class="h-[34px] px-3 rounded-default bg-jade-600 text-white text-xs font-medium hover:bg-jade-700 transition-colors disabled:opacity-50 flex items-center gap-1.5"
                        onclick={saveProfile}
                        disabled={updateProfileMutation.isPending || profileForm.nickname === meStore?.nickname}
                >
                    {#if updateProfileMutation.isPending}
                        <span class="loading loading-spinner loading-xs"></span>
                    {:else}
                        <Check size={14}/>
                    {/if}
                    保存
                </button>
            </div>
            {#if meStore?.isAdmin}
                <div class="pt-1">
                    <a
                            href={resolvePath('/admin')}
                            target="_blank"
                            class="inline-flex items-center gap-1.5 text-xs text-jade-600 hover:text-jade-700 dark:text-jade-400 dark:hover:text-jade-300 transition-colors"
                    >
                        <ExternalLink size={12}/>
                        前往后台管理系统修改更多信息
                    </a>
                </div>
            {/if}
        </section>

        <!-- Binding Section -->
        <section class="space-y-3">
            <div class="flex items-center justify-between">
                <h3 class="text-xs font-bold text-ink-400 uppercase tracking-wider"> 账号绑定 </h3>
                {#if bindingQuery.isFetching}
                    <span class="text-[10px] text-ink-400 animate-pulse"> 刷新中...</span>
                {/if}
            </div>

            <div class="space-y-2">
                {#if (providerQuery.data ?? []).length === 0}
                    <div class="text-xs text-ink-400 italic"> 暂无可用登录方式</div>
                {:else}
                    {#each providerQuery.data ?? [] as provider (provider.key)}
                        {@const isBound = boundProviderSet.has(provider.key)}
                        <div
                                class="flex items-center justify-between rounded-default border p-2.5 transition-colors
							{isBound
								? 'border-jade-200 bg-jade-50/30 dark:border-jade-800/50 dark:bg-jade-900/10'
								: 'border-ink-200 bg-white dark:border-ink-700 dark:bg-ink-800/40'}"
                        >
                            <div class="flex items-center gap-2">
                                <!-- Icon placeholder if any, or just name -->
                                <div class="text-sm font-medium text-ink-700 dark:text-ink-200">
                                    {provider.displayName}
                                </div>
                            </div>

                            {#if isBound}
                                <div class="flex items-center gap-2">
                                    <span class="text-[10px] bg-jade-100 text-jade-700 dark:bg-jade-900/50 dark:text-jade-400 px-1.5 py-0.5 rounded-default font-medium"> 已绑定 </span>
                                    <button
                                            class="flex items-center gap-1 text-xs text-ink-400 hover:text-cinnabar-600 transition-colors px-2 py-1 rounded-default hover:bg-cinnabar-50 dark:hover:bg-cinnabar-900/20"
                                            disabled={unbindMutation.isPending}
                                            onclick={()=> handleUnbind(provider.key)}
                                            title="解除绑定"
                                    >
                                        <Unlink size={14}/>
                                        <span> 解绑 </span>
                                    </button>
                                </div>
                            {:else}
                                <button
                                        class="flex items-center gap-1 text-xs text-jade-600 hover:text-jade-700 transition-colors px-2 py-1 rounded-default hover:bg-jade-50 dark:hover:bg-jade-900/20 font-medium"
                                        onclick={()=> startBindOAuth(provider)}
                                        disabled={bindingLoadingProvider === provider.key}
                                >
                                    {#if bindingLoadingProvider === provider.key}
                                        <span class="loading loading-spinner loading-xs"></span>
                                    {:else}
                                        <LinkIcon size={14}/>
                                    {/if}
                                    <span> 绑定 </span>
                                </button>
                            {/if}
                        </div>
                    {/each}
                {/if}
            </div>
        </section>

        <!-- Footer Actions -->
        <div class="pt-4 border-t border-ink-100 dark:border-ink-800 flex justify-end">
            <button
                    class="flex items-center gap-1.5 text-xs text-ink-500 hover:text-cinnabar-600 transition-colors px-2 py-1.5 rounded-default hover:bg-ink-100 dark:hover:bg-ink-800"
                    onclick={signOut}
            >
                <LogOut size={14}/>
                退出登录
            </button>
        </div>
    {/if}
</div>
