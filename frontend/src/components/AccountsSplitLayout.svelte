<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import type { Snippet } from 'svelte';
	import Accounts from '$components/Accounts.svelte';
	import type { Account } from '$lib/types';

	let {
		accounts = [],
		selectedAccount = null,
		isLargeScreen = false,
		accountsLoading = false,
		showRightPanel = true,
		children
	}: {
		accounts?: Account[];
		selectedAccount?: Account | null;
		isLargeScreen?: boolean;
		accountsLoading?: boolean;
		showRightPanel?: boolean;
		children?: Snippet;
	} = $props();

	const dispatch = createEventDispatcher<{
		select: { account: Account };
		updatedAccount: { account: Account };
		deleteAccount: { account: Account };
		newAccount: { account: Account };
	}>();
</script>

<div class="flex flex-col lg:flex-row">
	<div class="lg:w-80 lg:flex-shrink-0 lg:pr-6">
		<Accounts
			{accounts}
			{selectedAccount}
			isVertical={isLargeScreen}
			loading={accountsLoading}
			on:select={(event) => dispatch('select', event.detail)}
			on:updatedAccount={(event) => dispatch('updatedAccount', event.detail)}
			on:deleteAccount={(event) => dispatch('deleteAccount', event.detail)}
			on:newAccount={(event) => dispatch('newAccount', event.detail)}
		/>
	</div>

	<div class="hidden lg:block lg:w-px lg:bg-base-300"></div>

	{#if showRightPanel}
		<div class="flex-1 lg:flex lg:max-h-screen lg:flex-col lg:overflow-hidden lg:pl-6">
			{@render children?.()}
		</div>
	{/if}
</div>
