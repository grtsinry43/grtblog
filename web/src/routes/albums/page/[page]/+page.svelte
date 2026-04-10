<script lang="ts">
	import AlbumListPage from '$lib/features/album/components/AlbumListPage.svelte';
	import type { PageData } from './$types';

	let { data } = $props<{ data: PageData }>();
</script>

<AlbumListPage albums={data.albums} staggerKey={`albums-${data.albums.page}`} />
