export type GlobalNotificationItem = {
	id: number;
	content: string;
	publishAt: string;
	expireAt: string;
	allowClose: boolean;
	createdAt: string;
	updatedAt: string;
};

export type GlobalNotificationRealtimeUpsertPayload = {
	type: 'global.notification.created' | 'global.notification.updated';
	id: number;
	content: string;
	publishAt: string;
	expireAt: string;
	allowClose: boolean;
	at: string;
};

export type GlobalNotificationRealtimeDeletePayload = {
	type: 'global.notification.deleted';
	id: number;
};

export type GlobalNotificationRealtimePayload =
	| GlobalNotificationRealtimeUpsertPayload
	| GlobalNotificationRealtimeDeletePayload;
