<script lang="ts">
	import { goto } from '$app/navigation';
	import { Search, Bell, LogOut, Menu } from 'lucide-svelte';

	let isDropdownOpen = false;
	let categoriesUrl = '/categories';

	// Flag to indicate that a touch event already handled the toggle
	let touchHandled = false;

	const logout = async () => {
		localStorage.removeItem('authToken');
		document.cookie = 'authToken=; Path=/; Expires=Thu, 01 Jan 1970 00:00:01 GMT;';

		window.location.href = '/login';
	};

	function handleNavigation(url: string) {
		isDropdownOpen = false;
		goto(url);
	}
</script>

<div class="navbar bg-base-100">
	<div class="navbar-start">
		<div class="dropdown relative {isDropdownOpen ? 'dropdown-open' : ''}">
			<button
				type="button"
				class="btn btn-ghost lg:hidden"
				on:touchend={(event) => {
					// Prevent the synthetic click from firing after touchend
					event.preventDefault();
					event.stopPropagation();
					touchHandled = true;
					isDropdownOpen = !isDropdownOpen;
				}}
				on:click={(event) => {
					event.preventDefault();
					event.stopPropagation();
					// If the touch event already toggled the state, ignore this click
					if (touchHandled) {
						touchHandled = false;
						return;
					}
					isDropdownOpen = !isDropdownOpen;
				}}
			>
				<Menu size={20} class="h-5 w-5" />
			</button>
			{#if isDropdownOpen}
				<ul
					class="menu menu-sm dropdown-content bg-base-100 rounded-box z-[50] mt-3 w-52 p-2 shadow"
				>
					<li>
						<button
							type="button"
							on:click={() => handleNavigation(categoriesUrl)}
							class="text-lg"
							aria-label="Categories">Categories</button
						>
					</li>
				</ul>
			{/if}
		</div>
		<a href="/" class="btn btn-ghost text-xl">Wallet Tracker</a>
	</div>
	<div class="navbar-center hidden lg:flex">
		<ul class="menu menu-horizontal px-1">
			<li>
				<button
					type="button"
					on:click={() => handleNavigation(categoriesUrl)}
					class="text-lg"
					aria-label="Categories">Categories</button
				>
			</li>
		</ul>
	</div>
	<div class="navbar-end">
		<!-- Search Button -->
		<button aria-label="search" class="btn btn-ghost btn-circle">
			<Search size={20} class="h-5 w-5" />
		</button>

		<!-- Notifications Button -->
		<button aria-label="notifications" class="btn btn-ghost btn-circle">
			<div class="indicator">
				<Bell size={20} class="h-5 w-5" />
				<span class="badge badge-xs badge-primary indicator-item"></span>
			</div>
		</button>

		<!-- Logout Button -->
		<button aria-label="logout" class="btn btn-ghost" on:click={logout}>
			<div class="flex items-center space-x-2">
				<span class="hidden sm:inline">Logout</span>
				<LogOut size={20} class="h-5 w-5" />
			</div>
		</button>
	</div>
</div>

<style>
	.dropdown:not(.dropdown-open) .dropdown-content {
		display: none;
	}
</style>
