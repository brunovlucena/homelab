// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//
//	📄 TEMPLATES - Template processing for Knative Lambda service
//
//	🎯 Purpose: Process Go templates for Dockerfile and index.js generation
//	💡 Features: Template loading, variable substitution, error handling
//
//	🏛️ ARCHITECTURE:
//	📋 Template Data - Structured data for template variables
//	🔧 Template Processing - Go template execution with error handling
//	📁 File Embedding - Embed template files in Go binary
//	✅ Validation - Template validation and error reporting
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
package templates

import (
	"bytes"
	"context"
	"fmt"
	"text/template"
	"time"

	"knative-lambda-new/internal/handler/helpers"
	"knative-lambda-new/internal/observability"
	"knative-lambda-new/pkg/builds"
)

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  📋 TEMPLATE DATA - "Data structures for template processing"          │
// └─────────────────────────────────────────────────────────────────────────┘

// TemplateData represents the data structure passed to templates
// Contains only the fields that are actually used in templates
type TemplateData struct {
	FunctionName  string    `json:"function_name"`
	ThirdPartyId  string    `json:"third_party_id"`
	ParserId      string    `json:"parser_id"`
	NodeBaseImage string    `json:"node_base_image"`
	Timestamp     time.Time `json:"timestamp"`
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🔧 TEMPLATE PROCESSOR - "Template processing functionality"           │
// └─────────────────────────────────────────────────────────────────────────┘

// TemplateProcessor handles template processing operations
type TemplateProcessor struct {
	obs *observability.Observability
}

// NewTemplateProcessor creates a new template processor
func NewTemplateProcessor(obs *observability.Observability) *TemplateProcessor {
	return &TemplateProcessor{
		obs: obs,
	}
}

// ProcessTemplate processes a template with the given data
func (tp *TemplateProcessor) ProcessTemplate(ctx context.Context, templateName, templateContent string, data TemplateData) ([]byte, error) {
	tp.obs.Info(ctx, "Processing template",
		"template_name", templateName,
		"function_name", data.FunctionName,
		"parser_id", data.ParserId)

	// Parse the template
	tmpl, err := template.New(templateName).Parse(templateContent)
	if err != nil {
		tp.obs.Error(ctx, err, "Failed to parse template",
			"template_name", templateName)
		return nil, fmt.Errorf("failed to parse template %s: %w", templateName, err)
	}

	// Execute the template
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		tp.obs.Error(ctx, err, "Failed to execute template",
			"template_name", templateName)
		return nil, fmt.Errorf("failed to execute template %s: %w", templateName, err)
	}

	tp.obs.Info(ctx, "Successfully processed template",
		"template_name", templateName,
		"output_size", buf.Len())

	return buf.Bytes(), nil
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  📁 TEMPLATE FILES - "Embedded template file content"                  │
// └─────────────────────────────────────────────────────────────────────────┘

// DockerfileTemplate contains the Dockerfile template content
const DockerfileTemplate = `# Multi-stage Dockerfile for {{.FunctionName}}
# Stage 1: Build stage
ARG NODE_BASE_IMAGE
FROM {{.NodeBaseImage}} as builder

# Set working directory
WORKDIR /app

# Copy package files first (for better caching)
COPY package.json ./

# Create .npmrc with proper registry configuration
RUN echo "registry=https://registry.npmjs.org/" > .npmrc && \
    echo "fetch-retries=5" >> .npmrc && \
    echo "fetch-retry-mintimeout=10000" >> .npmrc && \
    echo "fetch-retry-maxtimeout=60000" >> .npmrc && \
    echo "fetch-retry-factor=2" >> .npmrc && \
    echo "network-timeout=60000" >> .npmrc && \
    echo "maxsockets=50" >> .npmrc && \
    echo "audit=false" >> .npmrc && \
    echo "fund=false" >> .npmrc && \
    echo "update-notifier=false" >> .npmrc && \
    echo "strict-ssl=true" >> .npmrc && \
    echo "user-agent=npm/kaniko-builder" >> .npmrc && \
    cat .npmrc

# Test network connectivity and DNS resolution
RUN echo "Testing network connectivity..." && \
    nslookup registry.npmjs.org || echo "DNS lookup failed" && \
    wget -q --spider https://registry.npmjs.org/ || echo "Registry connection failed" && \
    echo "Network test completed"

# Install dependencies with enhanced network resilience
RUN npm config set registry https://registry.npmjs.org/ && \
    npm config set fetch-retries 5 && \
    npm config set fetch-retry-mintimeout 10000 && \
    npm config set fetch-retry-maxtimeout 60000 && \
    npm config set fetch-retry-factor 2 && \
    npm config set prefer-offline false && \
    npm config set audit false && \
    npm config set fund false && \
    npm config set update-notifier false && \
    npm config set network-timeout 60000 && \
    npm config set maxsockets 50 && \
    npm config set strict-ssl true && \
    npm config set ca "" && \
    npm config set cafile "" && \
    npm config set user-agent "npm/kaniko-builder" && \
    npm config list && \
    npm install --only=production --no-audit --no-fund --no-update-notifier --verbose 2>&1 | tee /tmp/npm-install.log || \
    (echo "First attempt failed, trying alternative registry..." && \
     npm config set registry https://registry.npmjs.com/ && \
     npm install --only=production --no-audit --no-fund --no-update-notifier --verbose 2>&1 | tee -a /tmp/npm-install.log) || \
    (echo "Second attempt failed, clearing cache and retrying..." && \
     npm cache clean --force && \
     npm config set registry https://registry.npmjs.org/ && \
     npm install --only=production --no-audit --no-fund --no-update-notifier --verbose 2>&1 | tee -a /tmp/npm-install.log) && \
    echo "npm install completed successfully" && \
    npm list --depth=0

# Verify installation and run npm ci for production build
RUN npm ci --only=production --no-audit --no-fund --verbose 2>&1 | tee /tmp/npm-ci.log && \
    echo "npm ci completed successfully"

# Stage 2: Production stage
FROM {{.NodeBaseImage}} as production

# Create non-root user for security
RUN addgroup -g 1001 -S notifi && \
    adduser -S nodejs -u 1001

# Set working directory
WORKDIR /app

# Copy node_modules from builder stage
COPY --from=builder --chown=nodejs:notifi /app/node_modules ./node_modules

# Copy application code
COPY --chown=nodejs:notifi . .

# Set environment variables
ENV NODE_ENV=production
ENV HTTP_PORT=8080

# Expose ports
EXPOSE 8080

# Switch to non-root user
USER nodejs

# Start the application
CMD ["npm", "start"]
`

// PackageJSONTemplate contains the package.json template content
const PackageJSONTemplate = `{
  "name": "{{.FunctionName}}",
  "version": "1.0.0",
  "description": "Knative Lambda Function for {{.ThirdPartyId}} parser {{.ParserId}}",
  "main": "index.js",
  "type": "module",
  "scripts": {
    "start": "node index.js"
  },
  "dependencies": {
    "cloudevents": "^10.0.0"
  },
  "engines": {
    "node": ">=22.0.0"
  },
  "keywords": ["knative", "lambda", "cloudevents", "parser", "notifi"],
  "author": "Knative Lambda Builder",
  "license": "MIT",
  "publishConfig": {
    "registry": "https://registry.npmjs.org/"
  },
  "repository": {
    "type": "git",
    "url": "https://github.com/notifi-network/knative-lambda"
  }
}`

// IndexJSTemplate contains the index.js template content
const IndexJSTemplate = `import { CloudEvent, HTTP } from 'cloudevents';
import { createServer } from 'http';
import { fileURLToPath } from 'url';
import { dirname } from 'path';
import * as parser from './{{.ParserId}}.js';

// ES module equivalent of __filename and __dirname for compatibility
const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

// Make them available globally for parser compatibility
globalThis.__filename = __filename;
globalThis.__dirname = __dirname;

// Configuration from environment variables
const PORT = process.env.HTTP_PORT || process.env.PORT || 8080;
const FUNCTION_NAME = process.env.FUNCTION_NAME || '{{.FunctionName}}';
const THIRD_PARTY_ID = process.env.THIRD_PARTY_ID || '{{.ThirdPartyId}}';
const PARSER_ID = process.env.PARSER_ID || '{{.ParserId}}';

// Shared readiness state and shutdown handling
let isReady = false;
let isShuttingDown = false;
let server = null;

// Create a simple logger for the context
const createLogger = () => ({
  info: (message) => console.log('[INFO]', message),
  error: (message) => console.error('[ERROR]', message),
  warn: (message) => console.warn('[WARN]', message),
  debug: (message) => console.log('[DEBUG]', message)
});

// Graceful shutdown handler
const gracefulShutdown = (signal) => {
  console.log('[INFO] Received shutdown signal:', signal);
  isShuttingDown = true;
  isReady = false;
  
  if (server) {
    console.log('[INFO] Closing HTTP server...');
    server.close((err) => {
      if (err) {
        console.error('[ERROR] Error during server shutdown:', err);
        process.exit(1);
      }
      console.log('[INFO] HTTP server closed gracefully');
      process.exit(0);
    });
    
    // Force shutdown after 30 seconds
    setTimeout(() => {
      console.error('[ERROR] Forced shutdown after timeout');
      process.exit(1);
    }, 30000);
  } else {
    process.exit(0);
  }
};

// Setup signal handlers for graceful shutdown
process.on('SIGTERM', () => gracefulShutdown('SIGTERM'));
process.on('SIGINT', () => gracefulShutdown('SIGINT'));
process.on('SIGUSR2', () => gracefulShutdown('SIGUSR2')); // For nodemon compatibility

// Handle uncaught exceptions and unhandled rejections
process.on('uncaughtException', (error) => {
  console.error('[ERROR] Uncaught Exception:', error);
  gracefulShutdown('uncaughtException');
});

process.on('unhandledRejection', (reason, promise) => {
  console.error('[ERROR] Unhandled Rejection at:', promise, 'reason:', reason);
  gracefulShutdown('unhandledRejection');
});

// HTTP server with CloudEvent support
server = createServer(async (req, res) => {
  // Check if server is shutting down
  if (isShuttingDown) {
    res.writeHead(503, { 'Content-Type': 'text/plain' });
    res.end('Service shutting down');
    return;
  }

  // Health check endpoints
  if (req.url === '/healthz') {
    res.writeHead(200, { 'Content-Type': 'text/plain' });
    res.end('ok');
    return;
  }

  if (req.url === '/readyz') {
    if (isReady && !isShuttingDown) {
      res.writeHead(200, { 'Content-Type': 'text/plain' });
      res.end('ok');
    } else {
      res.writeHead(503, { 'Content-Type': 'text/plain' });
      res.end('not ready');
    }
    return;
  }

  // CloudEvent endpoints (POST requests to / or /events)
  if ((req.url === '/' || req.url === '/events') && req.method === 'POST') {
    try {
      console.log('Received CloudEvent request:', req.method, req.url);
      //console.log('Content-Type:', req.headers['content-type']);

      // Collect request body
      let body = '';
      req.on('data', (chunk) => {
        body += chunk.toString();
      });

      await new Promise((resolve) => {
        req.on('end', resolve);
      });

      //console.log('Request body:', body);

      // Parse the CloudEvent in structured mode
      let event;
      try {
        const eventData = JSON.parse(body);
        //console.log('Parsed event data:', JSON.stringify(eventData, null, 2));
        
        // Check if this is a complete CloudEvent or just the data
        const isCompleteCloudEvent = eventData.source && eventData.type && eventData.id;
        
        if (isCompleteCloudEvent) {
          // Create CloudEvent with proper structure
          event = new CloudEvent({
            specversion: eventData.specversion || '1.0',
            id: eventData.id,
            source: eventData.source,
            type: eventData.type,
            subject: eventData.subject,
            time: eventData.time,
            datacontenttype: eventData.datacontenttype || 'application/json',
            data: eventData.data
          });
        } else {
          console.log('Processing data-only payload, reconstructing CloudEvent');
          //console.log('Data received:', JSON.stringify(eventData, null, 2));
          
          // Reconstruct CloudEvent from data and headers
          const source = req.headers['ce-source'] || 'network.notifi.unknown';
          const type = req.headers['ce-type'] || 'network.notifi.lambda.parser.start';
          const id = req.headers['ce-id'] || 'generated-' + Date.now();
          const subject = req.headers['ce-subject'] || 'unknown-parser';
          
          event = new CloudEvent({
            specversion: '1.0',
            id: id,
            source: source,
            type: type,
            subject: subject,
            time: new Date().toISOString(),
            datacontenttype: 'application/json',
            data: eventData
          });
        }
      } catch (parseError) {
        throw new Error('Failed to parse CloudEvent: ' + parseError.message);
      }

      console.log('Parsed CloudEvent:', {
        id: event.id,
        type: event.type,
        source: event.source,
        subject: event.subject
      });

      // Create context object with logger
      const context = {
        log: createLogger()
      };

      // Call the CloudEvent handler
      const result = await handleCloudEvent(event, context);

      // Send response
      if (result && result instanceof CloudEvent) {
        // Return CloudEvent response
        res.writeHead(200, { 
          'Content-Type': 'application/json',
          'Ce-Specversion': result.specversion || '1.0',
          'Ce-Type': result.type,
          'Ce-Source': result.source,
          'Ce-Id': result.id,
          'Ce-Time': result.time
        });
        res.end(JSON.stringify(result.data));
      } else {
        // Return simple success response
        res.writeHead(200, { 'Content-Type': 'application/json' });
        res.end(JSON.stringify({ success: true, message: 'Event processed successfully' }));
      }

    } catch (error) {
      console.error('Error processing CloudEvent:', error);
      res.writeHead(500, { 'Content-Type': 'application/json' });
      res.end(JSON.stringify({
        error: 'Failed to process CloudEvent',
        message: error.message
      }));
    }
    return;
  }

  // Default GET handler for root path
  if (req.url === '/' && req.method === 'GET') {
    res.writeHead(200, { 'Content-Type': 'text/plain' });
    res.end('Knative Lambda Function - ' + FUNCTION_NAME);
    return;
  }

  // 404 for all other requests
  res.writeHead(404, { 'Content-Type': 'text/plain' });
  res.end('not found');
});

server.listen(PORT, () => {
  console.log('Knative Lambda Function listening on port: ' + PORT);
  console.log('Function Name: ' + FUNCTION_NAME);
  console.log('Third Party ID: ' + THIRD_PARTY_ID);
  console.log('Parser ID: ' + PARSER_ID);
});

// Scheduler service configuration
const SCHEDULER_URL = process.env.SCHEDULER_URL || 'http://notifi-scheduler.notifi.svc.cluster.local/fusion/execution/response';

/**
 * Main CloudEvent handler function
 * Processes CloudEvents and executes the parser
 */
const handleCloudEvent = async (event, context) => {
  const startTime = Date.now();
  console.log('Received CloudEvent with context:', context);

  // Validate event and event.data
  if (!event) {
    throw new Error('CloudEvent is null or undefined');
  }
  
  if (!event.data) {
    throw new Error('CloudEvent data is null or undefined');
  }

  // Ensure contextId exists
  const contextId = event.data.contextId || 'unknown-context';
   
  let processed = undefined;
  let result = undefined;

  try {
    // Validate parser module
    if (!parser || typeof parser !== 'object') {
      throw new Error('Parser module is not available or not an object');
    }

    const parserMethods = Object.keys(parser);

    // Validate event data before passing to parser
    if (event.data && typeof event.data === 'object') {
      // Add defensive programming - ensure data has expected structure
      const safeEventData = {
        ...event.data,
        contextId: contextId
      };

      if (parser.handle && typeof parser.handle === 'function') {
        context.log.info('Calling parser.handle with event data');
        processed = parser.handle(safeEventData);
      } else if (parser.parse && typeof parser.parse === 'function') {
        context.log.info('Calling parser.parse with event data');
        processed = parser.parse(safeEventData);
      } else {
        throw new Error('Parser module does not have handle() or parse() method. Available methods: ' + parserMethods.join(', '));
      }
    } else {
      throw new Error('Event data is not a valid object: ' + typeof event.data);
    }

    context.log.info('Parser call completed, result type: ' + typeof processed);
  } catch (error) {
    const errorMessageTrimmed = error.message.length > 512 ? error.message.substring(0, 509) + '...' : error.message;
    context.log.error("Error processing event with parser:", error.message);
    context.log.error("Error stack:", error.stack);
    result = {
      errorMessage: error.name + ': ' + errorMessageTrimmed,
      succeeded: false,
      contextId: contextId,
    };
  }

  if (!result) {
    try {
      // If the parser returns a promise, await it
      if (processed instanceof Promise) {
        context.log.info("Processing data asynchronously...");
        const appResult = await processed;
        // TODO: Validate appResult schema
        result = {
          succeeded: true,
          contextId: contextId,
          eventEntries: appResult,
        };
      } else {
        context.log.info("Processing data synchronously...");
        result = {
          succeeded: true,
          contextId: contextId,
          eventEntries: processed,
        };
      }
    } catch (error) {
      const errorMessageTrimmed = error.message.length > 512 ? error.message.substring(0, 509) + '...' : error.message;
      context.log.error("Error processing event with parser:", error);
      result = {
        errorMessage: error.name + ': ' + errorMessageTrimmed,
        succeeded: false,
        contextId: contextId,
      };
    }
  }

  // Handle sync vs async response modes
  const shouldHandleAsSync = event.responsetype === 'sync';
  
  // Debug: Log response mode decision
  context.log.info('Response mode check: event.responsetype = "' + event.responsetype + '", shouldHandleAsSync = ' + shouldHandleAsSync);

  if (shouldHandleAsSync) {
    context.log.info("Handling as synchronous response, returning result directly.");

    return new CloudEvent({
      source: event.source,
      type: 'network.notifi.lambda.result',
      data: result,
      subject: event.subject,
      id: event.id + '-response',
      time: new Date().toISOString()
    });
  }

  // Send to scheduler service asynchronously
  //context.log.info("Handling as asynchronous response, sending to Scheduler.");
  
  try {
    context.log.info('Sending processed data to scheduler service.');
    
    const response = await fetch(SCHEDULER_URL, {
      method: 'POST',
      headers: {
        'User-Agent': 'knative-lambda-handler/' + FUNCTION_NAME,
        'Content-Type': 'application/json',
        'Accept': 'application/json',
      },
      body: JSON.stringify(result),
    });

    if (!response.ok) {
      const errorText = await response.text();
      throw new Error('Scheduler service responded with ' + response.status + ': ' + response.statusText + ' - ' + errorText);
    }
    
    const responseData = await response.json();
    //context.log.info('Scheduler service response: ' + JSON.stringify(responseData));
    context.log.info('Successfully sent processed data to scheduler service');
    
  } catch (error) {
    context.log.error('Failed to send data to scheduler service: ' + error.message);
    
    // Return error response instead of throwing to allow proper HTTP response
    return new CloudEvent({
      source: event.source,
      type: 'network.notifi.lambda.error',
      data: {
        errorMessage: 'Failed to send data to scheduler service: ' + error.message,
        succeeded: false,
        contextId: contextId,
        functionName: FUNCTION_NAME,
        timestamp: new Date().toISOString()
      },
      subject: event.subject,
      id: event.id + '-error',
      time: new Date().toISOString()
    });
  }

  context.log.info('Successfully processed CloudEvent: ' + event.id);

  // Return success response
  return new CloudEvent({
    source: event.source,
    type: 'network.notifi.lambda.processed',
    data: result,
    subject: event.subject,
    id: event.id + '-processed',
    time: new Date().toISOString()
  });
};

isReady = true;

// Export the handler for testing
export { handleCloudEvent };`

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🔧 TEMPLATE HELPERS - "Helper functions for template processing"      │
// └─────────────────────────────────────────────────────────────────────────┘

// CreateTemplateData creates template data from a build request
// Only includes fields that are actually used in templates
func CreateTemplateData(buildRequest *builds.BuildRequest, nodeBaseImage string, runtimeCMD string) TemplateData {
	// Generate function name using the naming helper for proper validation and truncation
	functionName := helpers.GenerateFunctionName(buildRequest.ThirdPartyID, buildRequest.ParserID)

	return TemplateData{
		FunctionName:  functionName,
		ThirdPartyId:  buildRequest.ThirdPartyID,
		ParserId:      buildRequest.ParserID,
		NodeBaseImage: nodeBaseImage,
		Timestamp:     time.Now(),
	}
}

// ProcessDockerfileTemplate processes the Dockerfile template
func (tp *TemplateProcessor) ProcessDockerfileTemplate(ctx context.Context, data TemplateData) ([]byte, error) {
	return tp.ProcessTemplate(ctx, "Dockerfile", DockerfileTemplate, data)
}

// ProcessIndexJSTemplate processes the index.js template
func (tp *TemplateProcessor) ProcessIndexJSTemplate(ctx context.Context, data TemplateData) ([]byte, error) {
	return tp.ProcessTemplate(ctx, "index.js", IndexJSTemplate, data)
}

// ProcessPackageJSONTemplate processes the package.json template
func (tp *TemplateProcessor) ProcessPackageJSONTemplate(ctx context.Context, data TemplateData) ([]byte, error) {
	return tp.ProcessTemplate(ctx, "package.json", PackageJSONTemplate, data)
}
