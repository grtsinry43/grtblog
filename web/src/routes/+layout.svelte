<script lang="ts">
	import './layout.css';
	import favicon from '$lib/assets/favicon.svg';
	import Sidebar from '$lib/ui/layout/sidebar/Sidebar.svelte';
	import { initTheme, startThemeSync, themeManager } from '$lib/shared/theme/theme.svelte.js';
	import { onMount } from 'svelte';
	import { consoleLogInfo } from '$lib/features/console-info/index';
	import Toaster from '$lib/ui/ui/toaster/Toaster.svelte';
	import QueryRoot from '$lib/ui/common/QueryRoot.svelte';

	import "@fontsource/google-sans";
	import "@fontsource/noto-serif-sc";
	import "@fontsource-variable/victor-mono";
	import { websiteInfoCtx } from '$lib/features/website-info/context.js';

	let { children, data } = $props();

	const websiteInfoStore = websiteInfoCtx.mountModelData(data.websiteInfo ?? null);

	$effect(() => {
		websiteInfoCtx.syncModelData(websiteInfoStore, data.websiteInfo ?? null);
	});

	const websiteName = websiteInfoCtx.selectModelData((data) => data?.website_name || 'grtBlog');
	const keywords = websiteInfoCtx.selectModelData((data) => data?.keywords || 'blog, programming, technology, software development, web development, coding');
	const description = websiteInfoCtx.selectModelData((data) => data?.description || 'grtBlog - A personal blog about programming, technology, and software development.');
	const siteFavicon = websiteInfoCtx.selectModelData((data) => data?.favicon || favicon);
	const ogTitle = websiteInfoCtx.selectModelData((data) => data?.og_title || 'grtBlog');
	const ogType = websiteInfoCtx.selectModelData((data) => data?.og_type || 'website');
	const ogDescription = websiteInfoCtx.selectModelData((data) => data?.og_description || 'grtBlog - A personal blog about programming, technology, and software development.');
	const ogImage = websiteInfoCtx.selectModelData((data) => data?.og_image || '');
	const ogUrl = websiteInfoCtx.selectModelData((data) => data?.og_url || '');


	// Initialize theme on mount
	const theme = themeManager;

	onMount(() => {
		initTheme(theme);
		consoleLogInfo();
	});

	startThemeSync(theme);
</script>

<svelte:head>
	<link rel="icon" href={$siteFavicon} />
	<title>{$websiteName}</title>
	<meta name="viewport" content="width=device-width, initial-scale=1" />
	<meta name="description" content={$description} />
	<meta name="keywords" content={$keywords} />
	<meta name="author" content="grtinry43" />
	<meta property="og:title" content={$ogTitle} />
	<meta property="og:description" content={$ogDescription} />
	<meta property="og:type" content={$ogType} />
	<meta property="og:url" content={$ogUrl} />
	<meta property="og:image" content={$ogImage} />
	<meta name="twitter:card" content={$ogImage ? 'summary_large_image' : 'summary'} />
	<meta name="twitter:title" content={$ogTitle} />
	<meta name="twitter:description" content={$ogDescription} />
	<meta name="twitter:image" content={$ogImage} />
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

<Sidebar menuTree={data.navMenus ?? []} />
<!-- noise background -->
<div class="bg-noise" aria-hidden="true"></div>

<main class="page-wrapper max-w-[1200px] mx-auto px-4 sm:px-6 lg:px-8 py-10 md:py-16">
	<div class="content-container">
		{@render children()}
	</div>
</main>

<Toaster />
{#snippet authFallback()}
\t<div></div>
{/snippet}
<QueryRoot loader={() => import('$lib/features/auth/components/AuthClient.svelte')} fallback={authFallback} />

<style lang="postcss">
	@reference "./layout.css";

	:global(html) {
		scroll-behavior: smooth;
	}
</style>
