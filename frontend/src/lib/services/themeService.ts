// Theme service singleton for managing theme changes across components
class ThemeService {
	private static instance: ThemeService;
	private listeners: Set<() => void> = new Set();
	private observer: MutationObserver | null = null;
	private mediaQuery: MediaQueryList | null = null;
	private initialized = false;

	private constructor() {
		this.initialize();
	}

	public static getInstance(): ThemeService {
		if (!ThemeService.instance) {
			ThemeService.instance = new ThemeService();
		}
		return ThemeService.instance;
	}

	public updateThemeColor(theme: 'light' | 'dark') {
		const themeColorMetaTag = document.querySelector('meta[name="theme-color"]');
		if (themeColorMetaTag) {
			const themeColor = theme === 'dark' ? '#4F99FF' : '#006FF9'; // Adjust colors as needed
			themeColorMetaTag.setAttribute('content', themeColor);
		}
	}

	private initialize() {
		if (this.initialized || typeof window === 'undefined') return;

		// Listen for data-theme attribute changes
		this.observer = new MutationObserver((mutations) => {
			mutations.forEach((mutation) => {
				if (mutation.type === 'attributes' && mutation.attributeName === 'data-theme') {
					this.notifyListeners();
				}
			});
		});

		// Observe the document element for theme changes
		if (document.documentElement) {
			this.observer.observe(document.documentElement, {
				attributes: true,
				attributeFilter: ['data-theme']
			});
		}

		// Listen for system theme changes
		this.mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
		const handleMediaChange = () => {
			this.notifyListeners();
		};
		this.mediaQuery.addEventListener('change', handleMediaChange);

		this.initialized = true;
	}

	private notifyListeners() {
		this.listeners.forEach((callback) => callback());
	}

	public subscribe(callback: () => void): () => void {
		this.listeners.add(callback);

		// Return unsubscribe function
		return () => {
			this.listeners.delete(callback);
		};
	}

	public isDarkMode(): boolean {
		if (typeof window === 'undefined') return false;

		const dataTheme = document.documentElement.getAttribute('data-theme');
		const systemPrefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;

		// Prioritize explicit theme setting over system preference
		if (dataTheme) {
			return dataTheme.includes('dark');
		} else {
			return systemPrefersDark;
		}
	}

	public getThemeColors() {
		const isDark = this.isDarkMode();

		return {
			isDarkMode: isDark,
			// The "total" line + its gradient area — DaisyUI `primary`, so it matches the
			// Total legend pill (which uses the primary token) in both themes.
			seriesTotal: isDark ? '#4F99FF' : '#006FF9',
			legendColor: isDark ? '#e5e7eb' : '#374151',
			axisTextColor: isDark ? '#9ca3af' : '#6b7280',
			// Recessive: hairline gridlines close to the surface, not a strong grey.
			gridColor: isDark ? '#2a2a2a' : '#eef1f4',
			tooltipBg: isDark ? '#1f2937' : '#ffffff',
			tooltipTitleColor: isDark ? '#e5e7eb' : '#111827',
			tooltipBodyColor: isDark ? '#e5e7eb' : '#374151',
			tooltipBorderColor: isDark ? '#374151' : '#d1d5db'
		};
	}

	public destroy() {
		if (this.observer) {
			this.observer.disconnect();
			this.observer = null;
		}
		if (this.mediaQuery) {
			this.mediaQuery.removeEventListener('change', () => {});
		}
		this.listeners.clear();
		this.initialized = false;
	}
}

export const themeService = ThemeService.getInstance();
