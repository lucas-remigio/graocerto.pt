<script lang="ts">
	import { t } from '$lib/i18n';
	import { User, Mail, Trash2, Download, X } from 'lucide-svelte';
	import api_axios from './axios';

	export let showModal = false;
	export let logout: () => void;

	let showDeleteConfirmation = false;
	let isDeleting = false;
	let deleteConfirmText = '';

	const userEmail = localStorage.getItem('userEmail') || 'unknown@anonymous.pt';
	const userCreated = localStorage.getItem('userCreated') || 'Unknown';

	async function handleDeleteAccount() {
		if (deleteConfirmText !== 'DELETE') {
			return;
		}

		isDeleting = true;
		try {
			const response = await api_axios.delete('auth/delete-account');

			if (response.status === 200) {
				// Clear all local storage and redirect
				logout();
				localStorage.clear();
				window.location.href = '/';
			} else {
				const error = await response.data;
				alert(`Error deleting account: ${error}`);
			}
		} catch (error) {
			console.error('Delete account error:', error);
			alert('Failed to delete account. Please try again.');
		} finally {
			isDeleting = false;
		}
	}

	async function handleExportData() {
		try {
			const token = localStorage.getItem('token');
			const response = await fetch('/api/v1/auth/export-data', {
				method: 'GET',
				headers: {
					Authorization: `Bearer ${token}`
				}
			});

			if (response.ok) {
				const blob = await response.blob();
				const url = window.URL.createObjectURL(blob);
				const a = document.createElement('a');
				a.href = url;
				a.download = `grao-certo-data-${new Date().toISOString().split('T')[0]}.json`;
				document.body.appendChild(a);
				a.click();
				document.body.removeChild(a);
				window.URL.revokeObjectURL(url);
			} else {
				alert('Failed to export data. Please try again.');
			}
		} catch (error) {
			console.error('Export data error:', error);
			alert('Failed to export data. Please try again.');
		}
	}

	function closeModal() {
		showModal = false;
		showDeleteConfirmation = false;
		deleteConfirmText = '';
	}
</script>

{#if showModal}
	<!-- Modal backdrop -->
	<div class="modal modal-open">
		<div class="modal-box max-w-md">
			<!-- Header -->
			<div class="mb-6 flex items-center justify-between">
				<h3 class="flex items-center gap-2 text-lg font-bold">
					<User class="text-primary h-5 w-5" />
					{$t('profile.title', { default: 'Profile' })}
				</h3>
				<button class="btn btn-ghost btn-sm btn-circle" on:click={closeModal}>
					<X class="h-4 w-4" />
				</button>
			</div>

			{#if !showDeleteConfirmation}
				<!-- Profile Info -->
				<div class="mb-6 space-y-4">
					<div class="bg-base-200 rounded-lg p-4">
						<div class="mb-2 flex items-center gap-2">
							<Mail class="text-primary h-4 w-4" />
							<span class="text-sm font-medium">{$t('profile.email', { default: 'Email' })}</span>
						</div>
						<p class="text-base-content/80">{userEmail}</p>
					</div>

					<div class="bg-base-200 rounded-lg p-4">
						<div class="mb-2 flex items-center gap-2">
							<User class="text-primary h-4 w-4" />
							<span class="text-sm font-medium"
								>{$t('profile.member-since', { default: 'Member since' })}</span
							>
						</div>
						<p class="text-base-content/80">{userCreated}</p>
					</div>
				</div>

				<!-- Actions -->
				<div class="space-y-3">
					<!-- Export Data -->
					<button class="btn btn-outline w-full" on:click={handleExportData}>
						<Download class="mr-2 h-4 w-4" />
						{$t('profile.export-data', { default: 'Export My Data' })}
					</button>

					<!-- Delete Account -->
					<button
						class="btn btn-error btn-outline w-full"
						on:click={() => (showDeleteConfirmation = true)}
					>
						<Trash2 class="mr-2 h-4 w-4" />
						{$t('profile.delete-account', { default: 'Delete Account' })}
					</button>
				</div>
			{:else}
				<!-- Delete Confirmation -->
				<div class="space-y-4">
					<div class="alert alert-error">
						<Trash2 class="h-5 w-5" />
						<div>
							<h3 class="font-bold">
								{$t('profile.delete-warning.title', { default: 'Delete Account' })}
							</h3>
							<div class="text-xs">
								{$t('profile.delete-warning.description', {
									default:
										'This will permanently delete your account and all your data. This action cannot be undone.'
								})}
							</div>
						</div>
					</div>

					<div class="form-control">
						<label class="label" for="delete-confirm-input">
							<span class="label-text">
								{$t('profile.delete-warning.type-delete', { default: 'Type "DELETE" to confirm:' })}
							</span>
						</label>
						<input
							id="delete-confirm-input"
							type="text"
							bind:value={deleteConfirmText}
							placeholder="DELETE"
							class="input input-bordered"
							class:input-error={deleteConfirmText !== '' && deleteConfirmText !== 'DELETE'}
						/>
					</div>

					<div class="flex gap-2">
						<button
							class="btn btn-ghost flex-1"
							on:click={() => (showDeleteConfirmation = false)}
							disabled={isDeleting}
						>
							{$t('common.cancel', { default: 'Cancel' })}
						</button>
						<button
							class="btn btn-error flex-1"
							on:click={handleDeleteAccount}
							disabled={isDeleting || deleteConfirmText !== 'DELETE'}
						>
							{#if isDeleting}
								<span class="loading loading-spinner loading-sm"></span>
							{:else}
								{$t('common.delete', { default: 'Delete' })}
							{/if}
						</button>
					</div>
				</div>
			{/if}
		</div>
	</div>
{/if}
