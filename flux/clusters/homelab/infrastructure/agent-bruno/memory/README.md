memory/
├── incidents/          # Past incidents with resolutions
├── runbooks/           # Procedural knowledge
├── metrics_patterns/   # Anomaly patterns
└── code_snippets/      # kubectl/troubleshooting commands

# Use BM25 for:
- Error code lookups
- Service name filtering
- Command/log matching

# Use BGE-M3 for:
- Symptom descriptions
- Root cause matching
- Similar incident retrieval

# Combine weights based on query analysis:
if contains_error_code(query):
    w_bm25 = 0.7  # Favor exact match
    w_bge = 0.3
else:
    w_bm25 = 0.3
    w_bge = 0.7   # Favor semantic