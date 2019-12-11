#URL=localhost:8000
URL=myapp.local
echo "POST /configs"
curl -XPOST "$URL"/configs -d '{"name": "pod-1000","metadata": {"monitoring": {"enabled": "true"},"limits": {"cpu": {"enabled": "false","value": "900m"}}}}'
echo "\nGET /configs/pod-1000"
curl -XGET "$URL"/configs/pod-1000
echo "\nPUT /configs/pod-1000"
curl -XPUT "$URL"/configs/pod-1000 -d '{"name": "pod-1000","metadata": {"monitoring": {"enabled": "false"},"limits": {"cpu": {"enabled": "false","value": "800m"}}}}'
echo "\nGET /configs/pod-1000"
curl -XGET "$URL"/configs/pod-1000
echo "\nDELETE /configs/pod-1000"
curl -XDELETE "$URL"/configs/pod-1000
echo "\nGET /configs"
curl -XGET "$URL"/configs # | jq '.[].config.Data | select(.name=="pod-1000")'
