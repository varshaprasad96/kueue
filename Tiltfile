load('ext://restart_process', 'docker_build_with_restart')

yaml = kustomize("config/components/crd")
k8s_yaml(yaml)

docker_build('us-central1-docker.pkg.dev/k8s-staging-images/kueue/kueue', '.',
    dockerfile='Dockerfile')

yaml = kustomize("config/components/manager")

k8s_yaml(kustomize("config/default"), allow_duplicates=True)