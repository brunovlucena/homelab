// =============================================================================
// 📋 CORE DATA TYPES
// =============================================================================

/**
 * Project represents a portfolio project
 */
export interface Project {
  id: number;
  title: string;
  description: string;
  short_description: string;
  type: string;
  icon: string;
  github_url: string;
  live_url: string;
  technologies: string[];
  active: boolean;
  github_active: boolean;
}

/**
 * Skill represents a technical skill or technology
 */
export interface Skill {
  id: number;
  name: string;
  category: string;
  icon: string;
  order: number;
}

/**
 * Experience represents a work experience entry
 */
export interface Experience {
  id: number;
  title: string;
  company: string;
  start_date: string;
  end_date?: string;
  current: boolean;
  company_description?: string;
  description: string;
  technologies: string[];
  order: number;
  active?: boolean;
}

/**
 * Content represents dynamic content stored in the database
 * Note: Used for API responses with id/type/value format
 */
export interface Content {
  id?: number;
  key?: string;
  type?: string;
  value: string | ContentValue;
}

/**
 * ContentValue represents the structured value for content items
 */
export interface ContentValue {
  title?: string;
  description?: string;
  highlights?: Array<{ icon: string; text: string }>;
  email?: string;
  location?: string;
  linkedin?: string;
  github?: string;
  availability?: string;
  subtitle?: string;
}

// =============================================================================
// 📊 ANALYTICS TYPES
// =============================================================================

/**
 * Visitor represents a site visitor for analytics
 */
export interface Visitor {
  id: number;
  ip: string;
  user_agent: string;
  country?: string;
  city?: string;
  first_visit: string;
  last_visit: string;
  visit_count: number;
}

/**
 * AnalyticsData represents aggregated analytics data
 */
export interface AnalyticsData {
  total_visitors: number;
  unique_visitors: number;
  total_views: number;
  project_views: Record<number, number>;
}

// =============================================================================
// 🌐 API TYPES
// =============================================================================

/**
 * ApiResponse wraps API responses with metadata
 */
export interface ApiResponse<T> {
  data: T;
  message?: string;
  error?: string;
  request_id?: string;
}

/**
 * ApiError represents an API error response
 */
export interface ApiError {
  error: string;
  details?: string;
  request_id?: string;
} 