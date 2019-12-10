#URL=localhost:8000
URL=myapp.local
curl -XGET "$URL"/configs | jq '.[].config.Data | select(.name=="pod-1000") '
echo "POST"
curl -XPOST "$URL"/configs -d '{"name": "pod-1000","metadata": {"monitoring": {"enabled": "true"},"limits": {"cpu": {"enabled": "false","value": "900m"}}}}'
curl -XGET "$URL"/configs | jq '.[].config.Data | select(.name=="pod-1000") '
echo "PUT"
curl -XPUT "$URL"/configs/pod-1000 -d '{"name": "pod-1000","metadata": {"monitoring": {"enabled": "false"},"limits": {"cpu": {"enabled": "false","value": "900m"}}}}'
curl -XGET "$URL"/configs | jq '.[].config.Data | select(.name=="pod-1000") '
curl -XGET "$URL"/configs/pod-1000
echo "DELETE"
curl -XDELETE "$URL"/configs/pod-1000
curl -XGET "$URL"/configs | jq '.[].config.Data | select(.name=="pod-1000") '
