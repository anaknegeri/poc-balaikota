export const config = {
  API_BASE_URL:
    import.meta.env.VITE_API_BASE_URL || 'http://localhost:3002/api',
  REFRESH_INTERVAL: {
    DASHBOARD: import.meta.env.VITE_DASHBOARD_REFRESH_INTERVAL
      ? parseInt(import.meta.env.VITE_DASHBOARD_REFRESH_INTERVAL)
      : 30000, // 30 seconds
    ALERTS: import.meta.env.VITE_ALERTS_REFRESH_INTERVAL
      ? parseInt(import.meta.env.VITE_ALERTS_REFRESH_INTERVAL)
      : 15000, // 15 seconds
    CHARTS: import.meta.env.VITE_CHARTS_REFRESH_INTERVAL
      ? parseInt(import.meta.env.VITE_CHARTS_REFRESH_INTERVAL)
      : 60000, // 1 minute
  },
  CHART_LIMITS: {
    TRENDS: import.meta.env.VITE_CHART_TRENDS_LIMIT
      ? parseInt(import.meta.env.VITE_CHART_TRENDS_LIMIT)
      : 24,
    ALERTS: import.meta.env.VITE_CHART_ALERTS_LIMIT
      ? parseInt(import.meta.env.VITE_CHART_ALERTS_LIMIT)
      : 10,
  },
  UI: {
    THEME: import.meta.env.VITE_DEFAULT_THEME || 'light',
    ANIMATION_DURATION: import.meta.env.VITE_ANIMATION_DURATION
      ? parseInt(import.meta.env.VITE_ANIMATION_DURATION)
      : 300,
  },
  FEATURES: {
    AUTO_REFRESH:
      import.meta.env.VITE_ENABLE_AUTO_REFRESH !== undefined
        ? import.meta.env.VITE_ENABLE_AUTO_REFRESH !== 'false'
        : true,
    IMAGE_PREVIEW:
      import.meta.env.VITE_ENABLE_IMAGE_PREVIEW !== undefined
        ? import.meta.env.VITE_ENABLE_IMAGE_PREVIEW !== 'false'
        : true,
  },
} as const

// Environment validation with warnings instead of errors
export const validateConfig = () => {
  const warnings: string[] = []

  if (config.API_BASE_URL === 'http://localhost:3002/api') {
    warnings.push(
      'Using default API URL. Set VITE_API_BASE_URL in .env if different.'
    )
  }

  if (config.REFRESH_INTERVAL.DASHBOARD < 5000) {
    warnings.push(
      'Dashboard refresh interval is very low (< 5s). This may cause performance issues.'
    )
  }

  if (config.REFRESH_INTERVAL.ALERTS < 5000) {
    warnings.push(
      'Alerts refresh interval is very low (< 5s). This may cause performance issues.'
    )
  }

  return warnings.length === 0
}
