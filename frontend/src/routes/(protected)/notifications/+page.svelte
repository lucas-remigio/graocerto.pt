<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { Settings, ArrowLeft, Bell } from 'lucide-svelte';
	import { t } from '$lib/i18n';
	import { dataService } from '$lib/services/dataService';
	import { toastStore } from '$lib/stores/toast';
	import type {
		NotificationItem,
		NotificationPreferences,
		UpdateNotificationPreferencesPayload
	} from '$lib/types';
	import { formatCurrency } from '$lib/utils/currency';
	import { refreshUnreadNotificationCount } from '$lib/stores/notifications';

	let loading = $state(true);
	let savingPreferences = $state(false);
	let notifications: NotificationItem[] = $state([]);
	let preferences: NotificationPreferences | null = $state(null);
	let minDebitInput = $state('');
	let notifyDaysAheadInput = $state('1');
	let pushSupported = $state(false);
	let pushPermission = $state('default');
	let isRegisteringPush = $state(false);
	let currentEndpoint = $state<string | null>(null);

	const vapidKey = import.meta.env.VITE_PUBLIC_VAPID_KEY;

	let currentView: 'list' | 'settings' = $state('list');

	async function getActiveSubscription() {
		if (!('serviceWorker' in navigator)) return null;
		try {
			const registration = await navigator.serviceWorker.ready;
			const subscription = await registration.pushManager.getSubscription();
			return subscription;
		} catch (err) {
			console.error('Error getting subscription:', err);
			return null;
		}
	}

	async function checkPushStatus() {
		pushSupported = 'serviceWorker' in navigator && 'PushManager' in window;
		if (typeof Notification !== 'undefined') {
			pushPermission = Notification.permission;
		}

		if (pushSupported) {
			const sub = await getActiveSubscription();
			currentEndpoint = sub?.endpoint || null;
		}
	}

	function urlBase64ToUint8Array(base64String: string) {
		const padding = '='.repeat((4 - (base64String.length % 4)) % 4);
		const base64 = (base64String + padding).replace(/-/g, '+').replace(/_/g, '/');
		const rawData = window.atob(base64);
		const outputArray = new Uint8Array(rawData.length);
		for (let i = 0; i < rawData.length; ++i) {
			outputArray[i] = rawData.charCodeAt(i);
		}
		return outputArray;
	}

	async function handleRegisterPush() {
		if (!pushSupported || !vapidKey) {
			if (!vapidKey) console.error('VITE_PUBLIC_VAPID_KEY is missing');
			return;
		}
		isRegisteringPush = true;
		try {
			const permission = await Notification.requestPermission();
			pushPermission = permission;
			if (permission !== 'granted') {
				toastStore.error($t('notifications.push-permission-denied'));
				return;
			}

			const registration = await navigator.serviceWorker.ready;
			const subscription = await registration.pushManager.subscribe({
				userVisibleOnly: true,
				applicationServerKey: urlBase64ToUint8Array(vapidKey)
			});

			const subData = subscription.toJSON();
			if (subData.endpoint && subData.keys?.p256dh && subData.keys?.auth) {
				await dataService.registerPushSubscription({
					endpoint: subData.endpoint,
					p256dh: subData.keys.p256dh,
					auth: subData.keys.auth
				});
				currentEndpoint = subData.endpoint;
				// Refresh preferences to get updated push_endpoints list
				preferences = await dataService.fetchNotificationPreferences();
				toastStore.success($t('notifications.push-enabled-success'));
			}
		} catch (err) {
			console.error('Failed to register push:', err);
			toastStore.error($t('errors.server-error'));
		} finally {
			isRegisteringPush = false;
		}
	}

	async function handleTestPush() {
		try {
			await dataService.testPushNotification();
			toastStore.success('Solicitação de teste enviada!');
		} catch (err) {
			console.error(err);
			toastStore.error($t('errors.server-error'));
		}
	}

	let unreadCount = $derived(notifications.filter((n) => !n.is_read).length);

	let isActuallyRegistered = $derived(
		pushPermission === 'granted' &&
			currentEndpoint !== null &&
			(preferences?.push_endpoints || []).includes(currentEndpoint)
	);

	function buildNotificationMessage(notification: NotificationItem): string {
		if (notification.type !== 'recurring_due_tomorrow') return '';

		const hasDebits = notification.debit_count > 0;
		const hasCredits = notification.credit_count > 0;

		if (notification.notify_days_ahead <= 1) {
			if (hasDebits && hasCredits) {
				return $t('notifications.recurring-due-tomorrow-both', {
					values: {
						debitCount: notification.debit_count,
						debitTotal: formatCurrency(notification.total_debit),
						creditCount: notification.credit_count,
						creditTotal: formatCurrency(notification.total_credit)
					}
				});
			}
			if (hasCredits) {
				return $t('notifications.recurring-due-tomorrow-credits', {
					values: {
						count: notification.credit_count,
						total: formatCurrency(notification.total_credit)
					}
				});
			}
			return $t('notifications.recurring-due-tomorrow-debits', {
				values: {
					count: notification.debit_count,
					total: formatCurrency(notification.total_debit)
				}
			});
		}

		if (hasDebits && hasCredits) {
			return $t('notifications.recurring-due-in-days-both', {
				values: {
					days: notification.notify_days_ahead,
					debitCount: notification.debit_count,
					debitTotal: formatCurrency(notification.total_debit),
					creditCount: notification.credit_count,
					creditTotal: formatCurrency(notification.total_credit)
				}
			});
		}
		if (hasCredits) {
			return $t('notifications.recurring-due-in-days-credits', {
				values: {
					days: notification.notify_days_ahead,
					count: notification.credit_count,
					total: formatCurrency(notification.total_credit)
				}
			});
		}
		return $t('notifications.recurring-due-in-days-debits', {
			values: {
				days: notification.notify_days_ahead,
				count: notification.debit_count,
				total: formatCurrency(notification.total_debit)
			}
		});
	}

	function buildNotificationTitle(notification: NotificationItem): string {
		if (notification.notify_days_ahead <= 1) {
			return $t('notifications.recurring-due-tomorrow-title');
		}
		return $t('notifications.recurring-due-in-days-title', {
			values: { days: notification.notify_days_ahead }
		});
	}

	function getTargetDate(notification: NotificationItem): Date | null {
		if (!notification.target_date) return null;
		const parsed = new Date(notification.target_date);
		return Number.isNaN(parsed.getTime()) ? null : parsed;
	}

	function formatTargetDate(notification: NotificationItem): string {
		const target = getTargetDate(notification);
		if (!target) return '';
		return target.toLocaleDateString(undefined, {
			day: 'numeric',
			month: 'long'
		});
	}

	async function loadData() {
		loading = true;
		try {
			const [fetchedNotifications, fetchedPreferences] = await Promise.all([
				dataService.fetchNotifications(),
				dataService.fetchNotificationPreferences()
			]);
			notifications = fetchedNotifications;
			preferences = fetchedPreferences;
			notifyDaysAheadInput = String(fetchedPreferences.notify_days_ahead || 1);
			minDebitInput =
				fetchedPreferences.min_total_debit !== null &&
				fetchedPreferences.min_total_debit !== undefined
					? String(fetchedPreferences.min_total_debit)
					: '';
		} catch (err) {
			console.error(err);
			toastStore.error($t('errors.failed-load-data'));
		} finally {
			loading = false;
		}
	}

	async function handleMarkAsRead(notificationId: number) {
		try {
			await dataService.markNotificationAsRead(notificationId);
			notifications = notifications.map((notification) =>
				notification.id === notificationId ? { ...notification, is_read: true } : notification
			);
			refreshUnreadNotificationCount(true);
		} catch (err) {
			console.error(err);
			toastStore.error($t('errors.server-error'));
		}
	}

	async function handleOpenNotification(notification: NotificationItem) {
		if (!notification.account_token || !notification.target_date) {
			await handleMarkAsRead(notification.id);
			return;
		}

		const target = getTargetDate(notification);
		const month = target ? target.getMonth() + 1 : undefined;
		const year = target ? target.getFullYear() : undefined;
		await handleMarkAsRead(notification.id);
		await goto(
			`/home?account=${encodeURIComponent(notification.account_token)}${
				month && year ? `&month=${month}&year=${year}` : ''
			}`
		);
	}

	async function handleSavePreferences() {
		if (!preferences) return;
		savingPreferences = true;
		try {
			const minTotalDebit =
				minDebitInput.trim() === '' ? null : Math.max(0, Number.parseFloat(minDebitInput));
			const payload: UpdateNotificationPreferencesPayload = {
				enabled: preferences.enabled,
				notify_days_ahead: Number.parseInt(notifyDaysAheadInput, 10) || 1,
				min_total_debit: Number.isNaN(minTotalDebit as number) ? null : minTotalDebit
			};
			preferences = await dataService.updateNotificationPreferences(payload);
			notifyDaysAheadInput = String(preferences.notify_days_ahead || 1);

			// Show global success toast
			toastStore.success($t('notifications.preferences-saved'));
		} catch (err) {
			console.error(err);
			toastStore.error($t('errors.server-error'));
		} finally {
			savingPreferences = false;
		}
	}

	async function handleToggleEnabled() {
		if (!preferences) return;
		try {
			const payload: UpdateNotificationPreferencesPayload = {
				enabled: !preferences.enabled,
				notify_days_ahead: preferences.notify_days_ahead || 1,
				min_total_debit: preferences.min_total_debit ?? null
			};
			preferences = await dataService.updateNotificationPreferences(payload);
		} catch (err) {
			console.error(err);
			toastStore.error($t('errors.server-error'));
		}
	}

	onMount(() => {
		loadData();
		checkPushStatus();
	});
</script>

<div class="container mx-auto max-w-3xl p-4 sm:p-6 lg:p-8">
	<div class="mb-8 flex items-start justify-between">
		<div class="flex items-center gap-4">
			{#if currentView === 'settings'}
				<button
					class="btn btn-circle btn-ghost border border-base-300 bg-base-100 shadow-sm transition-all hover:bg-base-200"
					onclick={() => (currentView = 'list')}
					aria-label="Back"
				>
					<ArrowLeft size={18} />
				</button>
			{/if}
			<div>
				<h1 class="text-2xl font-bold tracking-tight">
					{currentView === 'settings'
						? $t('notifications.preferences-title')
						: $t('notifications.title')}
				</h1>
				{#if currentView === 'list'}
					<p class="mt-1 text-sm font-medium text-base-content/60">
						{$t('notifications.unread-count', { values: { count: unreadCount } })}
					</p>
				{/if}
			</div>
		</div>

		{#if currentView === 'list'}
			<div class="tooltip tooltip-left" data-tip={$t('notifications.preferences-title')}>
				<button
					class="btn btn-circle btn-ghost border border-base-300 bg-base-100 shadow-sm transition-all hover:bg-base-200"
					onclick={() => (currentView = 'settings')}
					aria-label="Settings"
				>
					<Settings size={20} class="text-center text-base-content/80" />
				</button>
			</div>
		{/if}
	</div>

	{#if loading}
		<div class="flex justify-center py-12">
			<span class="loading loading-spinner loading-lg text-primary"></span>
		</div>
	{:else if currentView === 'settings'}
		<div class="mb-6 overflow-hidden rounded-2xl border border-base-300 bg-base-100 shadow-sm">
			<div class="border-b border-base-300 bg-base-200/30 p-5 sm:px-8">
				<div class="flex items-center justify-between">
					<div class="pr-4">
						<h3 class="text-lg font-semibold text-base-content">
							{$t('notifications.enable-push')}
						</h3>
						<p class="mt-1 text-sm leading-relaxed text-base-content/60">
							{$t('notifications.enable-push-desc')}
						</p>
					</div>
					<label class="label cursor-pointer p-0">
						<input
							type="checkbox"
							class="toggle toggle-success toggle-lg shadow-sm"
							checked={preferences?.enabled ?? true}
							onchange={handleToggleEnabled}
						/>
					</label>
				</div>
			</div>

			<div class="border-b border-base-300 bg-base-200/10 p-5 sm:px-8">
				<div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
					<div class="flex-1 pr-4">
						<h3 class="text-lg font-semibold text-base-content">
							{$t('notifications.device-push')}
						</h3>
						<p class="mt-1 text-sm leading-relaxed text-base-content/60">
							{$t('notifications.device-push-desc')}
						</p>
					</div>
					<div class="flex items-center sm:shrink-0">
						{#if pushPermission === 'granted' && isActuallyRegistered}
							<div
								class="flex items-center gap-2.5 rounded-full bg-success/10 px-4 py-2 text-success ring-1 ring-inset ring-success/20 shadow-sm"
							>
								<span class="relative flex h-2.5 w-2.5">
									<span
										class="absolute inline-flex h-full w-full animate-ping rounded-full bg-success opacity-75"
									></span>
									<span class="relative inline-flex h-2.5 w-2.5 rounded-full bg-success"></span>
								</span>
								<span class="whitespace-nowrap text-xs font-bold uppercase tracking-widest">
									{$t('notifications.push-enabled')}
								</span>
							</div>
						{:else if pushPermission === 'granted' && !isActuallyRegistered}
							<button
								class="btn btn-warning btn-md w-full shadow-md transition-all hover:shadow-lg sm:w-auto"
								onclick={handleRegisterPush}
								disabled={isRegisteringPush}
							>
								{#if isRegisteringPush}
									<span class="loading loading-spinner loading-xs"></span>
								{/if}
								Sincronizar Dispositivo
							</button>
						{:else if !pushSupported}
							<div class="badge badge-ghost badge-md px-4 py-3 text-xs opacity-50 shadow-sm">
								{$t('notifications.push-not-supported')}
							</div>
						{:else}
							<button
								class="btn btn-primary btn-md w-full text-base-100 shadow-md transition-all hover:shadow-lg sm:w-auto"
								onclick={handleRegisterPush}
								disabled={isRegisteringPush}
							>
								{#if isRegisteringPush}
									<span class="loading loading-spinner loading-xs"></span>
								{/if}
								{$t('notifications.enable-on-device')}
							</button>
						{/if}
					</div>
				</div>

				{#if pushPermission === 'granted'}
					<div class="mt-4 flex justify-end border-t border-base-300/50 pt-4">
						<button class="btn btn-ghost btn-sm gap-2 text-base-content/60" onclick={handleTestPush}>
							<svg
								xmlns="http://www.w3.org/2000/svg"
								width="16"
								height="16"
								viewBox="0 0 24 24"
								fill="none"
								stroke="currentColor"
								stroke-width="2"
								stroke-linecap="round"
								stroke-linejoin="round"
								class="lucide lucide-send"
								><path d="m22 2-7 20-4-9-9-4Z" /><path d="M22 2 11 13" /></svg
							>
							Enviar Notificação de Teste
						</button>
					</div>
				{/if}
			</div>

			<div
				class="space-y-8 p-5 transition-opacity duration-200 sm:p-8"
				class:opacity-60={!(preferences?.enabled ?? true)}
				class:pointer-events-none={!(preferences?.enabled ?? true)}
			>
				<label class="form-control w-full space-y-1">
					<div class="label px-0">
						<span class="label-text text-base font-medium"
							>{$t('notifications.notify-days-ahead')}</span
						>
					</div>
					<input
						type="number"
						min="1"
						max="30"
						step="1"
						class="input input-bordered w-full max-w-sm shadow-sm transition-colors focus:border-primary focus:ring-1 focus:ring-primary"
						bind:value={notifyDaysAheadInput}
					/>
					<div class="label px-0">
						<span class="label-text-alt text-xs text-base-content/60"
							>{$t('notifications.notify-days-ahead-desc')}</span
						>
					</div>
				</label>

				<div class="divider"></div>

				<label class="form-control w-full space-y-1">
					<div class="label px-0">
						<span class="label-text text-base font-medium"
							>{$t('notifications.min-debit-threshold')} {$t('notifications.optional')}</span
						>
					</div>
					<div class="relative flex max-w-sm items-center">
						<span class="absolute left-4 font-medium text-base-content/50">€</span>
						<input
							type="number"
							min="0"
							step="0.01"
							class="input input-bordered w-full pl-10 shadow-sm transition-colors focus:border-primary focus:ring-1 focus:ring-primary"
							bind:value={minDebitInput}
							placeholder="0.00"
						/>
					</div>
					<div class="label px-0">
						<span class="label-text-alt max-w-lg text-xs leading-relaxed text-base-content/60"
							>{$t('notifications.min-debit-threshold-desc')}</span
						>
					</div>
				</label>

				<div class="flex items-center justify-end pt-4">
					<button
						class="btn btn-primary w-full min-w-[140px] font-medium text-base-100 shadow-lg sm:w-auto"
						disabled={savingPreferences}
						onclick={handleSavePreferences}
					>
						{#if savingPreferences}
							<span class="loading loading-spinner loading-xs"></span> {$t('common.loading')}
						{:else}
							{$t('common.save')}
						{/if}
					</button>
				</div>
			</div>
		</div>
	{:else if notifications.length === 0}
		<div
			class="flex flex-col items-center justify-center rounded-2xl border border-dashed border-base-300 bg-base-100/50 py-16 text-center shadow-sm"
		>
			<div class="mb-4 rounded-full bg-base-200/80 p-5 text-base-content/30 ring-8 ring-base-100">
				<Bell size={32} />
			</div>
			<h3 class="text-lg font-semibold tracking-tight text-base-content">
				{$t('notifications.catch-up-title')}
			</h3>
			<p class="mt-2 max-w-xs text-sm font-medium leading-relaxed text-base-content/60">
				{$t('notifications.empty')}
			</p>
		</div>
	{:else}
		<div class="flex flex-col gap-3">
			{#each notifications as notification (notification.id)}
				<div
					class="group relative rounded-2xl border p-4 shadow-sm transition-all duration-200 hover:border-primary/20 hover:shadow-md sm:p-5 {notification.is_read
						? 'border-base-300 bg-base-100'
						: 'border-primary/40 bg-primary/[0.04]'}"
				>
					<div class="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
						<button class="flex-1 text-left" onclick={() => handleOpenNotification(notification)}>
							<div class="mb-1 flex items-center gap-3">
								<span class="font-semibold tracking-tight text-base-content"
									>{buildNotificationTitle(notification)}</span
								>
								{#if !notification.is_read}
									<div
										class="flex items-center gap-1.5 rounded-full border border-primary/20 bg-primary/10 px-2 py-0.5"
									>
										<span class="h-1.5 w-1.5 rounded-full bg-primary"></span>
										<span class="text-[10px] font-bold uppercase tracking-wider text-primary"
											>{$t('notifications.new')}</span
										>
									</div>
								{/if}
							</div>
							<p class="max-w-xl text-sm font-medium leading-relaxed text-base-content/70">
								{buildNotificationMessage(notification)}
							</p>
							{#if notification.target_date}
								<p
									class="mt-3 flex items-center gap-1.5 text-[0.8rem] font-semibold uppercase tracking-wide text-base-content/50"
								>
									<span class="inline-block h-4 w-4 opacity-70">
										<svg
											xmlns="http://www.w3.org/2000/svg"
											viewBox="0 0 24 24"
											fill="none"
											stroke="currentColor"
											stroke-width="2"
											stroke-linecap="round"
											stroke-linejoin="round"
											class="lucide lucide-calendar"
											><path d="M8 2v4" /><path d="M16 2v4" /><rect
												width="18"
												height="18"
												x="3"
												y="4"
												rx="2"
											/><path d="M3 10h18" /></svg
										>
									</span>
									{$t('notifications.target-date', {
										values: { date: formatTargetDate(notification) }
									})}
								</p>
							{/if}
						</button>
						<div class="flex items-center sm:shrink-0">
							<button
								class="btn {notification.is_read
									? 'btn-ghost text-base-content/40'
									: 'btn-outline border-base-300'} btn-sm w-full font-medium transition-colors sm:w-auto"
								disabled={notification.is_read}
								onclick={() => handleMarkAsRead(notification.id)}
							>
								{$t('notifications.mark-as-read')}
							</button>
						</div>
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>
