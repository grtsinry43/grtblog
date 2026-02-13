<script lang="ts">
	import { browser } from '$app/environment';
	import { createMutation, useQueryClient } from '@tanstack/svelte-query';
	import { trackContentView } from '$lib/features/analytics/api';
	import type { TrackViewContentType } from '$lib/features/analytics/types';
	import { getOrCreateVisitorId, syncVisitorId } from '$lib/shared/visitor/visitor-id';

	const trackedViewKeys = new Map<string, number>();
	const dedupWindowMs = 30_000;

	interface Props {
		contentType: TrackViewContentType;
		contentId: number;
	}
	type TrackViewMutationInput = {
		contentType: TrackViewContentType;
		contentId: number;
	};

	let { contentType, contentId }: Props = $props();
	let inFlightKey = $state<string | null>(null);

	const queryClient = useQueryClient();

	const mutation = createMutation(() => ({
		mutationFn: async (input: TrackViewMutationInput) => {
			const visitorId = getOrCreateVisitorId();
			return trackContentView(undefined, {
				contentType: input.contentType,
				contentId: input.contentId,
				visitorId: visitorId || undefined
			});
		},
		retry: false,
		onSuccess: (result, input) => {
			syncVisitorId(result?.visitorId);
			queryClient.setQueryData(['analytics', 'view', input.contentType, input.contentId], true);
		},
		onSettled: () => {
			inFlightKey = null;
		}
	}));

	$effect(() => {
		if (!browser || contentId <= 0) return;
		const trackKey = `${contentType}:${contentId}`;
		const now = Date.now();
		const lastTrackedAt = trackedViewKeys.get(trackKey) ?? 0;
		if (now - lastTrackedAt < dedupWindowMs) return;
		if (inFlightKey === trackKey) return;
		const queryKey = ['analytics', 'view', contentType, contentId] as const;
		if (queryClient.getQueryData(queryKey)) {
			trackedViewKeys.set(trackKey, now);
			return;
		}
		inFlightKey = trackKey;
		trackedViewKeys.set(trackKey, now);
		mutation.mutate({
			contentType,
			contentId
		});
	});
</script>
