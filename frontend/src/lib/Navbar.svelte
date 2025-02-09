<script lang="ts">
	import { Search, Bell, LogOut, Menu } from 'lucide-svelte';
	let isDropdownOpen = false;

	let categoriesUrl = '/categories';

	const logout = async () => {
		localStorage.removeItem('authToken');
		document.cookie = 'authToken=; Path=/; Expires=Thu, 01 Jan 1970 00:00:01 GMT;';

		window.location.href = '/login';
	};
</script>

<div class="navbar bg-base-100">
	<div class="navbar-start">
		<div class="dropdown">
			<div
				tabindex="0"
				role="button"
				class="btn btn-ghost lg:hidden"
				on:click={() => (isDropdownOpen = !isDropdownOpen)}
				on:keydown={(e) => {
					if (e.key === 'Enter' || e.key === ' ') {
						e.preventDefault();
						isDropdownOpen = !isDropdownOpen;
					}
				}}
			>
				<Menu size={20} class="h-5 w-5" />
			</div>
			{#if isDropdownOpen}
				<ul
					class="menu menu-sm dropdown-content bg-base-100 rounded-box z-[1] mt-3 w-52 p-2 shadow"
				>
					<li><a href={categoriesUrl} class="text-lg">Categories</a></li>
				</ul>
			{/if}
		</div>
		<a href="/" class="btn btn-ghost text-xl">Wallet Tracker</a>
	</div>
	<div class="navbar-center hidden lg:flex">
		<ul class="menu menu-horizontal px-1">
			<li><a href={categoriesUrl}>Categories</a></li>
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
