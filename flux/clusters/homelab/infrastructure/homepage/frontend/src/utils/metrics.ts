// 📊 Frontend Metrics Utility
// Tracks client-side metrics and sends them to the backend API

interface MetricData {
  name: string;
  value: number;
  labels?: Record<string, string>;
  timestamp?: number;
}

interface WebVitalsMetric {
  name: string;
  value: number;
  delta: number;
  id: string;
  navigationType: string;
}

class MetricsCollector {
  private readonly apiUrl: string;
  private readonly batchSize: number = 10;
  private readonly flushInterval: number = 30000; // 30 seconds
  private metricsBuffer: MetricData[] = [];
  private flushTimer: NodeJS.Timeout | null = null;

  constructor() {
    // Get API URL from environment or use default
    this.apiUrl = import.meta.env.VITE_API_URL || '/api/v1';
    
    // Start periodic flush
    this.startPeriodicFlush();
    
    // Flush on page unload
    if (typeof window !== 'undefined') {
      window.addEventListener('beforeunload', () => this.flush());
    }
  }

  /**
   * Record a metric
   */
  public recordMetric(name: string, value: number, labels?: Record<string, string>): void {
    const metric: MetricData = {
      name,
      value,
      labels: {
        ...labels,
        environment: import.meta.env.VITE_APP_ENV || 'production',
        user_agent: this.getUserAgentInfo(),
      },
      timestamp: Date.now(),
    };

    this.metricsBuffer.push(metric);

    // Flush if buffer is full
    if (this.metricsBuffer.length >= this.batchSize) {
      this.flush();
    }
  }

  /**
   * Record a counter metric (increments by 1)
   */
  public incrementCounter(name: string, labels?: Record<string, string>): void {
    this.recordMetric(name, 1, labels);
  }

  /**
   * Record a histogram metric (duration in milliseconds)
   */
  public recordHistogram(name: string, durationMs: number, labels?: Record<string, string>): void {
    this.recordMetric(name, durationMs / 1000, labels); // Convert to seconds
  }

  /**
   * Record page view
   */
  public recordPageView(path: string): void {
    this.incrementCounter('homepage_frontend_page_views_total', {
      path,
      referrer: document.referrer || 'direct',
    });
  }

  /**
   * Record navigation timing
   */
  public recordNavigationTiming(): void {
    if (typeof window === 'undefined' || !window.performance) {
      return;
    }

    const perfData = window.performance.getEntriesByType('navigation')[0] as PerformanceNavigationTiming;
    if (!perfData) {
      return;
    }

    // DNS lookup time
    this.recordHistogram(
      'homepage_frontend_dns_lookup_duration_seconds',
      perfData.domainLookupEnd - perfData.domainLookupStart,
      { type: 'dns' }
    );

    // TCP connection time
    this.recordHistogram(
      'homepage_frontend_tcp_connection_duration_seconds',
      perfData.connectEnd - perfData.connectStart,
      { type: 'tcp' }
    );

    // Time to first byte
    this.recordHistogram(
      'homepage_frontend_ttfb_duration_seconds',
      perfData.responseStart - perfData.requestStart,
      { type: 'ttfb' }
    );

    // DOM content loaded
    this.recordHistogram(
      'homepage_frontend_dom_content_loaded_duration_seconds',
      perfData.domContentLoadedEventEnd - perfData.domContentLoadedEventStart,
      { type: 'dom_content_loaded' }
    );

    // Page load time
    this.recordHistogram(
      'homepage_frontend_page_load_duration_seconds',
      perfData.loadEventEnd - perfData.loadEventStart,
      { type: 'page_load' }
    );

    // Total time from navigation start to load complete
    this.recordHistogram(
      'homepage_frontend_total_load_duration_seconds',
      perfData.loadEventEnd - perfData.fetchStart,
      { type: 'total_load' }
    );
  }

  /**
   * Record API call metrics
   */
  public recordAPICall(
    endpoint: string,
    method: string,
    status: number,
    durationMs: number,
    success: boolean
  ): void {
    // Record API call count
    this.incrementCounter('homepage_frontend_api_requests_total', {
      endpoint,
      method,
      status: status.toString(),
      success: success.toString(),
    });

    // Record API call duration
    this.recordHistogram('homepage_frontend_api_request_duration_seconds', durationMs, {
      endpoint,
      method,
      status: status.toString(),
    });

    // Record errors separately
    if (!success) {
      this.incrementCounter('homepage_frontend_api_errors_total', {
        endpoint,
        method,
        status: status.toString(),
      });
    }
  }

  /**
   * Record user interaction
   */
  public recordInteraction(action: string, target: string, labels?: Record<string, string>): void {
    this.incrementCounter('homepage_frontend_user_interactions_total', {
      action,
      target,
      ...labels,
    });
  }

  /**
   * Record error
   */
  public recordError(errorType: string, message: string, stack?: string): void {
    this.incrementCounter('homepage_frontend_errors_total', {
      error_type: errorType,
      message: message.substring(0, 100), // Limit message length
    });

    // Log to console in development
    if (import.meta.env.DEV) {
      console.error('Metrics: Error recorded', { errorType, message, stack });
    }
  }

  /**
   * Record Web Vitals metric
   */
  public recordWebVital(metric: WebVitalsMetric): void {
    const labels = {
      metric_name: metric.name,
      navigation_type: metric.navigationType,
    };

    // Record the actual value
    this.recordMetric(`homepage_frontend_web_vitals_${metric.name.toLowerCase()}`, metric.value, labels);

    // Record delta (change since last measurement)
    this.recordMetric(`homepage_frontend_web_vitals_${metric.name.toLowerCase()}_delta`, metric.delta, labels);
  }

  /**
   * Record resource timing
   */
  public recordResourceTiming(): void {
    if (typeof window === 'undefined' || !window.performance) {
      return;
    }

    const resources = window.performance.getEntriesByType('resource') as PerformanceResourceTiming[];
    
    // Group by resource type
    const resourcesByType: Record<string, number[]> = {};
    const resourcesSizeByType: Record<string, number[]> = {};

    resources.forEach((resource) => {
      const type = this.getResourceType(resource.name);
      const duration = resource.responseEnd - resource.startTime;
      const size = resource.transferSize || 0;

      if (!resourcesByType[type]) {
        resourcesByType[type] = [];
        resourcesSizeByType[type] = [];
      }

      resourcesByType[type].push(duration);
      resourcesSizeByType[type].push(size);
    });

    // Record metrics for each resource type
    Object.entries(resourcesByType).forEach(([type, durations]) => {
      const avgDuration = durations.reduce((a, b) => a + b, 0) / durations.length;
      const sizes = resourcesSizeByType[type];
      const totalSize = sizes.reduce((a, b) => a + b, 0);

      this.recordHistogram('homepage_frontend_resource_load_duration_seconds', avgDuration, {
        resource_type: type,
      });

      this.recordMetric('homepage_frontend_resource_size_bytes', totalSize, {
        resource_type: type,
      });

      this.incrementCounter('homepage_frontend_resources_loaded_total', {
        resource_type: type,
        count: durations.length.toString(),
      });
    });
  }

  /**
   * Record component render time
   */
  public recordComponentRender(componentName: string, durationMs: number): void {
    this.recordHistogram('homepage_frontend_component_render_duration_seconds', durationMs, {
      component: componentName,
    });
  }

  /**
   * Flush metrics buffer to backend
   */
  private async flush(): Promise<void> {
    if (this.metricsBuffer.length === 0) {
      return;
    }

    const metricsToSend = [...this.metricsBuffer];
    this.metricsBuffer = [];

    try {
      // Send metrics to backend
      const response = await fetch(`${this.apiUrl}/metrics/frontend`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ metrics: metricsToSend }),
        // Use keepalive for beforeunload events
        keepalive: true,
      });

      if (!response.ok) {
        console.error('Failed to send metrics:', response.statusText);
        // Put metrics back in buffer for retry
        this.metricsBuffer.unshift(...metricsToSend);
      }
    } catch (error) {
      console.error('Error sending metrics:', error);
      // Put metrics back in buffer for retry
      this.metricsBuffer.unshift(...metricsToSend);
    }
  }

  /**
   * Start periodic flush
   */
  private startPeriodicFlush(): void {
    if (this.flushTimer) {
      clearInterval(this.flushTimer);
    }

    this.flushTimer = setInterval(() => {
      this.flush();
    }, this.flushInterval);
  }

  /**
   * Get resource type from URL
   */
  private getResourceType(url: string): string {
    const ext = url.split('.').pop()?.toLowerCase() || '';
    
    if (['jpg', 'jpeg', 'png', 'gif', 'svg', 'webp', 'avif'].includes(ext)) {
      return 'image';
    } else if (['css'].includes(ext)) {
      return 'stylesheet';
    } else if (['js', 'mjs'].includes(ext)) {
      return 'script';
    } else if (['woff', 'woff2', 'ttf', 'eot'].includes(ext)) {
      return 'font';
    } else if (url.includes('/api/')) {
      return 'api';
    }
    
    return 'other';
  }

  /**
   * Get simplified user agent info
   */
  private getUserAgentInfo(): string {
    if (typeof window === 'undefined') {
      return 'unknown';
    }

    const ua = window.navigator.userAgent;
    
    if (ua.includes('Chrome')) {
      return 'chrome';
    } else if (ua.includes('Firefox')) {
      return 'firefox';
    } else if (ua.includes('Safari')) {
      return 'safari';
    } else if (ua.includes('Edge')) {
      return 'edge';
    }
    
    return 'other';
  }

  /**
   * Stop metrics collection
   */
  public stop(): void {
    if (this.flushTimer) {
      clearInterval(this.flushTimer);
      this.flushTimer = null;
    }
    this.flush();
  }
}

// Export singleton instance
export const metricsCollector = new MetricsCollector();

// Export types
export type { MetricData, WebVitalsMetric };

