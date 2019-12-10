curl -XGET myapp.local/configs | jq '.[].config.Data | select(.name=="pod-1000") '
echo "POST"
curl -XPOST myapp.local/configs -d '{"name": "pod-1000","metadata": {"monitoring": {"enabled": "true"},"limits": {"cpu": {"enabled": "false","value": "900m"}}}}'
curl -XGET myapp.local/configs | jq '.[].config.Data | select(.name=="pod-1000") '
echo "PUT"
curl -XPUT myapp.local/configs/pod-1000 -d '{"name": "pod-1000","metadata": {"monitoring": {"enabled": "false"},"limits": {"cpu": {"enabled": "false","value": "900m"}}}}'
curl -XGET myapp.local/configs | jq '.[].config.Data | select(.name=="pod-1000") '
curl -XGET myapp.local/configs/pod-1000
echo "DELETE"
curl -XDELETE myapp.local/configs/pod-1000
curl -XGET myapp.local/configs | jq '.[].config.Data | select(.name=="pod-1000") '
