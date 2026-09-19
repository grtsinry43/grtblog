export type OwnerStatusMediaState = 'playing' | 'paused' | 'stopped' | '';

export type OwnerStatusMedia = {
	title?: string;
	artist?: string;
	thumbnail?: string;
	/** 当前播放进度（秒） */
	position?: number;
	/** 媒体总时长（秒） */
	duration?: number;
	/** 播放状态：playing / paused / stopped */
	state?: OwnerStatusMediaState;
};

/** 游戏卡片元数据，由上报端规则维护并随状态上报（category = game 时展示）。 */
export type OwnerStatusGame = {
	/** 游戏标准名：卡片标题 + Steam 搜索关键词（留空回退到 process） */
	name?: string;
	/** 封面横幅图 URL */
	cover?: string;
	/** 手写 slogan */
	slogan?: string;
	/** 手写一句话描述 */
	desc?: string;
	/** 点缀色（hex） */
	accent?: string;
	/** 点击跳转链接 */
	url?: string;
};

export type OwnerStatus = {
	ok: number;
	process?: string;
	extend?: string;
	/** 活动分类（如 game），来自上报端规则配置 */
	category?: string;
	/** 游戏卡片信息，来自上报端规则配置 */
	game?: OwnerStatusGame | null;
	media?: OwnerStatusMedia | null;
	timestamp?: number;
	adminPanelOnline?: boolean;
};

export type OwnerStatusRealtimePayload = OwnerStatus & {
	type?: string;
};

export type GameLookupResult = {
	found: boolean;
	source?: string;
	appId?: number;
	name?: string;
	shortDescription?: string;
	headerImage?: string;
	capsuleImage?: string;
	storeUrl?: string;
};
