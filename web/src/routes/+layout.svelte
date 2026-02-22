<script lang="ts">
	import './layout.css';
	import favicon from '$lib/assets/favicon.svg';
	import Sidebar from '$lib/ui/layout/sidebar/Sidebar.svelte';
	import MobileNavBar from '$lib/ui/layout/sidebar/MobileNavBar.svelte';
	import { initTheme, startThemeSync, themeManager } from '$lib/shared/theme/theme.svelte.js';
	import { onMount } from 'svelte';
	import { consoleLogInfo } from '$lib/features/console-info/index';
	import Toaster from '$lib/ui/primitives/toaster/Toaster.svelte';
	import QueryRoot from '$lib/ui/common/QueryRoot.svelte';
	import Loading from '$lib/ui/common/Loading.svelte';
	import { navigating } from '$app/stores';
	import { browser } from '$app/environment';
	import { page } from '$app/state';
	import { onNavigate } from '$app/navigation';
	import SearchModal from '$lib/ui/search/SearchModal.svelte';
	import Footer from '$lib/ui/layout/Footer.svelte';
	import FloatingWindow from '$lib/ui/common/FloatingWindow.svelte';
	import { uiState } from '$lib/shared/stores/ui.svelte';
	import { windowStore } from '$lib/shared/stores/windowStore.svelte';
	import { presenceStore } from '$lib/features/presence/store.svelte';
	import { ownerStatusStore } from '$lib/features/owner-status/store.svelte';
	import { resolvePresenceView } from '$lib/features/presence/resolve-view';
	import PresencePagesWindow from '$lib/features/presence/components/PresencePagesWindow.svelte';
	import ThinkingCommentsWindow from '$lib/features/thinking/components/ThinkingCommentsWindow.svelte';

	function handleKeydown(event: KeyboardEvent) {
		if ((event.metaKey || event.ctrlKey) && (event.key === 'k' || event.key === 'K')) {
			event.preventDefault();
			uiState.toggleSearch();
		}
	}

	onNavigate((navigation) => {
		if (typeof document === 'undefined' || !document.startViewTransition) return;
		const startViewTransition = document.startViewTransition.bind(document);

		return new Promise((resolve) => {
			startViewTransition(async () => {
				resolve();
				await navigation.complete;
			});
		});
	});

	import '@fontsource/google-sans';
	import '@fontsource/noto-serif-sc';
	import '@fontsource-variable/victor-mono';
	import { websiteInfoCtx } from '$lib/features/website-info/context.js';
	import { resolveSeoMeta } from '$lib/shared/seo/metadata';
	import {
		createEmptyDetailPanelModel,
		detailPanelCtx,
		type DetailPanelModel,
		type DetailPanelRelatedMoment,
		type DetailPanelRelatedPost
	} from '$lib/shared/detail-panel/context';

	let { children, data } = $props();
	let showRouteLoading = $state(false);

	websiteInfoCtx.mountModelData(() => data.websiteInfo ?? null);

	const readDetailPanelFromPageData = (view: unknown): DetailPanelModel => {
		const empty = createEmptyDetailPanelModel();
		if (!view || typeof view !== 'object') return empty;
		const viewData = view as {
			post?: {
				title?: string | null;
				toc?: DetailPanelModel['toc'] | null;
				relatedMoments?: DetailPanelRelatedMoment[] | null;
			};
			moment?: {
				title?: string | null;
				toc?: DetailPanelModel['toc'] | null;
				relatedPosts?: DetailPanelRelatedPost[] | null;
			};
			page?: {
				title?: string | null;
				toc?: DetailPanelModel['toc'] | null;
			};
		};

		if (viewData.post) {
			return {
				...empty,
				kind: 'post',
				title: viewData.post.title ?? '',
				toc: viewData.post.toc ?? [],
				relatedMoments: (viewData.post.relatedMoments ?? []).slice(0, 2)
			};
		}
		if (viewData.moment) {
			return {
				...empty,
				kind: 'moment',
				title: viewData.moment.title ?? '',
				toc: viewData.moment.toc ?? [],
				relatedPosts: (viewData.moment.relatedPosts ?? []).slice(0, 2)
			};
		}
		if (viewData.page) {
			return {
				...empty,
				kind: 'page',
				title: viewData.page.title ?? '',
				toc: viewData.page.toc ?? []
			};
		}

		return empty;
	};

	detailPanelCtx.mountModelData(() => {
		const pathname = page.url.pathname;
		const pageData = page.data;
		const hash = browser ? window.location.hash.replace(/^#/, '') : '';
		return {
			...readDetailPanelFromPageData(pageData),
			contentRoot: null,
			activeAnchor:
				pathname === (browser ? window.location.pathname : pathname) ? hash || null : null
		};
	});

	const websiteInfoStore = websiteInfoCtx.selectModelData((model) => model ?? null);
	const siteFavicon = $derived.by(() => $websiteInfoStore?.favicon || favicon);
	const seoMeta = $derived.by(() =>
		resolveSeoMeta({
			pathname: page.url.pathname,
			search: page.url.search,
			routeData: page.data,
			websiteInfo: $websiteInfoStore,
			origin: page.url.origin,
			fallbackSiteIcon: siteFavicon
		})
	);

	// Initialize theme on mount
	const theme = themeManager;

	function openPresenceWindow() {
		windowStore.open('在线页面', null, 'presence-pages');
	}

	onMount(() => {
		initTheme(theme);
		consoleLogInfo();
		presenceStore.start();
		ownerStatusStore.start();
		return () => {
			presenceStore.stop();
			ownerStatusStore.stop();
		};
	});

	startThemeSync(theme);

	$effect(() => {
		if (!browser) {
			return;
		}

		if ($navigating) {
			showRouteLoading = false;
			const timer = setTimeout(() => {
				if ($navigating) {
					showRouteLoading = true;
				}
			}, 2000);

			return () => {
				clearTimeout(timer);
			};
		}

		showRouteLoading = false;
	});

	$effect(() => {
		if (!browser) return;

		const report = resolvePresenceView(page.url.pathname, page.data);
		if (!report) return;

		presenceStore.reportView(report);
	});
</script>

<svelte:head>
	<link rel="icon" href={siteFavicon} />
	<title>{seoMeta.title}</title>
	<link rel="canonical" href={seoMeta.canonicalUrl} />
	<meta name="viewport" content="width=device-width, initial-scale=1" />
	<meta name="description" content={seoMeta.description} />
	<meta name="keywords" content={seoMeta.keywords} />
	<meta name="robots" content={seoMeta.robots} />
	<meta name="author" content="grtinry43" />
	<meta property="og:title" content={seoMeta.ogTitle} />
	<meta property="og:description" content={seoMeta.ogDescription} />
	<meta property="og:type" content={seoMeta.ogType} />
	<meta property="og:url" content={seoMeta.ogUrl} />
	<meta property="og:site_name" content={seoMeta.ogSiteName} />
	<meta property="og:image" content={seoMeta.ogImage} />
	<meta name="twitter:card" content={seoMeta.twitterCard} />
	<meta name="twitter:title" content={seoMeta.ogTitle} />
	<meta name="twitter:description" content={seoMeta.ogDescription} />
	<meta name="twitter:image" content={seoMeta.ogImage} />
	<script>
		// Inline script to prevent theme flash (fallback before Svelte hydrates)
		(function () {
			try {
				const theme = localStorage.getItem('theme') || 'system';
				const isDark =
					theme === 'dark' ||
					(theme === 'system' && window.matchMedia('(prefers-color-scheme: dark)').matches);
				document.documentElement.classList.toggle('dark', isDark);
			} catch (e) {}
		})();
	</script>
</svelte:head>

<div class="hidden md:block">
	<Sidebar menuTree={data.navMenus ?? []} />
</div>
<MobileNavBar menuTree={data.navMenus ?? []} />
<!-- noise background -->
<div class="bg-noise" aria-hidden="true"></div>

<div class="md:pl-24 transition-[padding] duration-300">
	<main
		class="page-wrapper mx-auto {page.url.pathname.startsWith('/timeline')
			? 'max-w-none px-0 py-0'
			: 'max-w-300 px-4 sm:px-6 lg:px-8 py-10 md:py-16'}"
	>
		<div class="content-container min-h-[60vh]">
			{@render children()}
		</div>
	</main>
	<Footer
		onlineCount={presenceStore.online}
		presenceConnected={presenceStore.isConnected}
		onOpenPresence={openPresenceWindow}
	/>
</div>

{#if showRouteLoading}
	<div
		class="fixed px-12 py-6 left-1/2 top-1/2 z-99999 -translate-x-1/2 -translate-y-1/2 pointer-events-none rounded-default border border-ink-200/70 bg-ink-50/80 shadow-subtle backdrop-blur-lg dark:border-ink-700/70 dark:bg-ink-900/80"
		aria-live="polite"
		aria-busy="true"
	>
		<Loading size="w-8 h-8" duration={900} class="gap-0" text="正在玩命加载中...莫慌" />
	</div>
{/if}

<SearchModal />
<FloatingWindow>
	{#if windowStore.kind === 'tag-contents'}
		<QueryRoot
			loader={() => import('$lib/features/tag/components/TagContentsWindow.svelte')}
			loaderProps={{ tagId: windowStore.data?.id, tagName: windowStore.data?.name }}
		/>
	{:else if windowStore.title === '申请友链'}
		<QueryRoot
			loader={() => import('$lib/features/friend-link/components/ApplyFriendForm.svelte')}
		/>
	{:else if windowStore.kind === 'presence-pages'}
		<PresencePagesWindow />
	{:else if windowStore.kind === 'thinking-comments'}
		<ThinkingCommentsWindow
			areaId={windowStore.data?.areaId}
			commentsCount={windowStore.data?.commentsCount ?? 0}
		/>
	{:else if windowStore.kind === 'user-center'}
		<QueryRoot
			loader={() => import('$lib/features/user-center/components/UserCenterWindow.svelte')}
		/>
	{:else}
		<div class="flex flex-col gap-3"></div>
	{/if}
</FloatingWindow>

<svelte:window onkeydown={handleKeydown} />

<Toaster />
{#snippet globalNotificationFallback()}
	<div></div>
{/snippet}
<QueryRoot
	loader={() =>
		import('$lib/features/global-notification/components/GlobalNotificationClient.svelte')}
	fallback={globalNotificationFallback}
/>
{#snippet authFallback()}
	<div></div>
{/snippet}
<QueryRoot
	loader={() => import('$lib/features/auth/components/AuthClient.svelte')}
	fallback={authFallback}
/>

<style lang="postcss">
	@reference "./layout.css";

	:global(html) {
		scroll-behavior: smooth;
		scroll-padding-top: 80px;
	}
</style>
