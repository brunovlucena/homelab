// 📊 Web Vitals Tracking
// Tracks Core Web Vitals (CLS, FID, LCP, FCP, TTFB) and sends to metrics collector

import { metricsCollector, WebVitalsMetric } from './metrics';

// Web Vitals types (simplified version)
interface Metric {
  name: string;
  value: number;
  delta: number;
  id: string;
  navigationType: 'navigate' | 'reload' | 'back_forward' | 'back_forward_cache' | 'prerender';
  rating?: 'good' | 'needs-improvement' | 'poor';
}

/**
 * Report Web Vital to metrics collector
 */
function reportWebVital(metric: Metric): void {
  const webVitalsMetric: WebVitalsMetric = {
    name: metric.name,
    value: metric.value,
    delta: metric.delta,
    id: metric.id,
    navigationType: metric.navigationType,
  };

  metricsCollector.recordWebVital(webVitalsMetric);

  // Log in development
  if (import.meta.env.DEV) {
    console.log('Web Vital:', metric.name, metric.value, metric.rating);
  }
}

/**
 * Initialize Web Vitals tracking
 * This is a simplified implementation that tracks basic performance metrics
 * For full Core Web Vitals, consider using the 'web-vitals' npm package
 */
export function initWebVitals(): void {
  if (typeof window === 'undefined') {
    return;
  }

  // Track Largest Contentful Paint (LCP)
  try {
    const observer = new PerformanceObserver((list) => {
      const entries = list.getEntries();
      const lastEntry = entries[entries.length - 1] as any;
      
      reportWebVital({
        name: 'LCP',
        value: lastEntry.renderTime || lastEntry.loadTime,
        delta: lastEntry.renderTime || lastEntry.loadTime,
        id: 'lcp',
        navigationType: 'navigate',
        rating: getRating('LCP', lastEntry.renderTime || lastEntry.loadTime),
      });
    });

    observer.observe({ type: 'largest-contentful-paint', buffered: true });
  } catch (e) {
    console.error('Failed to observe LCP:', e);
  }

  // Track First Input Delay (FID)
  try {
    const observer = new PerformanceObserver((list) => {
      const entries = list.getEntries();
      entries.forEach((entry: any) => {
        const fid = entry.processingStart - entry.startTime;
        
        reportWebVital({
          name: 'FID',
          value: fid,
          delta: fid,
          id: 'fid',
          navigationType: 'navigate',
          rating: getRating('FID', fid),
        });
      });
    });

    observer.observe({ type: 'first-input', buffered: true });
  } catch (e) {
    console.error('Failed to observe FID:', e);
  }

  // Track Cumulative Layout Shift (CLS)
  try {
    let clsValue = 0;
    const observer = new PerformanceObserver((list) => {
      const entries = list.getEntries();
      entries.forEach((entry: any) => {
        if (!entry.hadRecentInput) {
          clsValue += entry.value;
        }
      });

      reportWebVital({
        name: 'CLS',
        value: clsValue,
        delta: clsValue,
        id: 'cls',
        navigationType: 'navigate',
        rating: getRating('CLS', clsValue),
      });
    });

    observer.observe({ type: 'layout-shift', buffered: true });
  } catch (e) {
    console.error('Failed to observe CLS:', e);
  }

  // Track First Contentful Paint (FCP)
  try {
    const observer = new PerformanceObserver((list) => {
      const entries = list.getEntries();
      entries.forEach((entry: any) => {
        if (entry.name === 'first-contentful-paint') {
          reportWebVital({
            name: 'FCP',
            value: entry.startTime,
            delta: entry.startTime,
            id: 'fcp',
            navigationType: 'navigate',
            rating: getRating('FCP', entry.startTime),
          });
        }
      });
    });

    observer.observe({ type: 'paint', buffered: true });
  } catch (e) {
    console.error('Failed to observe FCP:', e);
  }

  // Track Time to First Byte (TTFB)
  try {
    const navTiming = performance.getEntriesByType('navigation')[0] as PerformanceNavigationTiming;
    if (navTiming) {
      const ttfb = navTiming.responseStart - navTiming.requestStart;
      
      reportWebVital({
        name: 'TTFB',
        value: ttfb,
        delta: ttfb,
        id: 'ttfb',
        navigationType: 'navigate',
        rating: getRating('TTFB', ttfb),
      });
    }
  } catch (e) {
    console.error('Failed to observe TTFB:', e);
  }

  // Track Interaction to Next Paint (INP) - experimental
  try {
    const observer = new PerformanceObserver((list) => {
      const entries = list.getEntries();
      entries.forEach((entry: any) => {
        const inp = entry.processingStart - entry.startTime + entry.duration;
        
        reportWebVital({
          name: 'INP',
          value: inp,
          delta: inp,
          id: 'inp',
          navigationType: 'navigate',
          rating: getRating('INP', inp),
        });
      });
    });

    // Use type assertion as durationThreshold is not in official TypeScript types yet
    observer.observe({ type: 'event', buffered: true, durationThreshold: 40 } as PerformanceObserverInit);
  } catch (e) {
    // INP is experimental and may not be supported
    if (import.meta.env.DEV) {
      console.log('INP tracking not supported');
    }
  }
}

/**
 * Get rating for a metric based on its value
 */
function getRating(metricName: string, value: number): 'good' | 'needs-improvement' | 'poor' {
  // Thresholds based on Web Vitals recommendations
  const thresholds: Record<string, { good: number; poor: number }> = {
    LCP: { good: 2500, poor: 4000 }, // milliseconds
    FID: { good: 100, poor: 300 }, // milliseconds
    CLS: { good: 0.1, poor: 0.25 }, // score
    FCP: { good: 1800, poor: 3000 }, // milliseconds
    TTFB: { good: 800, poor: 1800 }, // milliseconds
    INP: { good: 200, poor: 500 }, // milliseconds
  };

  const threshold = thresholds[metricName];
  if (!threshold) {
    return 'good';
  }

  if (value <= threshold.good) {
    return 'good';
  } else if (value <= threshold.poor) {
    return 'needs-improvement';
  } else {
    return 'poor';
  }
}

/**
 * Track custom performance mark
 */
export function trackPerformanceMark(name: string): void {
  if (typeof window === 'undefined' || !window.performance) {
    return;
  }

  try {
    performance.mark(name);
  } catch (e) {
    console.error('Failed to create performance mark:', e);
  }
}

/**
 * Track custom performance measure
 */
export function trackPerformanceMeasure(
  name: string,
  startMark: string,
  endMark?: string
): number | null {
  if (typeof window === 'undefined' || !window.performance) {
    return null;
  }

  try {
    const measure = performance.measure(name, startMark, endMark);
    metricsCollector.recordHistogram(`frontend_custom_measure_${name}`, measure.duration, {
      start_mark: startMark,
      end_mark: endMark || 'now',
    });
    return measure.duration;
  } catch (e) {
    console.error('Failed to create performance measure:', e);
    return null;
  }
}

/**
 * Clear performance marks and measures
 */
export function clearPerformanceMarks(name?: string): void {
  if (typeof window === 'undefined' || !window.performance) {
    return;
  }

  try {
    if (name) {
      performance.clearMarks(name);
      performance.clearMeasures(name);
    } else {
      performance.clearMarks();
      performance.clearMeasures();
    }
  } catch (e) {
    console.error('Failed to clear performance marks:', e);
  }
}

