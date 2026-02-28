<script lang="ts">
	import { postDetailCtx } from '$lib/features/post/context';
	import LeadIn from '$lib/ui/detail/LeadIn.svelte';

	const leadInStore = postDetailCtx.selectModelData((data) => data?.leadIn ?? '');
</script>

<LeadIn content={$leadInStore} label="引入" />