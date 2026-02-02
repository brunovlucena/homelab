import json
import sys

# Ler do stdin
data = json.load(sys.stdin)

target_ids = ['BVL-21', 'BVL-22', 'BVL-23', 'BVL-24', 'BVL-25', 'BVL-28']

for issue in data:
    identifier = issue.get('identifier', '')
    if identifier in target_ids:
        title = issue.get('title', '')
        # Verificar se tem duplicação
        if identifier in title and '-' in title.replace(identifier, ''):
            print(json.dumps({
                'id': issue['id'],
                'identifier': identifier,
                'title': title,
                'project': issue.get('project', '')
            }))
