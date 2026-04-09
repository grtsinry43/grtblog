import { browser } from '$app/environment';
import { toast } from 'svelte-sonner';
import { realtimeWSCore } from '$lib/shared/ws/realtime-core';
import type { SiteActivityPayload } from '$lib/features/realtime-activity/types';

class RealtimeActivityStore {
	private started = false;
	private unbindContent: (() => void) | null = null;
	private seen: Record<string, number> = {};

	start() {
		if (!browser || this.started) return;
		this.started = true;
		this.unbindContent = realtimeWSCore.onContent((payload: unknown) => {
			this.handlePayload(payload);
		});
		realtimeWSCore.start();
	}

	stop() {
		this.started = false;
		this.unbindContent?.();
		this.unbindContent = null;
		this.seen = {};
	}

	private handlePayload(payload: unknown) {
		if (!this.isSiteActivity(payload)) return;
		if (!this.started || !browser) return;

		const normalizedURL = this.normalizeURL(payload.url);
		if (!normalizedURL) return;

		const dedupeKey = `${payload.event}|${normalizedURL}|${payload.at}|${payload.title}|${payload.excerpt ?? ''}`;
		if (this.seen[dedupeKey]) return;
		this.seen[dedupeKey] = Date.now();
		this.pruneSeen();

		const copy = this.buildCopy(payload, dedupeKey);
		toast(copy.title, {
			description: copy.description,
			duration: 9000,
			classes: {
				actionButton:
					'inline-flex min-w-[3.5rem] shrink-0 items-center justify-center whitespace-nowrap'
			},
			action: {
				label: '去围观',
				onClick: () => {
					void this.openURL(normalizedURL);
				}
			}
		});
	}

	private isSiteActivity(payload: unknown): payload is SiteActivityPayload {
		if (!payload || typeof payload !== 'object') return false;
		if ((payload as { type?: string }).type !== 'site.activity') return false;

		const item = payload as Partial<SiteActivityPayload>;
		if (typeof item.event !== 'string' || typeof item.url !== 'string') return false;
		if (typeof item.title !== 'string' || typeof item.at !== 'string') return false;
		return true;
	}

	private normalizeURL(raw: string): string | null {
		const value = raw.trim();
		if (!value) return null;
		if (value.startsWith('/')) return value;
		if (/^https?:\/\//i.test(value)) return value;
		return null;
	}

	private pruneSeen() {
		const now = Date.now();
		for (const [key, timestamp] of Object.entries(this.seen)) {
			if (!Number.isFinite(timestamp) || now - timestamp > 15 * 60 * 1000) {
				delete this.seen[key];
			}
		}
	}

	private openURL(url: string) {
		if (!browser) return;
		window.location.href = url;
	}

	private buildCopy(
		payload: SiteActivityPayload,
		key: string
	): { title: string; description: string } {
		const title = this.fallbackTitle(payload.title, payload.contentType);
		const excerpt = (payload.excerpt || '').trim();

		switch (payload.event) {
			case 'article.updated':
				return {
					title: this.pick(['文章冒出新鲜热气', '文章刚刚二次发酵', '文章有了新版本'], key),
					description: `《${title}》刚被作者轻轻打磨，点开看看新变化。`
				};
			case 'article.hot_marked':
				return {
					title: this.pick(['热榜上新', '这篇文章冲上热门', '新晋热门文章出现'], key),
					description: `《${title}》刚被标记为热门，来看看它凭什么出圈。`
				};
			case 'moment.updated':
				return {
					title: this.pick(['手记冒泡啦', '手记有了新段落', '手记刚续上一笔'], key),
					description: `「${title}」更新完成，灵感还热乎着。`
				};
			case 'page.updated':
				return {
					title: this.pick(['页面悄悄焕新', '页面刚完成翻修', '页面有新内容上线'], key),
					description: `页面「${title}」已经更新，去看看最新版本。`
				};
			case 'thinking.updated':
				return {
					title: this.pick(['思考又长出了枝叶', '思考刚刚被续写', '一条思考有了后续'], key),
					description: `「${title}」补上了新想法，欢迎来围观。`
				};
			case 'comment.created':
				return {
					title: this.pick(['评论区有新动静', '有人留下了新评论', '评论区刚刚冒泡'], key),
					description: excerpt
						? `在「${title}」下出现新评论：${excerpt}`
						: `在「${title}」下出现了新评论，快去看看。`
				};
			case 'comment.approved':
				return {
					title: this.pick(['评论区有新动静', '有人留下了新评论', '评论区刚刚冒泡'], key),
					description: excerpt
						? `在「${title}」下出现新评论：${excerpt}`
						: `在「${title}」下出现了新评论，快去看看。`
				};
			default:
				return {
					title: '站点有新动态',
					description: `「${title}」有更新，去围观看看。`
				};
		}
	}

	private fallbackTitle(raw: string, contentType: string): string {
		const value = raw.trim();
		if (value) return value;
		switch (contentType) {
			case 'article':
				return '文章';
			case 'moment':
				return '手记';
			case 'page':
				return '页面';
			case 'thinking':
				return '思考';
			default:
				return '内容';
		}
	}

	private pick(candidates: string[], seed: string): string {
		if (candidates.length === 0) return '';
		let hash = 0;
		for (let i = 0; i < seed.length; i++) {
			hash = (hash * 31 + seed.charCodeAt(i)) >>> 0;
		}
		return candidates[hash % candidates.length] ?? candidates[0] ?? '';
	}
}

export const realtimeActivityStore = new RealtimeActivityStore();
