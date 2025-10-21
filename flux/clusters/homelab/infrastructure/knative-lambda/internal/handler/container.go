// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//
//	🎯 DEPENDENCY INJECTION CONTAINER - Component lifecycle management
//
//	🎯 Purpose: Manage component dependencies and lifecycle
//	💡 Features: Dependency injection, component composition, loose coupling
//
//	🏛️ ARCHITECTURE:
//	🔧 Component Management - Centralized dependency management
//	🔄 Lifecycle Control - Component initialization and shutdown
//	🔗 Interface Composition - Compose focused interfaces
//	📊 Dependency Graph - Clear dependency relationships
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
package handler

import (
	"context"
	"fmt"
	"sync"

	"knative-lambda-new/internal/config"
	"knative-lambda-new/internal/observability"
	"knative-lambda-new/internal/resilience"
)

// 🎯 ComponentContainerImpl - "Dependency injection container implementation"
type ComponentContainerImpl struct {
	// 🔧 Core Dependencies
	config *config.Config
	obs    *observability.Observability

	// 🎯 HTTP Components
	httpHandler       HTTPHandler
	cloudEventHandler CloudEventHandler

	// 🎯 Job Management Components
	jobManager      JobManager
	asyncJobCreator AsyncJobCreatorInterface

	// 🎯 Event Processing Components
	eventHandler EventHandler

	// 🎯 Service Management Components
	serviceManager ServiceManager

	// 🎯 Build Context Components
	buildContextManager BuildContextManager

	// 🚦 Rate Limiting
	rateLimiter *resilience.MultiLevelRateLimiter

	// 🔒 Thread Safety
	mu sync.RWMutex
}

// 🏗️ NewComponentContainer - "Create new component container"
func NewComponentContainer(config *config.Config, obs *observability.Observability) *ComponentContainerImpl {
	return &ComponentContainerImpl{
		config: config,
		obs:    obs,
	}
}

// 🔧 SetHTTPHandler - "Set HTTP handler component"
func (c *ComponentContainerImpl) SetHTTPHandler(handler HTTPHandler) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.httpHandler = handler
}

// 🔧 SetCloudEventHandler - "Set CloudEvent handler component"
func (c *ComponentContainerImpl) SetCloudEventHandler(handler CloudEventHandler) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cloudEventHandler = handler
}

// 🔧 SetJobManager - "Set job manager component"
func (c *ComponentContainerImpl) SetJobManager(manager JobManager) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.jobManager = manager
}

// 🔧 SetAsyncJobCreator - "Set async job creator component"
func (c *ComponentContainerImpl) SetAsyncJobCreator(creator AsyncJobCreatorInterface) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.asyncJobCreator = creator
}

// 🔧 SetEventHandler - "Set event handler component"
func (c *ComponentContainerImpl) SetEventHandler(handler EventHandler) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.eventHandler = handler
}

// 🔧 SetServiceManager - "Set service manager component"
func (c *ComponentContainerImpl) SetServiceManager(manager ServiceManager) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.serviceManager = manager
}

// 🔧 SetBuildContextManager - "Set build context manager component"
func (c *ComponentContainerImpl) SetBuildContextManager(manager BuildContextManager) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.buildContextManager = manager
}

// 🔧 SetRateLimiter - "Set rate limiter component"
func (c *ComponentContainerImpl) SetRateLimiter(limiter *resilience.MultiLevelRateLimiter) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.rateLimiter = limiter
}

// 📥 GetHTTPHandler - "Get HTTP handler component"
func (c *ComponentContainerImpl) GetHTTPHandler() HTTPHandler {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.httpHandler == nil {
		panic("HTTP handler not initialized")
	}
	return c.httpHandler
}

// 📥 GetCloudEventHandler - "Get CloudEvent handler component"
func (c *ComponentContainerImpl) GetCloudEventHandler() CloudEventHandler {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.cloudEventHandler == nil {
		panic("CloudEvent handler not initialized")
	}
	return c.cloudEventHandler
}

// 📥 GetJobManager - "Get job manager component"
func (c *ComponentContainerImpl) GetJobManager() JobManager {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.jobManager == nil {
		panic("Job manager not initialized")
	}
	return c.jobManager
}

// 📥 GetAsyncJobCreator - "Get async job creator component"
func (c *ComponentContainerImpl) GetAsyncJobCreator() AsyncJobCreatorInterface {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.asyncJobCreator == nil {
		panic("Async job creator not initialized")
	}
	return c.asyncJobCreator
}

// 📥 GetEventHandler - "Get event handler component"
func (c *ComponentContainerImpl) GetEventHandler() EventHandler {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.eventHandler
}

// 📥 GetServiceManager - "Get service manager component"
func (c *ComponentContainerImpl) GetServiceManager() ServiceManager {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.serviceManager == nil {
		panic("Service manager not initialized")
	}
	return c.serviceManager
}

// 📥 GetBuildContextManager - "Get build context manager component"
func (c *ComponentContainerImpl) GetBuildContextManager() BuildContextManager {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.buildContextManager == nil {
		panic("Build context manager not initialized")
	}
	return c.buildContextManager
}

// 📥 GetConfig - "Get configuration"
func (c *ComponentContainerImpl) GetConfig() *config.Config {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.config
}

// 📥 GetObservability - "Get observability instance"
func (c *ComponentContainerImpl) GetObservability() *observability.Observability {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.obs
}

// 📥 GetRateLimiter - "Get rate limiter component"
func (c *ComponentContainerImpl) GetRateLimiter() *resilience.MultiLevelRateLimiter {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.rateLimiter
}

// 🔄 Shutdown - "Gracefully shut down all components"
func (c *ComponentContainerImpl) Shutdown(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var errors []error

	// Shutdown rate limiter
	if c.rateLimiter != nil {
		if err := c.rateLimiter.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close rate limiter: %w", err))
		}
	}

	// Log shutdown completion
	if len(errors) > 0 {
		c.obs.Error(ctx, fmt.Errorf("component shutdown errors: %v", errors), "Component container shutdown completed with errors")
		return fmt.Errorf("component shutdown errors: %v", errors)
	}

	c.obs.Info(ctx, "Component container shutdown completed successfully")
	return nil
}

// 🔍 ValidateComponents - "Validate all components are initialized"
func (c *ComponentContainerImpl) ValidateComponents() error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var missing []string

	if c.httpHandler == nil {
		missing = append(missing, "HTTP handler")
	}
	if c.cloudEventHandler == nil {
		missing = append(missing, "CloudEvent handler")
	}
	if c.jobManager == nil {
		missing = append(missing, "Job manager")
	}

	if c.serviceManager == nil {
		missing = append(missing, "Service manager")
	}
	if c.buildContextManager == nil {
		missing = append(missing, "Build context manager")
	}
	if c.rateLimiter == nil {
		missing = append(missing, "Rate limiter")
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required components: %v", missing)
	}

	return nil
}
