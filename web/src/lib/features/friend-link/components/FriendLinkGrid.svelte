<script lang="ts">
	import type { FriendLink } from '../types';
	import FriendCard from './FriendCard.svelte';
	import { StaggerList } from '$lib/ui/animation';

	let { links } = $props<{ links: FriendLink[] }>();
</script>

<StaggerList class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 md:gap-8 lg:gap-10">
	{#each links as friend (friend.url)}
		<div class="friend-card-wrapper h-full">
			<FriendCard {friend} />
		</div>
	{/each}
</StaggerList>

<style lang="postcss">
	@reference "$routes/layout.css";
</style>
