import { writable } from 'svelte/store';
import { dataService } from '$lib/services/dataService';

export const unreadNotificationCount = writable(0);

export async function refreshUnreadNotificationCount() {
	try {
		const count = await dataService.fetchUnreadNotificationCount();
		unreadNotificationCount.set(count);
	} catch (error) {
		console.error('Failed to fetch unread notification count:', error);
	}
}
