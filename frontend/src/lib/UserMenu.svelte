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
		class="btn btn-circle btn-ghost"
		on:click={() => (isProfileDropdownOpen = !isProfileDropdownOpen)}
		aria-haspopup="true"
		aria-expanded={isProfileDropdownOpen}
		aria-label="User menu"
	>
		<User size={20} class="h-5 w-5" />
	</button>

	{#if isProfileDropdownOpen}
		<ul class="menu dropdown-content z-[100] mt-4 w-64 rounded-box bg-base-100 p-4 shadow">
			<!-- Profile Button with Email -->
			<li class="mb-2">
				<button
					class="group flex w-full items-center justify-between rounded-lg p-2 transition-colors hover:bg-base-200"
					on:click={openProfile}
				>
					<div class="flex items-center gap-2">
						<User class="h-4 w-4 text-primary" />
						<div class="text-left">
							<div class="text-xs font-medium text-base-content/60">
								{$t('profile.view-profile', { default: 'View Profile' })}
							</div>
							<div class="max-w-[180px] truncate text-sm text-base-content">
								{localStorage.getItem('userEmail') || 'unknown@anonymous.pt'}
							</div>
						</div>
					</div>
					<ChevronRight
						class="h-4 w-4 text-base-content/40 transition-colors group-hover:text-primary"
					/>
				</button>
			</li>

			<slot />

			<li class="mt-2 border-t border-base-200 pt-2">
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
