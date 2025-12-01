# ArthurVardevanyan-com

```bash
PROJECT_ID="$(vault kv get -field=project_id secret/gcp/project/av)"

cat << EOF > backend.conf
bucket = "tf-state-${PROJECT_ID}"
prefix = "terraform/state"
EOF

tofu init -backend-config=backend.conf
tofu plan
tofu apply
```
