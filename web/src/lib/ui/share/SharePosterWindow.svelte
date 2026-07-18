<script lang="ts">
	import { Download, Share2 } from 'lucide-svelte';
	import { toast } from 'svelte-sonner';
	import { downloadSharePoster, sharePosterImage } from '$lib/shared/share/share-poster';
	import { windowStore } from '$lib/shared/stores/windowStore.svelte';

	interface SharePosterWindowData extends Record<string, unknown> {
		posterBlob?: Blob;
		posterUrl?: string;
		title?: string;
	}

	const data = $derived((windowStore.data ?? {}) as SharePosterWindowData);

	async function handleShare() {
		if (!data.posterBlob) return;

		try {
			const shared = await sharePosterImage(data.posterBlob, data.title || '分享卡片');
			if (shared) {
				windowStore.close();
				return;
			}

			handleSave();
		} catch (error) {
			if ((error as Error).name !== 'AbortError') toast.error('分享失败，请稍后重试');
		}
	}

	function handleSave() {
		if (!data.posterBlob) return;
		downloadSharePoster(data.posterBlob, data.title || '分享卡片');
		toast.success('图片已保存');
	}
</script>

<div class="space-y-4">
	{#if data.posterUrl}
		<div
			class="flex max-h-[52vh] items-center justify-center overflow-hidden rounded-lg bg-ink-100/70 p-2.5 dark:bg-ink-950/60"
		>
			<img
				src={data.posterUrl}
				alt="生成的分享卡片预览"
				class="max-h-[49vh] w-auto rounded-sm object-contain shadow-lg"
			/>
		</div>
	{/if}

	<p class="text-center text-xs leading-5 text-ink-500 md:text-left">
		<span class="md:hidden">可以直接分享，也可以保存图片。</span>
		<span class="hidden md:inline">保存图片后，即可发送给朋友。</span>
	</p>

	<div class="grid grid-cols-2 gap-2 md:grid-cols-1">
		<button
			type="button"
			class="flex items-center justify-center gap-2 rounded-default bg-ink-900 px-4 py-3 text-sm text-white transition-colors active:scale-[0.99] dark:bg-jade-600 md:hidden"
			onclick={handleShare}
		>
			<Share2 size={16} /> 分享图片
		</button>
		<button
			type="button"
			class="flex items-center justify-center gap-2 rounded-default border border-ink-200 bg-white/50 px-4 py-3 text-sm text-ink-700 transition-colors active:scale-[0.99] dark:border-ink-700 dark:bg-ink-900/40 dark:text-ink-200 md:bg-ink-900 md:text-white md:hover:bg-jade-700 md:dark:bg-jade-600"
			onclick={handleSave}
		>
			<Download size={16} /> 保存图片
		</button>
	</div>
</div>
