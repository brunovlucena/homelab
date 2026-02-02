// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//
//	📜 JSON SCHEMAS - CloudEvent Payload Definitions
//
//	This file contains JSON Schema definitions for all CloudEvent types
//	supported by the Knative Lambda Operator.
//
//	Schema Version: 1.0.0
//	JSON Schema Draft: 2020-12
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
package schema

// SchemaVersion is the current version of the schemas
const SchemaVersion = "1.0.0"

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🚀 FUNCTION DEPLOY SCHEMA                                              │
// │  Event Type: io.knative.lambda.command.function.deploy                  │
// └─────────────────────────────────────────────────────────────────────────┘

// FunctionDeploySchema validates function.deploy CloudEvent payloads
const FunctionDeploySchema = `{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "io.knative.lambda.command.function.deploy",
  "title": "Function Deploy Event",
  "description": "Schema for deploying or updating a Lambda function",
  "type": "object",
  "required": ["metadata", "spec"],
  "additionalProperties": false,
  "properties": {
    "metadata": {
      "type": "object",
      "required": ["name"],
      "properties": {
        "name": {
          "type": "string",
          "minLength": 1,
          "maxLength": 63,
          "pattern": "^[a-z0-9]([-a-z0-9]*[a-z0-9])?$",
          "description": "Name of the Lambda function (RFC 1123 DNS label)"
        },
        "namespace": {
          "type": "string",
          "minLength": 1,
          "maxLength": 63,
          "pattern": "^[a-z0-9]([-a-z0-9]*[a-z0-9])?$",
          "description": "Kubernetes namespace"
        },
        "labels": {
          "type": "object",
          "additionalProperties": { "type": "string" }
        },
        "annotations": {
          "type": "object",
          "additionalProperties": { "type": "string" }
        }
      }
    },
    "spec": {
      "type": "object",
      "required": ["source", "runtime"],
      "properties": {
        "source": {
          "$ref": "#/$defs/sourceSpec"
        },
        "runtime": {
          "$ref": "#/$defs/runtimeSpec"
        },
        "scaling": {
          "$ref": "#/$defs/scalingSpec"
        },
        "resources": {
          "$ref": "#/$defs/resourceSpec"
        },
        "env": {
          "type": "array",
          "items": {
            "$ref": "#/$defs/envVar"
          }
        },
        "build": {
          "$ref": "#/$defs/buildSpec"
        },
        "eventing": {
          "type": "object",
          "properties": {
            "enabled": { "type": "boolean" }
          }
        }
      }
    }
  },
  "$defs": {
    "sourceSpec": {
      "type": "object",
      "required": ["type"],
      "properties": {
        "type": {
          "type": "string",
          "enum": ["minio", "s3", "gcs", "git", "inline", "image"],
          "description": "Source type for Lambda function code"
        },
        "minio": {
          "type": "object",
          "required": ["bucket", "key"],
          "properties": {
            "endpoint": { "type": "string" },
            "bucket": { "type": "string", "minLength": 1 },
            "key": { "type": "string", "minLength": 1 },
            "secretRef": { "$ref": "#/$defs/secretRef" }
          }
        },
        "s3": {
          "type": "object",
          "required": ["bucket", "key"],
          "properties": {
            "bucket": { "type": "string", "minLength": 1 },
            "key": { "type": "string", "minLength": 1 },
            "region": { "type": "string" },
            "secretRef": { "$ref": "#/$defs/secretRef" }
          }
        },
        "gcs": {
          "type": "object",
          "required": ["bucket", "key"],
          "properties": {
            "bucket": { "type": "string", "minLength": 1 },
            "key": { "type": "string", "minLength": 1 },
            "project": { "type": "string" },
            "secretRef": { "$ref": "#/$defs/secretRef" }
          }
        },
        "git": {
          "type": "object",
          "required": ["url"],
          "properties": {
            "url": { 
              "type": "string", 
              "minLength": 1,
              "pattern": "^(https?://|git@|ssh://)"
            },
            "ref": { "type": "string" },
            "path": { "type": "string" },
            "secretRef": { "$ref": "#/$defs/secretRef" }
          }
        },
        "inline": {
          "type": "object",
          "required": ["code"],
          "properties": {
            "code": { "type": "string", "minLength": 1 },
            "dependencies": { "type": "string" }
          }
        },
        "image": {
          "type": "object",
          "required": ["repository"],
          "properties": {
            "repository": { "type": "string", "minLength": 1 },
            "tag": { "type": "string" },
            "digest": { "type": "string" },
            "pullPolicy": { 
              "type": "string",
              "enum": ["Always", "IfNotPresent", "Never"]
            },
            "port": { "type": "integer", "minimum": 1, "maximum": 65535 }
          }
        }
      },
      "allOf": [
        {
          "if": { "properties": { "type": { "const": "minio" } } },
          "then": { "required": ["minio"] }
        },
        {
          "if": { "properties": { "type": { "const": "s3" } } },
          "then": { "required": ["s3"] }
        },
        {
          "if": { "properties": { "type": { "const": "gcs" } } },
          "then": { "required": ["gcs"] }
        },
        {
          "if": { "properties": { "type": { "const": "git" } } },
          "then": { "required": ["git"] }
        },
        {
          "if": { "properties": { "type": { "const": "inline" } } },
          "then": { "required": ["inline"] }
        },
        {
          "if": { "properties": { "type": { "const": "image" } } },
          "then": { "required": ["image"] }
        }
      ]
    },
    "runtimeSpec": {
      "type": "object",
      "required": ["language", "version"],
      "properties": {
        "language": {
          "type": "string",
          "enum": ["python", "nodejs", "go"],
          "description": "Programming language"
        },
        "version": {
          "type": "string",
          "minLength": 1,
          "description": "Language version (e.g., 3.11, 20, 1.21)"
        },
        "handler": {
          "type": "string",
          "description": "Handler function name"
        }
      }
    },
    "scalingSpec": {
      "type": "object",
      "properties": {
        "minReplicas": { "type": "integer", "minimum": 0 },
        "maxReplicas": { "type": "integer", "minimum": 1 },
        "targetConcurrency": { "type": "integer", "minimum": 1 },
        "scaleToZeroGracePeriod": { "type": "string" }
      }
    },
    "resourceSpec": {
      "type": "object",
      "properties": {
        "requests": { "$ref": "#/$defs/resourceRequirements" },
        "limits": { "$ref": "#/$defs/resourceRequirements" }
      }
    },
    "resourceRequirements": {
      "type": "object",
      "properties": {
        "memory": { "type": "string", "pattern": "^[0-9]+[KMGkmg]i?$" },
        "cpu": { "type": "string", "pattern": "^[0-9]+m?$" }
      }
    },
    "buildSpec": {
      "type": "object",
      "properties": {
        "timeout": { "type": "string" },
        "registry": { "type": "string" },
        "registryType": { 
          "type": "string",
          "enum": ["local", "ecr", "gcr", "ghcr", "dockerhub", "generic"]
        },
        "repository": { "type": "string" },
        "tag": { "type": "string" },
        "insecure": { "type": "boolean" },
        "forceRebuild": { "type": "boolean" }
      }
    },
    "envVar": {
      "type": "object",
      "required": ["name"],
      "properties": {
        "name": { "type": "string", "minLength": 1 },
        "value": { "type": "string" },
        "valueFrom": { "type": "object" }
      }
    },
    "secretRef": {
      "type": "object",
      "properties": {
        "name": { "type": "string", "minLength": 1 }
      }
    }
  }
}`

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🗑️ SERVICE DELETE SCHEMA                                               │
// │  Event Type: io.knative.lambda.command.service.delete                   │
// └─────────────────────────────────────────────────────────────────────────┘

// ServiceDeleteSchema validates service.delete CloudEvent payloads
const ServiceDeleteSchema = `{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "io.knative.lambda.command.service.delete",
  "title": "Service Delete Event",
  "description": "Schema for deleting a Lambda function",
  "type": "object",
  "properties": {
    "name": {
      "type": "string",
      "minLength": 1,
      "maxLength": 63,
      "pattern": "^[a-z0-9]([-a-z0-9]*[a-z0-9])?$",
      "description": "Name of the Lambda function to delete"
    },
    "namespace": {
      "type": "string",
      "minLength": 1,
      "maxLength": 63,
      "description": "Kubernetes namespace"
    }
  }
}`

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🔨 BUILD COMMAND SCHEMA                                                │
// │  Event Types: io.knative.lambda.command.build.*                         │
// └─────────────────────────────────────────────────────────────────────────┘

// BuildCommandSchema validates build command CloudEvent payloads
const BuildCommandSchema = `{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "io.knative.lambda.command.build",
  "title": "Build Command Event",
  "description": "Schema for build commands (start, cancel, retry)",
  "type": "object",
  "required": ["name"],
  "properties": {
    "name": {
      "type": "string",
      "minLength": 1,
      "maxLength": 63,
      "pattern": "^[a-z0-9]([-a-z0-9]*[a-z0-9])?$",
      "description": "Name of the Lambda function"
    },
    "namespace": {
      "type": "string",
      "minLength": 1,
      "maxLength": 63,
      "description": "Kubernetes namespace"
    },
    "forceRebuild": {
      "type": "boolean",
      "description": "Force rebuild even if image exists"
    }
  }
}`

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🚀 INVOKE SCHEMA                                                       │
// │  Event Types: io.knative.lambda.invoke.*                                │
// └─────────────────────────────────────────────────────────────────────────┘

// InvokeSchema validates invoke CloudEvent payloads
const InvokeSchema = `{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "io.knative.lambda.invoke",
  "title": "Invoke Event",
  "description": "Schema for invoking a Lambda function",
  "type": "object",
  "properties": {
    "payload": {
      "description": "Payload to pass to the Lambda function (any JSON value)"
    },
    "correlationId": {
      "type": "string",
      "description": "Correlation ID for tracking"
    },
    "contextId": {
      "type": "string",
      "description": "Context ID for tracking"
    },
    "timeout": {
      "type": "string",
      "pattern": "^[0-9]+(s|m|h)$",
      "description": "Execution timeout (e.g., 30s, 5m)"
    }
  }
}`

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  📊 LIFECYCLE BUILD SCHEMA                                              │
// │  Event Types: io.knative.lambda.lifecycle.build.*                       │
// └─────────────────────────────────────────────────────────────────────────┘

// LifecycleBuildSchema validates build lifecycle CloudEvent payloads
const LifecycleBuildSchema = `{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "io.knative.lambda.lifecycle.build",
  "title": "Build Lifecycle Event",
  "description": "Schema for build lifecycle events",
  "type": "object",
  "required": ["name", "namespace"],
  "properties": {
    "name": {
      "type": "string",
      "description": "Name of the Lambda function"
    },
    "namespace": {
      "type": "string",
      "description": "Kubernetes namespace"
    },
    "jobName": {
      "type": "string",
      "description": "Name of the build job"
    },
    "imageURI": {
      "type": "string",
      "description": "URI of the built image"
    },
    "startedAt": {
      "type": "string",
      "format": "date-time",
      "description": "Build start time"
    },
    "completedAt": {
      "type": "string",
      "format": "date-time",
      "description": "Build completion time"
    },
    "duration": {
      "type": "string",
      "description": "Build duration"
    },
    "error": {
      "type": "string",
      "description": "Error message if build failed"
    },
    "attempt": {
      "type": "integer",
      "minimum": 1,
      "description": "Build attempt number"
    }
  }
}`

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  📨 RESPONSE SCHEMA                                                     │
// │  Event Types: io.knative.lambda.response.*                              │
// └─────────────────────────────────────────────────────────────────────────┘

// ResponseSchema validates response CloudEvent payloads
const ResponseSchema = `{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "io.knative.lambda.response",
  "title": "Response Event",
  "description": "Schema for Lambda function execution responses",
  "type": "object",
  "properties": {
    "statusCode": {
      "type": "integer",
      "minimum": 100,
      "maximum": 599,
      "description": "HTTP status code"
    },
    "body": {
      "description": "Response body (any JSON value)"
    },
    "headers": {
      "type": "object",
      "additionalProperties": { "type": "string" },
      "description": "Response headers"
    },
    "error": {
      "type": "string",
      "description": "Error message if execution failed"
    },
    "errorType": {
      "type": "string",
      "description": "Error type/category"
    },
    "stackTrace": {
      "type": "array",
      "items": { "type": "string" },
      "description": "Stack trace for errors"
    },
    "duration": {
      "type": "string",
      "description": "Execution duration"
    },
    "correlationId": {
      "type": "string",
      "description": "Correlation ID from the invoke event"
    }
  }
}`
