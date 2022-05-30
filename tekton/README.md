# Tekton

```bash
tkn -n arthur-vardevanyan pipeline start arthur-vardevanyan -s pipeline-sa \
    --workspace=name=data,volumeClaimTemplateFile=tekton/base/pvc.yaml \
    --param="IMAGE=registry.arthurvardevanyan.com/apps/arthur-vardevanyan" \
    --param="git-url=https://git.arthurvardevanyan.com/ArthurVardevanyan/ArthurVardevanyan-com.git" \
    --param="git-name=ArthurVardevanyan/arthur-vardevanyan" \
    --param="git-commit=$(git log --format=oneline | cut -d ' ' -f 1 | head -n 1)"
    --showlog
```
