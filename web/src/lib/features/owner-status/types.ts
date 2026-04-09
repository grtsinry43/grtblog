export type OwnerStatusMedia = {
	title?: string;
	artist?: string;
	thumbnail?: string;
};

export type OwnerStatus = {
	ok: number;
	process?: string;
	extend?: string;
	media?: OwnerStatusMedia | null;
	timestamp?: number;
	adminPanelOnline?: boolean;
};

export type OwnerStatusRealtimePayload = OwnerStatus & {
	type?: string;
};
