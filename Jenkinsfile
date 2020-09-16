node {
  def tag="docker.isc.ru.nl/rdr/tool/rdr-emailer:latest"

  stage('Checkout') {
    checkout scm
  }

  stage('Build') {
    sh "docker build -t ${tag} --force-rm ."
  }

  stage('Push iRODS') {
    sh "docker push ${tag}"
  }
}