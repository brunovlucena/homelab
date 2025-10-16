// 📊 React Hook for Metrics Tracking
import { useEffect, useCallback, useRef } from 'react';
import { useLocation } from 'react-router-dom';
import { metricsCollector } from '../utils/metrics';

/**
 * Hook to track page views automatically
 */
export function usePageViewTracking(): void {
  const location = useLocation();

  useEffect(() => {
    // Track page view on mount and location change
    metricsCollector.recordPageView(location.pathname);
  }, [location.pathname]);
}

/**
 * Hook to track component render time
 */
export function useComponentRenderTime(componentName: string): void {
  const renderStartTime = useRef<number>(Date.now());

  useEffect(() => {
    const renderTime = Date.now() - renderStartTime.current;
    metricsCollector.recordComponentRender(componentName, renderTime);
  }, [componentName]);
}

/**
 * Hook to track user interactions
 */
export function useInteractionTracking() {
  const trackInteraction = useCallback(
    (action: string, target: string, labels?: Record<string, string>) => {
      metricsCollector.recordInteraction(action, target, labels);
    },
    []
  );

  return { trackInteraction };
}

/**
 * Hook to track API calls with automatic timing
 */
export function useAPIMetrics() {
  const trackAPICall = useCallback(
    async <T,>(
      endpoint: string,
      method: string,
      apiCallFn: () => Promise<T>
    ): Promise<T> => {
      const startTime = Date.now();
      let status = 0;
      let success = false;

      try {
        const result = await apiCallFn();
        status = 200; // Assume success
        success = true;
        return result;
      } catch (error: any) {
        status = error?.response?.status || 500;
        success = false;
        
        // Record error
        metricsCollector.recordError(
          'api_call_failed',
          `${method} ${endpoint} failed: ${error?.message || 'Unknown error'}`,
          error?.stack
        );
        
        throw error;
      } finally {
        const duration = Date.now() - startTime;
        metricsCollector.recordAPICall(endpoint, method, status, duration, success);
      }
    },
    []
  );

  return { trackAPICall };
}

/**
 * Hook to track errors in a component
 */
export function useErrorTracking(componentName: string) {
  const trackError = useCallback(
    (error: Error, errorInfo?: any) => {
      metricsCollector.recordError(
        `${componentName}_error`,
        error.message,
        error.stack
      );
    },
    [componentName]
  );

  return { trackError };
}

/**
 * Hook to track custom metrics
 */
export function useCustomMetric() {
  const recordMetric = useCallback(
    (name: string, value: number, labels?: Record<string, string>) => {
      metricsCollector.recordMetric(name, value, labels);
    },
    []
  );

  const incrementCounter = useCallback(
    (name: string, labels?: Record<string, string>) => {
      metricsCollector.incrementCounter(name, labels);
    },
    []
  );

  const recordHistogram = useCallback(
    (name: string, durationMs: number, labels?: Record<string, string>) => {
      metricsCollector.recordHistogram(name, durationMs, labels);
    },
    []
  );

  return { recordMetric, incrementCounter, recordHistogram };
}

/**
 * Hook to track performance timing for a component lifecycle
 */
export function usePerformanceTracking(
  componentName: string,
  trackUnmount: boolean = false
) {
  const mountTime = useRef<number>(Date.now());

  useEffect(() => {
    // Track mount time
    const timeToMount = Date.now() - mountTime.current;
    metricsCollector.recordHistogram(
      'frontend_component_mount_duration_seconds',
      timeToMount,
      { component: componentName }
    );

    return () => {
      if (trackUnmount) {
        // Track component lifetime
        const lifetime = Date.now() - mountTime.current;
        metricsCollector.recordHistogram(
          'frontend_component_lifetime_seconds',
          lifetime,
          { component: componentName }
        );
      }
    };
  }, [componentName, trackUnmount]);
}

/**
 * Hook to track form submissions
 */
export function useFormMetrics(formName: string) {
  const trackFormSubmit = useCallback(
    (success: boolean, errorMessage?: string) => {
      metricsCollector.incrementCounter('frontend_form_submissions_total', {
        form: formName,
        success: success.toString(),
      });

      if (!success && errorMessage) {
        metricsCollector.recordError(`form_${formName}_error`, errorMessage);
      }
    },
    [formName]
  );

  const trackFormField = useCallback(
    (fieldName: string, action: 'focus' | 'blur' | 'change') => {
      metricsCollector.incrementCounter('frontend_form_field_interactions_total', {
        form: formName,
        field: fieldName,
        action,
      });
    },
    [formName]
  );

  return { trackFormSubmit, trackFormField };
}

/**
 * Hook to track button clicks
 */
export function useButtonMetrics() {
  const trackButtonClick = useCallback((buttonName: string, context?: string) => {
    metricsCollector.recordInteraction('click', buttonName, {
      context: context || 'unknown',
    });
  }, []);

  return { trackButtonClick };
}

/**
 * Hook to track session duration
 */
export function useSessionTracking() {
  const sessionStart = useRef<number>(Date.now());

  useEffect(() => {
    // Track session start
    metricsCollector.incrementCounter('frontend_sessions_started_total');

    // Track session end on unmount
    return () => {
      const sessionDuration = Date.now() - sessionStart.current;
      metricsCollector.recordHistogram(
        'frontend_session_duration_seconds',
        sessionDuration
      );
      metricsCollector.incrementCounter('frontend_sessions_ended_total');
    };
  }, []);
}

/**
 * Hook to track scroll depth
 */
export function useScrollDepthTracking(pageName: string) {
  const maxScrollDepth = useRef<number>(0);
  const reported25 = useRef<boolean>(false);
  const reported50 = useRef<boolean>(false);
  const reported75 = useRef<boolean>(false);
  const reported100 = useRef<boolean>(false);

  useEffect(() => {
    const handleScroll = () => {
      const windowHeight = window.innerHeight;
      const documentHeight = document.documentElement.scrollHeight;
      const scrollTop = window.pageYOffset || document.documentElement.scrollTop;
      
      const scrollPercent = (scrollTop / (documentHeight - windowHeight)) * 100;
      
      if (scrollPercent > maxScrollDepth.current) {
        maxScrollDepth.current = scrollPercent;
      }

      // Track milestone scroll depths
      if (scrollPercent >= 25 && !reported25.current) {
        reported25.current = true;
        metricsCollector.incrementCounter('frontend_scroll_depth_total', {
          page: pageName,
          depth: '25',
        });
      }
      if (scrollPercent >= 50 && !reported50.current) {
        reported50.current = true;
        metricsCollector.incrementCounter('frontend_scroll_depth_total', {
          page: pageName,
          depth: '50',
        });
      }
      if (scrollPercent >= 75 && !reported75.current) {
        reported75.current = true;
        metricsCollector.incrementCounter('frontend_scroll_depth_total', {
          page: pageName,
          depth: '75',
        });
      }
      if (scrollPercent >= 100 && !reported100.current) {
        reported100.current = true;
        metricsCollector.incrementCounter('frontend_scroll_depth_total', {
          page: pageName,
          depth: '100',
        });
      }
    };

    window.addEventListener('scroll', handleScroll, { passive: true });
    
    return () => {
      window.removeEventListener('scroll', handleScroll);
      
      // Report final max scroll depth
      metricsCollector.recordMetric('frontend_max_scroll_depth_percent', maxScrollDepth.current, {
        page: pageName,
      });
    };
  }, [pageName]);
}

