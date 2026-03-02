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
