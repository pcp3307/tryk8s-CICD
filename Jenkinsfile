node{
    def project = 'analog-ship-130206'
    def appName = 'k8stest'
    def tag = "v_${env.BUILD_NUMBER}"
    def img = "gcr.io/${project}/${appName}-${env.BRANCH_NAME}"
    def imgWithTag = "${img}:${tag}"
    
    def devNamespace = 'develop'
    def proNamespace = 'production'

    checkout scm

    stage '建立映像檔'
    sh("docker build -t ${imgWithTag} .")

    stage '放置映像檔'
    sh("gcloud docker -- push ${imgWithTag} ")

    stage '部署'
    // replace as new image
    sh("sed -i.bak 's#gcr.io/k8stest-image#${imgWithTag}#' ./k8s/deploy.yaml")
    switch (env.BRANCH_NAME) {
      case "master":
        // replace namespace settings
        sh("sed -i.bak 's#env: current#env: ${devNamespace}#' ./k8s/service.yaml")
        sh("sed -i.bak 's#env: current#env: ${devNamespace}#' ./k8s/deploy.yaml")
        sh("kubectl --namespace=${devNamespace} apply -f ./k8s/service.yaml")
        sh("kubectl --namespace=${devNamespace} apply -f ./k8s/deploy.yaml")
        break
    }
}
