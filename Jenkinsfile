pipeline {
    agent none
    stages {
        stage('Build') {
            agent {
                dockerfile {
                    filename 'Dockerfile.build'
                }
            }
            steps {
                sh "go get -d -v ./..."
                sh "CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo"
            }
        }
    }
}