<script lang="ts">
	import { LogOut, User } from 'lucide-svelte';

	export let logout: () => void;

	let isProfileDropdownOpen = false;

	function handleLogout() {
		isProfileDropdownOpen = false;
		logout();
	}
</script>

<div class="dropdown dropdown-end {isProfileDropdownOpen ? 'dropdown-open' : ''}">
	<button
		class="btn btn-ghost btn-circle"
		on:click={() => (isProfileDropdownOpen = !isProfileDropdownOpen)}
		aria-haspopup="true"
		aria-expanded={isProfileDropdownOpen}
		aria-label="User menu"
	>
		<User size={20} class="h-5 w-5" />
	</button>

	{#if isProfileDropdownOpen}
		<ul class="dropdown-content menu bg-base-100 rounded-box z-[100] mt-4 w-64 p-4 shadow">
			<li class="mb-2 flex items-center justify-center">
				<span
					class="text-base-content select-text rounded px-2 py-1 text-center text-sm font-medium"
				>
					{localStorage.getItem('userEmail') || 'unknown@anonymous.pt'}
				</span>
			</li>

			<slot />

			<li class="border-base-200 mt-2 border-t pt-2">
				<button class="btn btn-error btn-sm w-full" on:click={handleLogout}>
					<LogOut size={18} class="mr-2" />
				</button>
			</li>
		</ul>
	{/if}
</div>
