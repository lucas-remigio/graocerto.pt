import { writable } from 'svelte/store';
import { dataService } from '$lib/services/dataService';

export const unreadNotificationCount = writable(0);
let hasFetchedInitial = false;

export async function refreshUnreadNotificationCount(force = false) {
	if (hasFetchedInitial && !force) return;

	try {
		const count = await dataService.fetchUnreadNotificationCount();
		unreadNotificationCount.set(count);
		hasFetchedInitial = true;
	} catch (error) {
		console.error('Failed to fetch unread notification count:', error);
	}
}
