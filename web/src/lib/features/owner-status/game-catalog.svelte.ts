import { browser } from '$app/environment';
import { lookupGame } from '$lib/features/owner-status/api';
import type { GameLookupResult } from '$lib/features/owner-status/types';

export type GameLookupState =
	{ status: 'idle' | 'loading' } | { status: 'done'; data: GameLookupResult } | { status: 'error' };

/**
 * 服务端游戏查询的模块级缓存。
 *
 * 本地映射表未命中时，按 process 名称懒查一次 `/public/game-lookup`，
 * 结果（含未命中）都缓存起来，避免 hover 卡片反复打开时重复请求。
 */
class GameCatalogStore {
	entries = $state<Record<string, GameLookupState>>({});

	/** 确保指定名称已触发查询；可重复调用，重复名称自动去重。 */
	ensure(name: string) {
		if (!browser) return;
		const key = name.trim().toLowerCase();
		if (!key) return;
		const existing = this.entries[key];
		if (existing && existing.status !== 'error') return;

		this.entries = { ...this.entries, [key]: { status: 'loading' } };
		lookupGame(name)
			.then((data) => {
				this.entries = { ...this.entries, [key]: { status: 'done', data } };
			})
			.catch(() => {
				this.entries = { ...this.entries, [key]: { status: 'error' } };
			});
	}

	get(name: string): GameLookupState {
		const key = name.trim().toLowerCase();
		return this.entries[key] ?? { status: 'idle' };
	}
}

export const gameCatalogStore = new GameCatalogStore();
