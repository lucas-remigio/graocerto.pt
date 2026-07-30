<script lang="ts">
	// Tiny dependency-free sparkline (SVG polyline). Purely presentational.
	let {
		data = [],
		color = '#6366f1',
		width = 72,
		height = 22
	}: { data?: number[]; color?: string; width?: number; height?: number } = $props();

	let points = $derived.by(() => {
		if (data.length < 2) return '';
		const max = Math.max(...data, 1);
		const stepX = width / (data.length - 1);
		return data
			.map((v, i) => {
				const x = i * stepX;
				const y = height - (Math.max(v, 0) / max) * height;
				return `${x.toFixed(1)},${y.toFixed(1)}`;
			})
			.join(' ');
	});
</script>

{#if points}
	<svg {width} {height} viewBox="0 0 {width} {height}" class="shrink-0" aria-hidden="true">
		<polyline
			{points}
			fill="none"
			stroke={color}
			stroke-width="1.5"
			stroke-linecap="round"
			stroke-linejoin="round"
		/>
	</svg>
{/if}
