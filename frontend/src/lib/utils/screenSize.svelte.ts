import { onDestroy, onMount } from 'svelte';

const LARGE_SCREEN_MIN = 1024; // Tailwind `lg` breakpoint

/**
 * Reactive `isLargeScreen` flag, true at the Tailwind `lg` breakpoint and up,
 * kept in sync with window resizes. Call once during component init; it wires
 * and tears down its own resize listener.
 *
 * Usage: `const screen = useIsLargeScreen();` then read `screen.value`.
 */
export function useIsLargeScreen(): { readonly value: boolean } {
	const state = $state({ value: false });

	function update() {
		state.value = window.innerWidth >= LARGE_SCREEN_MIN;
	}

	onMount(() => {
		update();
		window.addEventListener('resize', update);
	});

	onDestroy(() => {
		if (typeof window !== 'undefined') window.removeEventListener('resize', update);
	});

	return state;
}
