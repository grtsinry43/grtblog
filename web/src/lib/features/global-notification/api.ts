import { getApi } from '$lib/shared/clients/api';
import type { GlobalNotificationItem } from '$lib/features/global-notification/types';

export function fetchActiveGlobalNotifications(svelteFetch?: typeof fetch) {
	return getApi(svelteFetch)<GlobalNotificationItem[]>('/public/global-notifications');
}
