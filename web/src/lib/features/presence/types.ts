export type PresenceContentType = 'article' | 'moment' | 'page' | 'thinking';

export type PresenceClientReport = {
	contentType: PresenceContentType;
	url: string;
};

export type PresencePageItem = {
	contentType: PresenceContentType;
	title: string;
	url: string;
	connections: number;
};

export type PresenceSnapshotPayload = {
	type: 'presence.snapshot';
	online: number;
	pages: PresencePageItem[];
};
