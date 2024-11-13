#!groovy

def randomString() {
    char[] chars = "abcdefghijklmnopqrstuvwxyz".toCharArray();
    StringBuilder sb = new StringBuilder(6);
    Random random = new Random();
    for (int i = 0; i < 6; i++) {
        char c = chars[random.nextInt(chars.length)];
        sb.append(c);
    }
    return sb.toString();
}

node('aws&&docker') {
try {
    def major = '1'
    def commitSha = sh(script: "git rev-parse --short HEAD", returnStdout: true).trim()
    def commitCount = sh(script: "git rev-list --count HEAD", returnStdout: true).trim()
    def version = "${major}.${commitCount}.${commitSha.take(7)}"

    currentBuild.displayName = "#${env.BUILD_NUMBER}: ${version}"
    
    def arch_arm = 'arm64'
    def runtime_go = 'provided.al2'

    def artifact = "${git.repo}-lambda.zip"
    def versionedArtifact = "${git.repo}-lambda-${version}.zip"
    def buildImage
    def url

    stage('Build Docker image') {
        buildImage = docker.build('build-image', '--file Dockerfile.aarch .')
        sh "mkdir -p artifact_publish"
    }

    cwd = pwd() 
    buildImage.inside {
        stage('Build') {
            sh "id -u"
            sh "go version"
            sh "ARTIFACT=${artifact} ./build.sh -v -l all"

            dir('publish') {
                // Add version file to zip
                sh "echo ${version} > version.txt"
                sh "zip -u bin/${artifact} version.txt"
                sh "mv * ../artifact_publish/"
            }
        }

    withCredentials([[$class: 'AmazonWebServicesCredentialsBinding',
            credentialsId    : '<YOUR_CREDENTIAL_ID>',
            accessKeyVariable: 'AWS_ACCESS_KEY_ID',
            secretKeyVariable: 'AWS_SECRET_ACCESS_KEY']]) {
        stage('Upload to S3 buckets') {
            buildImage.inside {
                dir('artifact_publish') {
                    // Upload to S3
                    sh "aws s3 cp bin/${artifact} s3://git-credential-service-artifacts/lambdas/arm64/provided.al2/git-credential-service/latest/${artifact}"
                    sh "aws s3 cp ${cwd}/deploy/openapi-githubcredential.yml s3://git-credential-service-artifacts/api/openapi-githubcredential.yml"
                }
            }
        }

        stage('Deploy Service') {
            logLevel = "INFO"
            if (env.BRANCH_NAME == 'main') {
                logLevel = "ERROR"
            }
            user = "${currentBuild.getBuildCauses()[0].userId}"
            if (user == "") {
                user = "unknown"
            }
            ranPrefix = randomString()
            sh "sed -i 's/RESOURCENAMEPREFIX/${ranPrefix}/g' deploy/openapi-githubcredential.yml"
            sh "aws cloudformation deploy \
                --stack-name ${ranPrefix}-git-credential-service \
                --template-file deploy/template.yml \
                --capabilities CAPABILITY_NAMED_IAM \
                --parameter-overrides \
                    ResourceNamePrefix=${ranPrefix} \
                    ResourceBucket=git-credential-service-artifacts \
                    LogLevel=${logLevel} \
                --region us-east-1 \
                --tags Owner=${user} Service='GitHub Credential API'"
        }
    }
} finally {
    deleteDir()
}}