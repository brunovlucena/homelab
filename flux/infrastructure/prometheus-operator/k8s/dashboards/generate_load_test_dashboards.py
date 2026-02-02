#!/usr/bin/env python3
"""Generate Grafana dashboards for K6 load tests"""
import json
import os

def create_panel_base(panel_id, title, grid_pos, targets, panel_type='timeseries', unit='short', decimals=None):
    """Create a base panel configuration"""
    config = {
        'datasource': {'type': 'prometheus', 'uid': 'prometheus'},
        'fieldConfig': {
            'defaults': {
                'color': {'mode': 'thresholds'},
                'mappings': [],
                'thresholds': {'mode': 'absolute', 'steps': [{'color': 'green', 'value': None}]},
                'unit': unit,
            },
            'overrides': []
        },
        'gridPos': grid_pos,
        'id': panel_id,
        'options': {},
        'targets': targets,
        'title': title,
        'type': panel_type,
    }
    if decimals is not None:
        config['fieldConfig']['defaults']['decimals'] = decimals
    if panel_type == 'timeseries':
        config['fieldConfig']['defaults']['color'] = {'mode': 'palette-classic'}
        config['fieldConfig']['defaults']['custom'] = {
            'axisCenteredZero': False,
            'axisLabel': '',
            'axisPlacement': 'auto',
            'drawStyle': 'line',
            'fillOpacity': 20,
            'gradientMode': 'none',
            'lineInterpolation': 'smooth',
            'lineWidth': 2,
            'pointSize': 5,
            'scaleDistribution': {'type': 'linear'},
            'showPoints': 'never',
            'spanNulls': False,
            'stacking': {'group': 'A', 'mode': 'none'},
            'thresholdsStyle': {'mode': 'off'}
        }
        config['options'] = {
            'legend': {'calcs': ['mean', 'max'], 'displayMode': 'table', 'placement': 'bottom', 'showLegend': True},
            'tooltip': {'mode': 'multi', 'sort': 'desc'}
        }
    elif panel_type == 'stat':
        config['options'] = {
            'colorMode': 'value',
            'graphMode': 'area',
            'justifyMode': 'auto',
            'orientation': 'auto',
            'reduceOptions': {'calcs': ['lastNotNull'], 'fields': '', 'values': False},
            'textMode': 'auto'
        }
    return config

def create_prometheus_dashboard():
    """Create Prometheus load test dashboard"""
    panels = [
        {
            'collapsed': False,
            'gridPos': {'h': 1, 'w': 24, 'x': 0, 'y': 0},
            'id': 1,
            'panels': [],
            'title': '📊 Test Overview',
            'type': 'row'
        },
        create_panel_base(2, 'Active VUs', {'h': 4, 'w': 4, 'x': 0, 'y': 1}, [{
            'expr': 'sum(k6_vus{testid=~"$testid"})',
            'legendFormat': 'VUs',
            'refId': 'A'
        }], 'stat', 'short'),
        create_panel_base(3, 'Request Rate', {'h': 4, 'w': 4, 'x': 4, 'y': 1}, [{
            'expr': 'sum(rate(k6_http_reqs_total{testid=~"$testid"}[1m]))',
            'legendFormat': 'RPS',
            'refId': 'A'
        }], 'stat', 'reqps'),
        create_panel_base(4, 'HTTP Error Rate', {'h': 4, 'w': 4, 'x': 8, 'y': 1}, [{
            'expr': 'sum(rate(k6_http_req_failed_total{testid=~"$testid"}[5m])) / sum(rate(k6_http_reqs_total{testid=~"$testid"}[5m])) or vector(0)',
            'legendFormat': 'Error Rate',
            'refId': 'A'
        }], 'stat', 'percentunit', 2),
        create_panel_base(5, 'HTTP P95 Latency', {'h': 4, 'w': 4, 'x': 12, 'y': 1}, [{
            'expr': 'histogram_quantile(0.95, sum by (le) (rate(k6_http_req_duration_bucket{testid=~"$testid"}[5m]))) * 1000',
            'legendFormat': 'P95',
            'refId': 'A'
        }], 'stat', 'ms', 0),
        create_panel_base(6, 'Query Success Rate', {'h': 4, 'w': 4, 'x': 16, 'y': 1}, [{
            'expr': 'sum(rate(prometheus_query_success{testid=~"$testid"}[5m]))',
            'legendFormat': 'Success Rate',
            'refId': 'A'
        }], 'stat', 'percentunit', 2),
        create_panel_base(7, 'Total Queries', {'h': 4, 'w': 4, 'x': 20, 'y': 1}, [{
            'expr': 'sum(increase(prometheus_queries_total{testid=~"$testid"}[5m]))',
            'legendFormat': 'Total',
            'refId': 'A'
        }], 'stat', 'short'),
        {
            'collapsed': False,
            'gridPos': {'h': 1, 'w': 24, 'x': 0, 'y': 5},
            'id': 10,
            'panels': [],
            'title': '📈 Prometheus Query Performance',
            'type': 'row'
        },
        create_panel_base(11, 'Query Rate by Type', {'h': 8, 'w': 12, 'x': 0, 'y': 6}, [{
            'expr': 'sum by (query_type) (rate(prometheus_queries_total{testid=~"$testid"}[1m]))',
            'legendFormat': '{{query_type}}',
            'refId': 'A'
        }], 'timeseries', 'reqps'),
        create_panel_base(12, 'Query Duration Percentiles', {'h': 8, 'w': 12, 'x': 12, 'y': 6}, [
            {'expr': 'histogram_quantile(0.50, sum by (le) (rate(prometheus_query_duration_ms_bucket{testid=~"$testid"}[1m])))', 'legendFormat': 'P50', 'refId': 'A'},
            {'expr': 'histogram_quantile(0.90, sum by (le) (rate(prometheus_query_duration_ms_bucket{testid=~"$testid"}[1m])))', 'legendFormat': 'P90', 'refId': 'B'},
            {'expr': 'histogram_quantile(0.95, sum by (le) (rate(prometheus_query_duration_ms_bucket{testid=~"$testid"}[1m])))', 'legendFormat': 'P95', 'refId': 'C'},
            {'expr': 'histogram_quantile(0.99, sum by (le) (rate(prometheus_query_duration_ms_bucket{testid=~"$testid"}[1m])))', 'legendFormat': 'P99', 'refId': 'D'},
        ], 'timeseries', 'ms'),
    ]
    
    return {
        'annotations': {'list': []},
        'editable': True,
        'fiscalYearStartMonth': 0,
        'graphTooltip': 1,
        'id': None,
        'links': [],
        'liveNow': False,
        'panels': panels,
        'refresh': '10s',
        'schemaVersion': 39,
        'style': 'dark',
        'tags': ['k6', 'load-testing', 'prometheus'],
        'templating': {
            'list': [
                {
                    'current': {'selected': False, 'text': 'prometheus', 'value': 'prometheus'},
                    'hide': 0,
                    'includeAll': False,
                    'label': 'Datasource',
                    'multi': False,
                    'name': 'datasource',
                    'options': [],
                    'query': 'prometheus',
                    'refresh': 1,
                    'regex': '',
                    'skipUrlSync': False,
                    'type': 'datasource'
                },
                {
                    'allValue': '.*',
                    'current': {'selected': True, 'text': 'All', 'value': '$__all'},
                    'datasource': {'type': 'prometheus', 'uid': 'prometheus'},
                    'definition': 'label_values(k6_http_reqs_total, testid)',
                    'hide': 0,
                    'includeAll': True,
                    'label': 'Test ID',
                    'multi': True,
                    'name': 'testid',
                    'options': [],
                    'query': {'query': 'label_values(k6_http_reqs_total, testid)', 'refId': 'StandardVariableQuery'},
                    'refresh': 2,
                    'regex': '',
                    'skipUrlSync': False,
                    'sort': 2,
                    'type': 'query'
                }
            ]
        },
        'time': {'from': 'now-30m', 'to': 'now'},
        'timepicker': {'refresh_intervals': ['5s', '10s', '30s', '1m', '5m', '15m', '30m', '1h']},
        'timezone': 'browser',
        'title': 'K6 Prometheus Load Testing',
        'uid': 'k6-prometheus-load',
        'version': 1,
        'weekStart': ''
    }

def create_tempo_dashboard():
    """Create Tempo load test dashboard"""
    panels = [
        {
            'collapsed': False,
            'gridPos': {'h': 1, 'w': 24, 'x': 0, 'y': 0},
            'id': 1,
            'panels': [],
            'title': '🔍 Test Overview',
            'type': 'row'
        },
        create_panel_base(2, 'Active VUs', {'h': 4, 'w': 4, 'x': 0, 'y': 1}, [{
            'expr': 'sum(k6_vus{testid=~"$testid"})',
            'legendFormat': 'VUs',
            'refId': 'A'
        }], 'stat', 'short'),
        create_panel_base(3, 'Request Rate', {'h': 4, 'w': 4, 'x': 4, 'y': 1}, [{
            'expr': 'sum(rate(k6_http_reqs_total{testid=~"$testid"}[1m]))',
            'legendFormat': 'RPS',
            'refId': 'A'
        }], 'stat', 'reqps'),
        create_panel_base(4, 'HTTP Error Rate', {'h': 4, 'w': 4, 'x': 8, 'y': 1}, [{
            'expr': 'sum(rate(k6_http_req_failed_total{testid=~"$testid"}[5m])) / sum(rate(k6_http_reqs_total{testid=~"$testid"}[5m])) or vector(0)',
            'legendFormat': 'Error Rate',
            'refId': 'A'
        }], 'stat', 'percentunit', 2),
        create_panel_base(5, 'Ingest Success Rate', {'h': 4, 'w': 4, 'x': 12, 'y': 1}, [{
            'expr': 'sum(rate(tempo_ingest_success{testid=~"$testid"}[5m]))',
            'legendFormat': 'Success Rate',
            'refId': 'A'
        }], 'stat', 'percentunit', 2),
        create_panel_base(6, 'Query Success Rate', {'h': 4, 'w': 4, 'x': 16, 'y': 1}, [{
            'expr': 'sum(rate(tempo_query_success{testid=~"$testid"}[5m]))',
            'legendFormat': 'Success Rate',
            'refId': 'A'
        }], 'stat', 'percentunit', 2),
        create_panel_base(7, 'Total Ingestions', {'h': 4, 'w': 4, 'x': 20, 'y': 1}, [{
            'expr': 'sum(increase(tempo_ingests_total{testid=~"$testid"}[5m]))',
            'legendFormat': 'Total',
            'refId': 'A'
        }], 'stat', 'short'),
        {
            'collapsed': False,
            'gridPos': {'h': 1, 'w': 24, 'x': 0, 'y': 5},
            'id': 10,
            'panels': [],
            'title': '📊 Trace Ingestion & Query Performance',
            'type': 'row'
        },
        create_panel_base(11, 'Ingest Rate', {'h': 8, 'w': 12, 'x': 0, 'y': 6}, [{
            'expr': 'sum(rate(tempo_ingests_total{testid=~"$testid"}[1m]))',
            'legendFormat': 'Ingests/sec',
            'refId': 'A'
        }], 'timeseries', 'ops'),
        create_panel_base(12, 'Ingest & Query Duration', {'h': 8, 'w': 12, 'x': 12, 'y': 6}, [
            {'expr': 'histogram_quantile(0.95, sum by (le) (rate(tempo_ingest_duration_ms_bucket{testid=~"$testid"}[1m])))', 'legendFormat': 'Ingest P95', 'refId': 'A'},
            {'expr': 'histogram_quantile(0.95, sum by (le) (rate(tempo_query_duration_ms_bucket{testid=~"$testid"}[1m])))', 'legendFormat': 'Query P95', 'refId': 'B'},
        ], 'timeseries', 'ms'),
    ]
    
    return {
        'annotations': {'list': []},
        'editable': True,
        'fiscalYearStartMonth': 0,
        'graphTooltip': 1,
        'id': None,
        'links': [],
        'liveNow': False,
        'panels': panels,
        'refresh': '10s',
        'schemaVersion': 39,
        'style': 'dark',
        'tags': ['k6', 'load-testing', 'tempo'],
        'templating': {
            'list': [
                {
                    'current': {'selected': False, 'text': 'prometheus', 'value': 'prometheus'},
                    'hide': 0,
                    'includeAll': False,
                    'label': 'Datasource',
                    'multi': False,
                    'name': 'datasource',
                    'options': [],
                    'query': 'prometheus',
                    'refresh': 1,
                    'regex': '',
                    'skipUrlSync': False,
                    'type': 'datasource'
                },
                {
                    'allValue': '.*',
                    'current': {'selected': True, 'text': 'All', 'value': '$__all'},
                    'datasource': {'type': 'prometheus', 'uid': 'prometheus'},
                    'definition': 'label_values(k6_http_reqs_total, testid)',
                    'hide': 0,
                    'includeAll': True,
                    'label': 'Test ID',
                    'multi': True,
                    'name': 'testid',
                    'options': [],
                    'query': {'query': 'label_values(k6_http_reqs_total, testid)', 'refId': 'StandardVariableQuery'},
                    'refresh': 2,
                    'regex': '',
                    'skipUrlSync': False,
                    'sort': 2,
                    'type': 'query'
                }
            ]
        },
        'time': {'from': 'now-30m', 'to': 'now'},
        'timepicker': {'refresh_intervals': ['5s', '10s', '30s', '1m', '5m', '15m', '30m', '1h']},
        'timezone': 'browser',
        'title': 'K6 Tempo Load Testing',
        'uid': 'k6-tempo-load',
        'version': 1,
        'weekStart': ''
    }

def create_loki_dashboard():
    """Create Loki load test dashboard"""
    panels = [
        {
            'collapsed': False,
            'gridPos': {'h': 1, 'w': 24, 'x': 0, 'y': 0},
            'id': 1,
            'panels': [],
            'title': '📝 Test Overview',
            'type': 'row'
        },
        create_panel_base(2, 'Active VUs', {'h': 4, 'w': 4, 'x': 0, 'y': 1}, [{
            'expr': 'sum(k6_vus{testid=~"$testid"})',
            'legendFormat': 'VUs',
            'refId': 'A'
        }], 'stat', 'short'),
        create_panel_base(3, 'Request Rate', {'h': 4, 'w': 4, 'x': 4, 'y': 1}, [{
            'expr': 'sum(rate(k6_http_reqs_total{testid=~"$testid"}[1m]))',
            'legendFormat': 'RPS',
            'refId': 'A'
        }], 'stat', 'reqps'),
        create_panel_base(4, 'HTTP Error Rate', {'h': 4, 'w': 4, 'x': 8, 'y': 1}, [{
            'expr': 'sum(rate(k6_http_req_failed_total{testid=~"$testid"}[5m])) / sum(rate(k6_http_reqs_total{testid=~"$testid"}[5m])) or vector(0)',
            'legendFormat': 'Error Rate',
            'refId': 'A'
        }], 'stat', 'percentunit', 2),
        create_panel_base(5, 'Push Success Rate', {'h': 4, 'w': 4, 'x': 12, 'y': 1}, [{
            'expr': 'sum(rate(loki_push_success{testid=~"$testid"}[5m]))',
            'legendFormat': 'Success Rate',
            'refId': 'A'
        }], 'stat', 'percentunit', 2),
        create_panel_base(6, 'Query Success Rate', {'h': 4, 'w': 4, 'x': 16, 'y': 1}, [{
            'expr': 'sum(rate(loki_query_success{testid=~"$testid"}[5m]))',
            'legendFormat': 'Success Rate',
            'refId': 'A'
        }], 'stat', 'percentunit', 2),
        create_panel_base(7, 'Total Pushes', {'h': 4, 'w': 4, 'x': 20, 'y': 1}, [{
            'expr': 'sum(increase(loki_pushes_total{testid=~"$testid"}[5m]))',
            'legendFormat': 'Total',
            'refId': 'A'
        }], 'stat', 'short'),
        {
            'collapsed': False,
            'gridPos': {'h': 1, 'w': 24, 'x': 0, 'y': 5},
            'id': 10,
            'panels': [],
            'title': '📊 Log Push & Query Performance',
            'type': 'row'
        },
        create_panel_base(11, 'Log Push Rate', {'h': 8, 'w': 12, 'x': 0, 'y': 6}, [{
            'expr': 'sum(rate(loki_pushes_total{testid=~"$testid"}[1m]))',
            'legendFormat': 'Pushes/sec',
            'refId': 'A'
        }], 'timeseries', 'ops'),
        create_panel_base(12, 'Push & Query Duration', {'h': 8, 'w': 12, 'x': 12, 'y': 6}, [
            {'expr': 'histogram_quantile(0.95, sum by (le) (rate(loki_push_duration_ms_bucket{testid=~"$testid"}[1m])))', 'legendFormat': 'Push P95', 'refId': 'A'},
            {'expr': 'histogram_quantile(0.95, sum by (le) (rate(loki_query_duration_ms_bucket{testid=~"$testid"}[1m])))', 'legendFormat': 'Query P95', 'refId': 'B'},
        ], 'timeseries', 'ms'),
    ]
    
    return {
        'annotations': {'list': []},
        'editable': True,
        'fiscalYearStartMonth': 0,
        'graphTooltip': 1,
        'id': None,
        'links': [],
        'liveNow': False,
        'panels': panels,
        'refresh': '10s',
        'schemaVersion': 39,
        'style': 'dark',
        'tags': ['k6', 'load-testing', 'loki'],
        'templating': {
            'list': [
                {
                    'current': {'selected': False, 'text': 'prometheus', 'value': 'prometheus'},
                    'hide': 0,
                    'includeAll': False,
                    'label': 'Datasource',
                    'multi': False,
                    'name': 'datasource',
                    'options': [],
                    'query': 'prometheus',
                    'refresh': 1,
                    'regex': '',
                    'skipUrlSync': False,
                    'type': 'datasource'
                },
                {
                    'allValue': '.*',
                    'current': {'selected': True, 'text': 'All', 'value': '$__all'},
                    'datasource': {'type': 'prometheus', 'uid': 'prometheus'},
                    'definition': 'label_values(k6_http_reqs_total, testid)',
                    'hide': 0,
                    'includeAll': True,
                    'label': 'Test ID',
                    'multi': True,
                    'name': 'testid',
                    'options': [],
                    'query': {'query': 'label_values(k6_http_reqs_total, testid)', 'refId': 'StandardVariableQuery'},
                    'refresh': 2,
                    'regex': '',
                    'skipUrlSync': False,
                    'sort': 2,
                    'type': 'query'
                }
            ]
        },
        'time': {'from': 'now-30m', 'to': 'now'},
        'timepicker': {'refresh_intervals': ['5s', '10s', '30s', '1m', '5m', '15m', '30m', '1h']},
        'timezone': 'browser',
        'title': 'K6 Loki Load Testing',
        'uid': 'k6-loki-load',
        'version': 1,
        'weekStart': ''
    }

def create_linkerd_dashboard():
    """Create Linkerd load test dashboard"""
    panels = [
        {
            'collapsed': False,
            'gridPos': {'h': 1, 'w': 24, 'x': 0, 'y': 0},
            'id': 1,
            'panels': [],
            'title': '🔗 Test Overview',
            'type': 'row'
        },
        create_panel_base(2, 'Active VUs', {'h': 4, 'w': 4, 'x': 0, 'y': 1}, [{
            'expr': 'sum(k6_vus{testid=~"$testid"})',
            'legendFormat': 'VUs',
            'refId': 'A'
        }], 'stat', 'short'),
        create_panel_base(3, 'Request Rate', {'h': 4, 'w': 4, 'x': 4, 'y': 1}, [{
            'expr': 'sum(rate(k6_http_reqs_total{testid=~"$testid"}[1m]))',
            'legendFormat': 'RPS',
            'refId': 'A'
        }], 'stat', 'reqps'),
        create_panel_base(4, 'HTTP Error Rate', {'h': 4, 'w': 4, 'x': 8, 'y': 1}, [{
            'expr': 'sum(rate(k6_http_req_failed_total{testid=~"$testid"}[5m])) / sum(rate(k6_http_reqs_total{testid=~"$testid"}[5m])) or vector(0)',
            'legendFormat': 'Error Rate',
            'refId': 'A'
        }], 'stat', 'percentunit', 2),
        create_panel_base(5, 'API Success Rate', {'h': 4, 'w': 4, 'x': 12, 'y': 1}, [{
            'expr': 'sum(rate(linkerd_api_success{testid=~"$testid"}[5m]))',
            'legendFormat': 'Success Rate',
            'refId': 'A'
        }], 'stat', 'percentunit', 2),
        create_panel_base(6, 'Total API Requests', {'h': 4, 'w': 4, 'x': 16, 'y': 1}, [{
            'expr': 'sum(increase(linkerd_api_requests_total{testid=~"$testid"}[5m]))',
            'legendFormat': 'Total',
            'refId': 'A'
        }], 'stat', 'short'),
        create_panel_base(7, 'API P95 Latency', {'h': 4, 'w': 4, 'x': 20, 'y': 1}, [{
            'expr': 'histogram_quantile(0.95, sum by (le) (rate(linkerd_api_duration_ms_bucket{testid=~"$testid"}[5m])))',
            'legendFormat': 'P95',
            'refId': 'A'
        }], 'stat', 'ms', 0),
        {
            'collapsed': False,
            'gridPos': {'h': 1, 'w': 24, 'x': 0, 'y': 5},
            'id': 10,
            'panels': [],
            'title': '📊 Linkerd Metrics API Performance',
            'type': 'row'
        },
        create_panel_base(11, 'API Request Rate by Endpoint', {'h': 8, 'w': 12, 'x': 0, 'y': 6}, [{
            'expr': 'sum by (endpoint) (rate(linkerd_api_requests_total{testid=~"$testid"}[1m]))',
            'legendFormat': '{{endpoint}}',
            'refId': 'A'
        }], 'timeseries', 'reqps'),
        create_panel_base(12, 'API Duration Percentiles', {'h': 8, 'w': 12, 'x': 12, 'y': 6}, [
            {'expr': 'histogram_quantile(0.50, sum by (le) (rate(linkerd_api_duration_ms_bucket{testid=~"$testid"}[1m])))', 'legendFormat': 'P50', 'refId': 'A'},
            {'expr': 'histogram_quantile(0.90, sum by (le) (rate(linkerd_api_duration_ms_bucket{testid=~"$testid"}[1m])))', 'legendFormat': 'P90', 'refId': 'B'},
            {'expr': 'histogram_quantile(0.95, sum by (le) (rate(linkerd_api_duration_ms_bucket{testid=~"$testid"}[1m])))', 'legendFormat': 'P95', 'refId': 'C'},
            {'expr': 'histogram_quantile(0.99, sum by (le) (rate(linkerd_api_duration_ms_bucket{testid=~"$testid"}[1m])))', 'legendFormat': 'P99', 'refId': 'D'},
        ], 'timeseries', 'ms'),
    ]
    
    return {
        'annotations': {'list': []},
        'editable': True,
        'fiscalYearStartMonth': 0,
        'graphTooltip': 1,
        'id': None,
        'links': [],
        'liveNow': False,
        'panels': panels,
        'refresh': '10s',
        'schemaVersion': 39,
        'style': 'dark',
        'tags': ['k6', 'load-testing', 'linkerd'],
        'templating': {
            'list': [
                {
                    'current': {'selected': False, 'text': 'prometheus', 'value': 'prometheus'},
                    'hide': 0,
                    'includeAll': False,
                    'label': 'Datasource',
                    'multi': False,
                    'name': 'datasource',
                    'options': [],
                    'query': 'prometheus',
                    'refresh': 1,
                    'regex': '',
                    'skipUrlSync': False,
                    'type': 'datasource'
                },
                {
                    'allValue': '.*',
                    'current': {'selected': True, 'text': 'All', 'value': '$__all'},
                    'datasource': {'type': 'prometheus', 'uid': 'prometheus'},
                    'definition': 'label_values(k6_http_reqs_total, testid)',
                    'hide': 0,
                    'includeAll': True,
                    'label': 'Test ID',
                    'multi': True,
                    'name': 'testid',
                    'options': [],
                    'query': {'query': 'label_values(k6_http_reqs_total, testid)', 'refId': 'StandardVariableQuery'},
                    'refresh': 2,
                    'regex': '',
                    'skipUrlSync': False,
                    'sort': 2,
                    'type': 'query'
                }
            ]
        },
        'time': {'from': 'now-30m', 'to': 'now'},
        'timepicker': {'refresh_intervals': ['5s', '10s', '30s', '1m', '5m', '15m', '30m', '1h']},
        'timezone': 'browser',
        'title': 'K6 Linkerd Load Testing',
        'uid': 'k6-linkerd-load',
        'version': 1,
        'weekStart': ''
    }

if __name__ == '__main__':
    dashboards = {
        'k6-prometheus-load-dashboard.json': create_prometheus_dashboard(),
        'k6-tempo-load-dashboard.json': create_tempo_dashboard(),
        'k6-loki-load-dashboard.json': create_loki_dashboard(),
        'k6-linkerd-load-dashboard.json': create_linkerd_dashboard(),
    }
    
    for filename, dashboard in dashboards.items():
        with open(filename, 'w') as f:
            json.dump(dashboard, f, indent=2)
        print(f'Created {filename}')
