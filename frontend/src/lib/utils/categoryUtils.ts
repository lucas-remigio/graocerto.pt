import type { CategoryDto } from '$lib/types';

export interface CategoryGroup {
	parent: CategoryDto | null;
	children: CategoryDto[];
}

/**
 * Organises a flat category list into parent/children groups for hierarchical display.
 * Orphaned children (parent deleted or missing) are collected in a trailing group.
 */
export function buildCategoryGroups(cats: CategoryDto[]): CategoryGroup[] {
	const parents = cats.filter((c) => !c.parent_category_id);
	const children = cats.filter((c) => c.parent_category_id);

	const groups: CategoryGroup[] = parents.map((parent) => ({
		parent,
		children: children.filter((child) => child.parent_category_id === parent.id)
	}));

	const orphans = children.filter(
		(child) => !parents.some((p) => p.id === child.parent_category_id)
	);
	if (orphans.length > 0) {
		groups.push({ parent: null, children: orphans });
	}

	return groups;
}

/* ---- colour helpers (WCAG relative luminance) ---- */

function hexToRgb(hex: string): { r: number; g: number; b: number } {
	const h = hex.replace('#', '');
	return {
		r: parseInt(h.substring(0, 2), 16),
		g: parseInt(h.substring(2, 4), 16),
		b: parseInt(h.substring(4, 6), 16)
	};
}

function channelLuminance(color: number): number {
	const c = color / 255;
	return c <= 0.03928 ? c / 12.92 : Math.pow((c + 0.055) / 1.055, 2.4);
}

function relativeLuminance(hex: string): number {
	const { r, g, b } = hexToRgb(hex);
	return 0.2126 * channelLuminance(r) + 0.7152 * channelLuminance(g) + 0.0722 * channelLuminance(b);
}

/**
 * Returns a tailwind text-color class ('text-gray-900' | 'text-gray-100') that
 * contrasts against the given hex background colour, based on relative luminance.
 */
export function getContrastTextClass(backgroundColor: string): string {
	return relativeLuminance(backgroundColor) > 0.5 ? 'text-gray-900' : 'text-gray-100';
}
