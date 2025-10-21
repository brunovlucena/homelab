{{/*
Expand the name of the chart.
*/}}
{{- define "knative-lambda.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "knative-lambda.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "knative-lambda.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "knative-lambda.labels" -}}
helm.sh/chart: {{ include "knative-lambda.chart" . }}
app.kubernetes.io/name: {{ .Values.app.name }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/component: {{ .Values.app.component }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/part-of: {{ .Values.app.partOf }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "knative-lambda.selectorLabels" -}}
app.kubernetes.io/name: {{ .Values.app.name }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/component: {{ .Values.app.component }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "knative-lambda.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default .Values.app.name .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Create the namespace name
*/}}
{{- define "knative-lambda.namespace" -}}
{{- if .Values.namespace.create }}
{{- .Values.namespace.name }}
{{- else }}
{{- .Release.Namespace }}
{{- end }}
{{- end }}

{{/*
Create the full image name
*/}}
{{- define "knative-lambda.image" -}}
{{- printf "%s/%s:%s" .Values.image.registry .Values.image.repository .Values.image.tag }}
{{- end }}

{{/*
Create Knative broker URL
*/}}
{{- define "knative-lambda.brokerUrl" -}}
{{- if .Values.broker.url -}}
{{- .Values.broker.url -}}
{{- else -}}
{{- printf "http://%s-broker-ingress.%s.svc.cluster.local" (default (printf "%s-broker-%s" .Values.app.name .Values.environment) .Values.env.brokerName) (include "knative-lambda.namespace" .) -}}
{{- end -}}
{{- end }}

{{/*
Create Knative broker port
*/}}
{{- define "knative-lambda.brokerPort" -}}
{{- if .Values.broker.port -}}
{{- .Values.broker.port -}}
{{- else -}}
{{- "80" -}}
{{- end -}}
{{- end }}

{{/*
Create common environment variables
*/}}
{{- define "knative-lambda.env" -}}
- name: HTTP_PORT
  value: {{ .Values.env.httpPort | quote }}
- name: METRICS_PORT
  value: {{ .Values.env.metricsPort | quote }}
- name: S3_SOURCE_BUCKET
  value: {{ .Values.env.s3SourceBucket | quote }}
- name: S3_TEMP_BUCKET
  value: {{ .Values.env.s3TmpBucket | quote }}
- name: ECR_REGISTRY
  value: {{ .Values.env.ecrBaseRegistry | quote }}
- name: AWS_REGION
  value: {{ .Values.env.awsRegion | quote }}
- name: AWS_ACCOUNT_ID
  value: {{ .Values.aws.accountId | quote }}
- name: WORKER_POOL_SIZE
  value: {{ .Values.env.workerPoolSize | quote }}
- name: WORKER_POOL_CAPACITY
  value: {{ .Values.env.workerPoolCapacity | quote }}
- name: EVENT_QUEUE_SIZE
  value: {{ .Values.env.eventQueueSize | quote }}
- name: LOG_LEVEL
  value: {{ .Values.env.logLevel | quote }}
- name: TRACING_ENABLED
  value: {{ .Values.env.tracingEnabled | quote }}
- name: OTEL_EXPORTER_OTLP_ENDPOINT
  value: {{ .Values.env.otelExporterOtlpEndpoint | quote }}

{{- /* DEFAULT_TRIGGER_NAMESPACE is set in builder.yaml template */ -}}
{{- if .Values.env.kubernetesNamespace }}
- name: NAMESPACE
  value: {{ .Values.env.kubernetesNamespace | quote }}
{{- end }}
{{- if .Values.env.apiTimeout }}
- name: API_TIMEOUT
  value: {{ .Values.env.apiTimeout | quote }}
{{- end }}
{{- end }}

{{/*
Create prometheus annotations
*/}}
{{- define "knative-lambda.prometheusAnnotations" -}}
{{- if .Values.monitoring.enabled }}
prometheus.io/scrape: {{ .Values.monitoring.prometheus.scrape | quote }}
prometheus.io/port: {{ .Values.monitoring.prometheus.port | quote }}
prometheus.io/path: {{ .Values.monitoring.prometheus.path | quote }}
{{- end }}
{{- end }}

{{/*
Create AWS IAM role ARN
*/}}
{{- define "knative-lambda.awsRoleArn" -}}
{{- printf "arn:aws:iam::%s:role/%s" .Values.aws.accountId .Values.roleName }}
{{- end }}

{{/*
Create ECR repository URL
*/}}
{{- define "knative-lambda.ecrRepository" -}}
{{- printf "%s.dkr.ecr.%s.amazonaws.com/%s" .Values.aws.accountId .Values.aws.region .Values.image.repository }}
{{- end }}

{{/*
Create full ECR image URL
*/}}
{{- define "knative-lambda.ecrImage" -}}
{{- printf "%s:%s" (include "knative-lambda.ecrRepository" .) .Values.image.tag }}
{{- end }}

{{/*
Create full sidecar ECR image URL
*/}}
{{- define "knative-lambda.sidecarImage" -}}
{{- printf "%s/%s:%s" .Values.sidecar.image.registry .Values.sidecar.image.repository .Values.sidecar.image.tag }}
{{- end }}

{{/*
Create full metrics pusher ECR image URL
*/}}
{{- define "knative-lambda.metricsPusherImage" -}}
{{- printf "%s/%s:%s" .Values.metricsPusher.image.registry .Values.metricsPusher.image.repository .Values.metricsPusher.image.tag }}
{{- end }}

{{/*
Create Knative service autoscaling annotations
*/}}
{{- define "knative-lambda.autoscalingAnnotations" -}}
autoscaling.knative.dev/class: "kpa.autoscaling.knative.dev"
autoscaling.knative.dev/minScale: {{ .minScale | quote }}
autoscaling.knative.dev/maxScale: {{ .maxScale | quote }}
autoscaling.knative.dev/target: {{ .targetConcurrency | quote }}
autoscaling.knative.dev/scaleToZeroGracePeriod: {{ .scaleToZeroGracePeriod | quote }}
autoscaling.knative.dev/scaleDownDelay: {{ .scaleDownDelay | quote }}
autoscaling.knative.dev/stableWindow: {{ .stableWindow | quote }}
{{- end }}

{{/*
Create Knative service networking annotations
*/}}
{{- define "knative-lambda.networkingAnnotations" -}}
networking.knative.dev/ingress.class: "kourier.ingress.networking.knative.dev"
{{- end }}

{{/*
Create common annotations for dynamic services
*/}}
{{- define "knative-lambda.dynamicServiceAnnotations" -}}
{{- $config := .Values.dynamicServices.defaults }}
{{- include "knative-lambda.autoscalingAnnotations" $config }}
{{ include "knative-lambda.networkingAnnotations" . }}
{{- range $key, $value := .Values.dynamicServices.template.annotations }}
{{ $key }}: {{ $value | quote }}
{{- end }}
{{- end }}

{{/*
Create RabbitMQ connection string
*/}}
{{- define "knative-lambda.rabbitmqConnectionString" -}}
{{- printf "amqp://notifi:notifi@%s:5672/%%2F" (include "knative-lambda.rabbitmqHost" .) }}
{{- end }}

{{/*
Create RabbitMQ namespace with environment suffix
*/}}
{{- define "knative-lambda.rabbitmqNamespace" -}}
{{- printf "rabbitmq-%s" .Values.environment }}
{{- end }}

{{/*
Create RabbitMQ cluster host with environment-specific namespace
*/}}
{{- define "knative-lambda.rabbitmqHost" -}}
{{- printf "rabbitmq-cluster-%s.rabbitmq-%s.svc.cluster.local" .Values.environment .Values.environment }}
{{- end }}

{{/*
Create sidecar environment variables
*/}}
{{- define "knative-lambda.sidecarEnv" -}}
# Kaniko monitoring configuration
- name: KANIKO_NAMESPACE
  value: {{ include "knative-lambda.namespace" . }}
- name: KANIKO_POD_NAME
  valueFrom:
    fieldRef:
      fieldPath: metadata.name
- name: KANIKO_CONTAINER_NAME
  value: {{ .Values.sidecar.monitoring.kanikoContainerName | quote }}
- name: MONITOR_INTERVAL
  value: {{ .Values.sidecar.monitoring.pollInterval | quote }}
- name: BUILD_TIMEOUT
  value: {{ .Values.sidecar.monitoring.buildTimeout | quote }}

# Knative broker configuration
- name: KNATIVE_BROKER_URL
  value: {{ include "knative-lambda.brokerUrl" . }}

# Logging configuration
- name: LOG_LEVEL
  value: {{ .Values.sidecar.logging.level | quote }}
- name: LOG_FORMAT
  value: {{ .Values.sidecar.logging.format | quote }}
- name: SERVICE_NAME
  value: "sidecar"

# Security configuration
- name: TLS_ENABLED
  value: "false"
- name: METRICS_ENABLED
  value: "true"
- name: METRICS_PORT
  value: "9092"
- name: METRICS_PATH
  value: "/metrics"

# Runtime job-specific variables (to be set dynamically by builder)
# - BUILD_JOB_NAME
# - IMAGE_URI  
# - THIRD_PARTY_ID
# - PARSER_ID
# - CORRELATION_ID
{{- end }}

{{/*
Create sidecar security context
*/}}
{{- define "knative-lambda.sidecarSecurityContext" -}}
runAsUser: {{ .Values.sidecar.security.runAsUser }}
runAsGroup: {{ .Values.sidecar.security.runAsGroup }}
runAsNonRoot: {{ .Values.sidecar.security.runAsNonRoot }}
readOnlyRootFilesystem: {{ .Values.sidecar.security.readOnlyRootFilesystem }}
allowPrivilegeEscalation: {{ .Values.sidecar.security.allowPrivilegeEscalation }}
capabilities:
  {{- toYaml .Values.sidecar.security.capabilities | nindent 2 }}
{{- end }}

{{/*
Create sidecar container definition for use in Kaniko jobs
*/}}
{{- define "knative-lambda.sidecarContainer" -}}
name: sidecar
image: {{ include "knative-lambda.sidecarImage" . }}
imagePullPolicy: {{ .Values.sidecar.image.pullPolicy }}
env:
{{- include "knative-lambda.sidecarEnv" . | nindent 2 }}
resources:
  {{- toYaml .Values.sidecar.resources | nindent 2 }}
securityContext:
  {{- include "knative-lambda.sidecarSecurityContext" . | nindent 2 }}
{{- end }} 

{{/*
🔧 VALIDATION HELPERS - Ensure proper data types for Knative configuration
   🚨 CRITICAL: Prevents "cannot unmarshal string into Go struct field" errors
   📊 Purpose: Validate and convert values to proper types before template rendering
*/}}

{{/*
🔢 VALIDATE INTEGER - Ensures value is an integer, not a string
   Usage: {{ include "validateInteger" (dict "value" .Values.someValue "default" 50) }}
   Returns: Integer value, never a string
*/}}
{{- define "validateInteger" -}}
{{- $value := .value | default .default -}}
{{- if kindIs "string" $value -}}
{{- /* Convert string to int, with fallback to default */ -}}
{{- $intValue := atoi $value -}}
{{- if eq $intValue 0 -}}
{{- /* If atoi returns 0, check if original was "0" */ -}}
{{- if eq $value "0" -}}
0
{{- else -}}
{{- .default -}}
{{- end -}}
{{- else -}}
{{- $intValue -}}
{{- end -}}
{{- else -}}
{{- $value -}}
{{- end -}}
{{- end -}}

{{/*
🎯 VALIDATE CONCURRENCY VALUES - Ensures all concurrency-related values are integers
   Usage: {{ include "validateConcurrency" (dict "value" .Values.containerConcurrency "default" 50) }}
   Returns: Integer value for Knative concurrency fields
*/}}
{{- define "validateConcurrency" -}}
{{- include "validateInteger" (dict "value" .value "default" .default) -}}
{{- end -}}

{{/*
📊 VALIDATE SCALING VALUES - Ensures all scaling-related values are integers
   Usage: {{ include "validateScaling" (dict "value" .Values.maxScale "default" 10) }}
   Returns: Integer value for Knative scaling fields
*/}}
{{- define "validateScaling" -}}
{{- include "validateInteger" (dict "value" .value "default" .default) -}}
{{- end -}}

{{/*
⏰ VALIDATE DURATION - Ensures duration values are strings (for time.Duration)
   Usage: {{ include "validateDuration" (dict "value" .Values.timeout "default" "30s") }}
   Returns: String value for duration fields
*/}}
{{- define "validateDuration" -}}
{{- $value := .value | default .default -}}
{{- if kindIs "string" $value -}}
{{- $value -}}
{{- else -}}
{{- .default -}}
{{- end -}}
{{- end -}}

{{/*
🔍 VALIDATE KNATIVE SERVICE CONFIG - Comprehensive validation for Knative service creation
   Usage: {{ include "validateKnativeConfig" . }}
   Returns: Validated configuration object
*/}}
{{- define "validateKnativeConfig" -}}
{{- $config := dict -}}
{{- $_ := set $config "containerConcurrency" (include "validateConcurrency" (dict "value" .Values.containerConcurrency "default" 50)) -}}
{{- $_ := set $config "targetConcurrency" (include "validateConcurrency" (dict "value" .Values.targetConcurrency "default" 70)) -}}
{{- $_ := set $config "targetUtilization" (include "validateConcurrency" (dict "value" .Values.targetUtilization "default" 50)) -}}
{{- $_ := set $config "target" (include "validateConcurrency" (dict "value" .Values.target "default" 70)) -}}
{{- $_ := set $config "minScale" (include "validateScaling" (dict "value" .Values.minScale "default" 0)) -}}
{{- $_ := set $config "maxScale" (include "validateScaling" (dict "value" .Values.maxScale "default" 50)) -}}
{{- $_ := set $config "scaleToZeroGracePeriod" (include "validateDuration" (dict "value" .Values.scaleToZeroGracePeriod "default" "30s")) -}}
{{- $_ := set $config "scaleDownDelay" (include "validateDuration" (dict "value" .Values.scaleDownDelay "default" "0s")) -}}
{{- $_ := set $config "stableWindow" (include "validateDuration" (dict "value" .Values.stableWindow "default" "10s")) -}}
{{- $config | toJson -}}
{{- end -}}

{{/*
🚨 ERROR PREVENTION - Template function to prevent common Knative configuration errors
   Usage: {{ include "preventKnativeErrors" . }}
   Purpose: Validates configuration before rendering to prevent admission webhook failures
*/}}
{{- define "preventKnativeErrors" -}}
{{- /* Validate that critical numeric fields are integers, not strings */ -}}
{{- $errors := list -}}

{{- /* Check containerConcurrency */ -}}
{{- if kindIs "string" .Values.containerConcurrency -}}
{{- $errors = append $errors "containerConcurrency must be integer, not string" -}}
{{- end -}}

{{- /* Check targetConcurrency */ -}}
{{- if kindIs "string" .Values.targetConcurrency -}}
{{- $errors = append $errors "targetConcurrency must be integer, not string" -}}
{{- end -}}

{{- /* Check target */ -}}
{{- if kindIs "string" .Values.target -}}
{{- $errors = append $errors "target must be integer, not string" -}}
{{- end -}}

{{- /* Check targetUtilization */ -}}
{{- if kindIs "string" .Values.targetUtilization -}}
{{- $errors = append $errors "targetUtilization must be integer, not string" -}}
{{- end -}}

{{- /* Check minScale */ -}}
{{- if kindIs "string" .Values.minScale -}}
{{- $errors = append $errors "minScale must be integer, not string" -}}
{{- end -}}

{{- /* Check maxScale */ -}}
{{- if kindIs "string" .Values.maxScale -}}
{{- $errors = append $errors "maxScale must be integer, not string" -}}
{{- end -}}

{{- /* If there are errors, fail the template */ -}}
{{- if gt (len $errors) 0 -}}
{{- printf "🚨 KNATIVE CONFIGURATION ERRORS:\n%s" (join $errors "\n") | fail -}}
{{- end -}}
{{- end -}} 