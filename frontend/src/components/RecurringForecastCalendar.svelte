<script lang="ts">
	import type { CategoryDto, RecurringForecastItem } from '$lib/types';
	import { t } from '$lib/i18n';
	import { locale } from 'svelte-i18n';
	import { ArrowRightLeft } from 'lucide-svelte';
	import { formatCurrency } from '$lib/utils/currency';
	import { getContrastTextClass } from '$lib/utils/categoryUtils';
	import { TransactionTypeId } from '$lib/transaction_types_types';

	let {
		items = [],
		categories = [],
		days = 30
	}: { items?: RecurringForecastItem[]; categories?: CategoryDto[]; days?: number } = $props();

	let currentLocale = $derived($locale || 'pt');

	// --- date helpers (local, tz-safe) ---
	function ymd(d: Date): string {
		const y = d.getFullYear();
		const m = String(d.getMonth() + 1).padStart(2, '0');
		const day = String(d.getDate()).padStart(2, '0');
		return `${y}-${m}-${day}`;
	}
	function parseLocal(s: string): Date {
		const [y, m, d] = s.split('-').map(Number);
		return new Date(y, m - 1, d);
	}
	function addDays(s: string, n: number): string {
		const d = parseLocal(s);
		d.setDate(d.getDate() + n);
		return ymd(d);
	}

	const today = ymd(new Date());
	let windowEnd = $derived(addDays(today, days));

	// --- category helpers ---
	function getCategoryById(categoryId: number): CategoryDto | undefined {
		return categories.find((category) => category.id === categoryId);
	}
	function getParentCategoryName(category: CategoryDto | undefined): string | null {
		if (!category?.parent_category_id) return null;
		if (category.parent_category?.category_name) return category.parent_category.category_name;
		return getCategoryById(category.parent_category_id)?.category_name || null;
	}

	function signedAmount(item: RecurringForecastItem): number {
		return item.transaction_type_id === TransactionTypeId.Debit ? -item.amount : item.amount;
	}

	// --- group items by day ---
	interface DayData {
		items: RecurringForecastItem[];
		net: number;
	}
	let itemsByDay = $derived.by(() => {
		const map: Record<string, DayData> = {};
		for (const item of items) {
			const key = item.date?.split('T')[0] ?? '';
			if (!key) continue;
			(map[key] ??= { items: [], net: 0 }).items.push(item);
			map[key].net += signedAmount(item);
		}
		return map;
	});

	// --- calendar structure: months -> weeks (Monday-start) -> days ---
	interface MonthBlock {
		key: string;
		label: string;
		weeks: (string | null)[][];
	}

	function mondayIndex(dateStr: string): number {
		// JS getDay(): Sunday=0..Saturday=6 -> Monday=0..Sunday=6
		return (parseLocal(dateStr).getDay() + 6) % 7;
	}

	function monthDays(monthKey: string): string[] {
		const [y, m] = monthKey.split('-').map(Number);
		const last = new Date(y, m, 0).getDate();
		const out: string[] = [];
		for (let d = 1; d <= last; d++) {
			out.push(`${monthKey}-${String(d).padStart(2, '0')}`);
		}
		return out;
	}

	function toWeeks(daysList: string[]): (string | null)[][] {
		const weeks: (string | null)[][] = [];
		let week: (string | null)[] = [];
		for (const day of daysList) {
			if (week.length === 0) {
				const pad = mondayIndex(day);
				for (let i = 0; i < pad; i++) week.push(null);
			}
			week.push(day);
			if (week.length === 7) {
				weeks.push(week);
				week = [];
			}
		}
		if (week.length) {
			while (week.length < 7) week.push(null);
			weeks.push(week);
		}
		return weeks;
	}

	let months = $derived.by(() => {
		const result: MonthBlock[] = [];
		let cursor = today.slice(0, 7);
		const endMonth = windowEnd.slice(0, 7);
		// iterate months from today's month through windowEnd's month
		while (cursor <= endMonth) {
			const [y, m] = cursor.split('-').map(Number);
			result.push({
				key: cursor,
				label: new Date(y, m - 1, 1).toLocaleDateString(currentLocale, {
					month: 'long',
					year: 'numeric'
				}),
				weeks: toWeeks(monthDays(cursor))
			});
			// advance one month
			const next = new Date(y, m, 1);
			cursor = `${next.getFullYear()}-${String(next.getMonth() + 1).padStart(2, '0')}`;
		}
		return result;
	});

	// Weekday headers (Monday-start), localized
	let weekdayLabels = $derived.by(() => {
		// 2024-01-01 is a Monday
		return Array.from({ length: 7 }, (_, i) =>
			new Date(2024, 0, 1 + i).toLocaleDateString(currentLocale, { weekday: 'short' })
		);
	});

	function inWindow(day: string): boolean {
		return day >= today && day <= windowEnd;
	}

	// --- popover state ---
	let popover = $state<{ date: string; x: number; y: number; pinned: boolean } | null>(null);

	function openPopover(target: HTMLElement, date: string, pinned: boolean) {
		if (!itemsByDay[date]) return;
		const rect = target.getBoundingClientRect();
		const x = Math.min(Math.max(rect.left + rect.width / 2, 150), window.innerWidth - 150);
		popover = { date, x, y: rect.bottom + 8, pinned };
	}
	function handleEnter(e: MouseEvent, date: string) {
		if (popover?.pinned) return;
		openPopover(e.currentTarget as HTMLElement, date, false);
	}
	function handleLeave() {
		if (!popover?.pinned) popover = null;
	}
	function handleClick(e: MouseEvent, date: string) {
		if (!itemsByDay[date]) return;
		if (popover?.pinned && popover.date === date) {
			popover = null;
			return;
		}
		openPopover(e.currentTarget as HTMLElement, date, true);
	}

	let popoverData = $derived(popover ? itemsByDay[popover.date] : null);

	function formatFullDate(day: string): string {
		return parseLocal(day).toLocaleDateString(currentLocale, {
			weekday: 'long',
			day: 'numeric',
			month: 'long'
		});
	}
</script>

<div class="space-y-6">
	{#each months as month (month.key)}
		<div>
			<h3 class="mb-2 text-center font-semibold capitalize">{month.label}</h3>
			<div class="grid grid-cols-7 gap-1">
				{#each weekdayLabels as label}
					<div
						class="pb-1 text-center text-[0.65rem] font-semibold uppercase tracking-wide text-base-content/50"
					>
						{label}
					</div>
				{/each}

				{#each month.weeks as week}
					{#each week as day}
						{#if !day}
							<div class="min-h-[3.5rem] rounded-lg"></div>
						{:else if itemsByDay[day]}
							{@const data = itemsByDay[day]}
							<button
								type="button"
								class="flex min-h-[3.5rem] cursor-pointer flex-col gap-0.5 rounded-lg border border-base-300 bg-base-100 p-1 text-left transition-colors hover:border-primary hover:shadow-sm {day ===
								today
									? 'ring-2 ring-primary'
									: ''}"
								onmouseenter={(e) => handleEnter(e, day)}
								onmouseleave={handleLeave}
								onclick={(e) => handleClick(e, day)}
							>
								<div class="flex items-center justify-between">
									<span
										class="text-[0.7rem] font-medium {day === today
											? 'text-primary'
											: 'text-base-content/70'}"
									>
										{Number(day.split('-')[2])}
									</span>
									{#if data.items.length > 1}
										<span class="badge badge-ghost badge-xs">{data.items.length}</span>
									{/if}
								</div>

								<div class="flex flex-col gap-0.5">
									{#each data.items.slice(0, 2) as item}
										{@const cat = getCategoryById(item.category_id)}
										<span
											class="flex items-center gap-0.5 truncate rounded px-1 py-[1px] text-[0.6rem] leading-tight {cat
												? getContrastTextClass(cat.color)
												: 'text-gray-900'}"
											style="background-color: {cat?.color || '#d1d5db'};"
										>
											{#if item.recurring_transfer_group_id}
												<ArrowRightLeft size={9} class="shrink-0" />
											{/if}
											<span class="truncate">{formatCurrency(item.amount)}</span>
										</span>
									{/each}
									{#if data.items.length > 2}
										<span class="px-1 text-[0.6rem] leading-tight text-base-content/60">
											+{data.items.length - 2}
										</span>
									{/if}
								</div>
							</button>
						{:else}
							<div
								class="flex min-h-[3.5rem] flex-col rounded-lg border p-1 {inWindow(day)
									? 'border-base-300 bg-base-100'
									: 'border-transparent bg-base-200/30 opacity-50'} {day === today
									? 'ring-2 ring-primary'
									: ''}"
							>
								<span
									class="text-[0.7rem] font-medium {day === today
										? 'text-primary'
										: 'text-base-content/70'}"
								>
									{Number(day.split('-')[2])}
								</span>
							</div>
						{/if}
					{/each}
				{/each}
			</div>
		</div>
	{/each}
</div>

<!-- Hover / tap details popover -->
{#if popover?.pinned}
	<button
		class="fixed inset-0 z-40 cursor-default"
		aria-label={$t('common.close')}
		onclick={() => (popover = null)}
	></button>
{/if}

{#if popover && popoverData}
	<div
		class="fixed z-50 w-72 -translate-x-1/2 rounded-xl border border-base-300 bg-base-100 p-3 shadow-xl"
		style="left: {popover.x}px; top: {popover.y}px;"
	>
		<div class="mb-2 flex items-center justify-between border-b border-base-200 pb-2">
			<span class="text-sm font-semibold capitalize">{formatFullDate(popover.date)}</span>
			<span
				class="text-sm font-bold {popoverData.net >= 0 ? 'text-success' : 'text-error'}"
			>
				{popoverData.net >= 0 ? '+' : ''}{formatCurrency(popoverData.net)}
			</span>
		</div>
		<div class="flex max-h-64 flex-col gap-1.5 overflow-y-auto">
			{#each popoverData.items as item (`${item.recurring_rule_id}-${item.date}`)}
				{@const cat = getCategoryById(item.category_id)}
				{@const parent = getParentCategoryName(cat)}
				<div class="flex items-center gap-2">
					<span
						class="h-3 w-3 shrink-0 rounded-full"
						style="background-color: {cat?.color || '#d1d5db'};"
					></span>
					<div class="min-w-0 flex-1">
						<div class="flex items-center gap-1 text-xs font-medium">
							{#if item.recurring_transfer_group_id}
								<ArrowRightLeft size={11} class="shrink-0 text-info" />
							{/if}
							<span class="truncate">
								{#if parent}<span class="text-base-content/50">{parent} · </span>{/if}
								{cat?.category_name || 'N/A'}
							</span>
						</div>
						{#if item.description}
							<div class="truncate text-[0.7rem] text-base-content/60">{item.description}</div>
						{/if}
					</div>
					<span
						class="shrink-0 text-xs font-semibold {item.transaction_type_id ===
						TransactionTypeId.Debit
							? 'text-error'
							: 'text-success'}"
					>
						{item.transaction_type_id === TransactionTypeId.Debit ? '-' : '+'}{formatCurrency(
							item.amount
						)}
					</span>
				</div>
			{/each}
		</div>
	</div>
{/if}
