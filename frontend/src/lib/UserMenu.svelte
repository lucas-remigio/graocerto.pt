<script lang="ts">
	import { ChevronRight, LogOut, User } from 'lucide-svelte';
	import { t } from '$lib/i18n';
	import ProfileModal from './ProfileModal.svelte';

	export let logout: () => void;

	let isProfileDropdownOpen = false;
	let showProfileModal = false;

	function handleLogout() {
		isProfileDropdownOpen = false;
		logout();
	}

	function openProfile() {
		isProfileDropdownOpen = false;
		showProfileModal = true;
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
			<!-- Profile Button with Email -->
			<li class="mb-2">
				<button
					class="hover:bg-base-200 group flex w-full items-center justify-between rounded-lg p-2 transition-colors"
					on:click={openProfile}
				>
					<div class="flex items-center gap-2">
						<User class="text-primary h-4 w-4" />
						<div class="text-left">
							<div class="text-base-content/60 text-xs font-medium">
								{$t('profile.view-profile', { default: 'View Profile' })}
							</div>
							<div class="text-base-content max-w-[180px] truncate text-sm">
								{localStorage.getItem('userEmail') || 'unknown@anonymous.pt'}
							</div>
						</div>
					</div>
					<ChevronRight
						class="text-base-content/40 group-hover:text-primary h-4 w-4 transition-colors"
					/>
				</button>
			</li>

			<slot />

			<li class="border-base-200 mt-2 border-t pt-2">
				<button class="btn btn-error btn-sm w-full" on:click={handleLogout}>
					<LogOut size={18} class="mr-2" />
					{$t('auth.logout', { default: 'Logout' })}
				</button>
			</li>
		</ul>
	{/if}
</div>

<!-- Profile Modal -->
<ProfileModal bind:showModal={showProfileModal} {logout} />
