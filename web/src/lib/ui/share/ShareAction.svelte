<script lang="ts">
	import { onDestroy } from 'svelte';
	import { cubicOut } from 'svelte/easing';
	import { fade, fly } from 'svelte/transition';
	import { Check, Download, Image, LoaderCircle, Share2, X } from 'lucide-svelte';
	import { toast } from 'svelte-sonner';
	import { portal } from '$lib/shared/actions/portal';
	import { windowStore } from '$lib/shared/stores/windowStore.svelte';
	import {
		downloadSharePoster,
		generateSharePoster,
		readSharePosterContent,
		sharePageLink,
		type SharePosterContent
	} from '$lib/shared/share/share-poster';

	let {
		label = '',
		iconSize = 16,
		className = '',
		buttonTitle = '分享',
		shareTitle = '',
		shareDescription = '',
		shareImageUrl = ''
	} = $props<{
		label?: string;
		iconSize?: number;
		className?: string;
		buttonTitle?: string;
		shareTitle?: string;
		shareDescription?: string;
		shareImageUrl?: string;
	}>();

	let shareRoot: HTMLDivElement;
	let isMenuOpen = $state(false);
	let isPosterOpen = $state(false);
	let isGenerating = $state(false);
	let posterBlob: Blob | null = $state(null);
	let posterUrl = $state('');
	let content: SharePosterContent | null = $state(null);

	function toggleShareMenu() {
		const metadata = readSharePosterContent();
		content = {
			...metadata,
			description: shareDescription.trim() || metadata.description,
			imageUrl: shareImageUrl.trim() || metadata.imageUrl,
			title: shareTitle.trim() || metadata.title
		};
		isMenuOpen = !isMenuOpen;
	}

	function closeShareMenu() {
		isMenuOpen = false;
	}

	function closePosterPanel() {
		isPosterOpen = false;
	}

	async function handleLinkShare() {
		if (!content) return;
		try {
			const result = await sharePageLink(content);
			if (result === 'copied') toast.success('链接已复制到剪贴板');
			closeShareMenu();
		} catch (error) {
			if ((error as Error).name !== 'AbortError') toast.error('分享失败，请稍后重试');
		}
	}

	async function handlePosterGeneration() {
		if (!content || isGenerating) return;
		isGenerating = true;
		try {
			const blob = await generateSharePoster(content);
			if (posterUrl) URL.revokeObjectURL(posterUrl);
			posterBlob = blob;
			posterUrl = URL.createObjectURL(blob);
			closeShareMenu();
			if (window.matchMedia('(max-width: 767px)').matches) {
				windowStore.open(
					'分享卡片',
					{ posterBlob: blob, posterUrl, title: content.title },
					'share-poster'
				);
			} else {
				isPosterOpen = true;
			}
		} catch (error) {
			console.error('Failed to generate share poster:', error);
			toast.error('分享卡片生成失败，请稍后重试');
		} finally {
			isGenerating = false;
		}
	}

	function handlePosterDownload() {
		if (!posterBlob || !content) return;
		downloadSharePoster(posterBlob, content.title);
		toast.success('图片已保存');
	}

	function resetPoster() {
		if (posterUrl) URL.revokeObjectURL(posterUrl);
		posterBlob = null;
		posterUrl = '';
	}

	function returnToShareMenu() {
		closePosterPanel();
		resetPoster();
		isMenuOpen = true;
	}

	function handleKeydown(event: KeyboardEvent) {
		if (event.key !== 'Escape') return;
		if (isPosterOpen) closePosterPanel();
		else if (isMenuOpen) closeShareMenu();
	}

	function handleWindowPointerDown(event: PointerEvent) {
		if (!isMenuOpen || shareRoot.contains(event.target as Node)) return;
		closeShareMenu();
	}

	function cleanup() {
		if (windowStore.kind === 'share-poster' && windowStore.data?.posterUrl === posterUrl) {
			windowStore.close();
		}
		resetPoster();
	}

	onDestroy(cleanup);
</script>

<svelte:window onkeydown={handleKeydown} onpointerdown={handleWindowPointerDown} />

<div class="relative inline-flex" bind:this={shareRoot}>
	<button
		type="button"
		class={className}
		title={buttonTitle}
		aria-haspopup="menu"
		aria-expanded={isMenuOpen}
		onclick={toggleShareMenu}
	>
		<Share2 size={iconSize} />
		{#if label}<span>{label}</span>{/if}
	</button>

	{#if isMenuOpen}
		<div
			class="absolute top-full right-0 z-[120] mt-2 w-52 overflow-hidden rounded-default border border-ink-200/70 bg-[#fbfaf7]/98 p-1.5 shadow-deep backdrop-blur-md dark:border-ink-700 dark:bg-ink-900/98"
			role="menu"
			aria-label="选择分享方式"
			transition:fly={{ y: -5, duration: 160, easing: cubicOut }}
		>
			<button
				type="button"
				class="flex w-full items-center gap-3 rounded-md px-3 py-2.5 text-left transition-colors hover:bg-ink-100/80 dark:hover:bg-ink-800"
				role="menuitem"
				onclick={handleLinkShare}
			>
				<span
					class="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-ink-100 text-ink-600 dark:bg-ink-800 dark:text-ink-300"
				>
					<Share2 size={15} />
				</span>
				<span>
					<span class="block text-sm font-medium text-ink-800 dark:text-ink-100">系统分享</span>
					<span class="mt-0.5 block text-[11px] text-ink-400">分享链接或复制地址</span>
				</span>
			</button>

			<button
				type="button"
				class="flex w-full items-center gap-3 rounded-md px-3 py-2.5 text-left transition-colors hover:bg-jade-500/[0.08] dark:hover:bg-jade-900/25"
				role="menuitem"
				disabled={isGenerating}
				onclick={handlePosterGeneration}
			>
				<span
					class="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-jade-500/10 text-jade-700 dark:text-jade-400"
				>
					{#if isGenerating}
						<LoaderCircle size={15} class="animate-spin" />
					{:else}
						<Image size={15} />
					{/if}
				</span>
				<span>
					<span class="block text-sm font-medium text-ink-800 dark:text-ink-100">生成分享卡片</span>
					<span class="mt-0.5 block text-[11px] text-ink-400">根据本页内容生成</span>
				</span>
			</button>
		</div>
	{/if}
</div>

{#if isPosterOpen && posterUrl}
	<div use:portal class="fixed inset-0 z-[110] flex items-center justify-center px-4 py-6">
		<button
			type="button"
			class="absolute inset-0 bg-ink-950/30 backdrop-blur-sm dark:bg-black/60"
			aria-label="关闭分享面板"
			onclick={closePosterPanel}
			transition:fade={{ duration: 180 }}
		></button>

		<div
			class="relative z-10 max-h-full w-full max-w-[880px] overflow-auto rounded-default border border-ink-200/70 bg-[#fbfaf7] shadow-deep dark:border-ink-800 dark:bg-ink-900"
			role="dialog"
			aria-modal="true"
			aria-label="分享此页"
			transition:fly={{ y: 18, duration: 320, easing: cubicOut }}
		>
			<header
				class="flex items-start justify-between border-b border-ink-200/60 px-6 py-5 dark:border-ink-800/70"
			>
				<div>
					<div class="mb-1 font-serif text-lg font-medium text-ink-900 dark:text-ink-50">
						分享卡片
					</div>
					<p class="text-xs leading-5 text-ink-500">根据本页内容生成，可保存或直接分享。</p>
				</div>
				<button
					type="button"
					class="rounded-full p-2 text-ink-400 transition-colors hover:bg-ink-100 hover:text-ink-700 dark:hover:bg-ink-800 dark:hover:text-ink-200"
					aria-label="关闭"
					onclick={closePosterPanel}
				>
					<X size={18} />
				</button>
			</header>

			<div class="grid gap-6 p-5 md:grid-cols-[minmax(0,1fr)_260px] md:p-7">
				<div
					class="flex max-h-[65vh] items-center justify-center overflow-hidden rounded-lg bg-ink-100/70 p-3 dark:bg-ink-950/60"
				>
					<img
						src={posterUrl}
						alt="生成的分享卡片预览"
						class="max-h-[62vh] w-auto rounded-sm object-contain shadow-lg"
					/>
				</div>
				<div class="flex flex-col justify-between gap-8 py-1">
					<div class="space-y-4">
						<div class="flex items-center gap-2 text-sm font-medium text-ink-800 dark:text-ink-100">
							<span
								class="flex h-6 w-6 items-center justify-center rounded-full bg-jade-500/10 text-jade-600 dark:text-jade-400"
								><Check size={14} /></span
							>
							分享卡片已生成
						</div>
						<p class="text-xs leading-6 text-ink-500">保存图片后，即可发送给朋友。</p>
					</div>
					<div class="space-y-2">
						<button
							type="button"
							class="flex w-full items-center justify-center gap-2 rounded-default bg-ink-900 px-4 py-3 text-sm text-white transition-colors hover:bg-jade-700 dark:bg-jade-600 dark:hover:bg-jade-500"
							onclick={handlePosterDownload}
						>
							<Download size={16} /> 保存图片
						</button>
						<button
							type="button"
							class="w-full py-2 text-xs text-ink-400 transition-colors hover:text-ink-700 dark:hover:text-ink-200"
							onclick={returnToShareMenu}
						>
							返回分享方式
						</button>
					</div>
				</div>
			</div>
		</div>
	</div>
{/if}
