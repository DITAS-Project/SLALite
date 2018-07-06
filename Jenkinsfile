pipeline {
    agent none
    stages {
        stage('Build and test') {
            agent {
                dockerfile {
                    filename 'Dockerfile.build'
                }
            }
            steps {
                sh "export GOPATH=${WORKSPACE}"
                sh "ln -sf ${WORKSPACE} ${WORKSPACE}/src"
                sh "cd /go/src/SLALite"
                sh "rm -rf vendor"
                sh "dep ensure"
                sh "CGO_ENABLED=0 GOOS=linux go build -a -o SLALite"
		        // Test y build en go?
		        sh "go test ./..."	
            }
        }
        stage('Image creation') {
            agent any
            options {
                skipDefaultCheckout true
            }
            steps {
                // The Dockerfile.artifact copies the code into the image and run the jar generation.
                echo 'Creating the image...'

                // This will search for a Dockerfile.artifact in the working directory and build the image to the local repository
                sh "docker build -t \"ditas/slalite\" -f Dockerfile.artifact ."
                echo "Done"
		    
                // Get the password from a file. This reads the file from the host, not the container. Slaves already have the password in there.
                echo 'Retrieving Docker Hub password from /opt/ditas-docker-hub.passwd...'
                script {
                    password = readFile '/opt/ditas-docker-hub.passwd'
                }
                echo "Done"

                echo 'Login to Docker Hub as ditasgeneric...'
                sh "docker login -u ditasgeneric -p ${password}"
                echo "Done"

                echo "Pushing the image ditas/slalite:latest..."
                sh "docker push ditas/slalite:latest"
                echo "Done "
            }
        }
        stage('Image deploy') {
            agent any
            options {
                // Don't need to checkout Git again
                skipDefaultCheckout true
            }
            steps {
                // TODO: Uncomment this when the previous stages run correctly
                // TODO: Remember to edit 'deploy-staging.sh' and configure the ports
                // Deploy to Staging environment calling the deployment script
                sh './jenkins/deploy/deploy-staging.sh'
            }
        }
    }
}
