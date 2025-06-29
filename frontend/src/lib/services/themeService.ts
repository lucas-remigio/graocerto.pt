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
			legendColor: isDark ? '#e5e7eb' : '#374151',
			axisTextColor: isDark ? '#e5e7eb' : '#111827',
			gridColor: isDark ? '#374151' : '#f3f4f6',
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
