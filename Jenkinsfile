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
                // Build
		        sh "echo skipping"
                //sh "jenkins/build.sh ${WORKSPACE}"
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
                sh "docker build -t \"ditas/slalite:staging\" -f Dockerfile ."
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

                echo "Pushing the image ditas/slalite:staging..."
                sh "docker push ditas/slalite:staging"
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
