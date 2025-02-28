@Library(['dcpMonorepo', 'workflow-api-library']) _

ArrayList<Map> filesToReplace = []

JOB_ENABLED = false

pipeline {
    agent {label 'IDC-Builder'}
    when {
        expression { JOB_ENABLED }
    }
    parameters {
        string(name: "BRANCH_TO_DEPLOY", defaultValue: 'main', trim: true, description: "Branch or commit hash of frameworks.cloud.devcloud.services.idc repository that need to be deployed")
        choice(name: "IDC_ENV", choices: ["", "staging", "prod"], description: "Target IDC environment")
        choice(name: "LAYER", choices: ["", "all", "global", "regional", "az"])
    }
    stages {
        stage("Verify pipeline parameters") {
            steps{
                script {
                    if (!params.IDC_ENV || !params.LAYER) {
                        error("Parameters IDC_ENV and LAYER should not be empty")
                    }
                    if (params.LAYER == "all") {
                        env.TARGET_DIRECTORIES = "local/secrets/${IDC_ENV}/helm-values/idc-global-services,local/secrets/${IDC_ENV}/helm-values/idc-regional"
                        env.MAKE_COMMAND = "helmfile-generate-argocd-values"
                    } else {
                        env.TARGET_DIRECTORIES = "local/secrets/${IDC_ENV}/helm-values/idc-regional"
                        env.MAKE_COMMAND = "helmfile-generate-argocd-${params.LAYER}-values"
                    }
                    if (params.IDC_ENV == "prod") {
                        env.VAULT_ADDR="https://internal-placeholder.com"
                        env.VAULT_CREDS_ID = "vault-prod-approle"
                    } else if (params.IDC_ENV == "staging") {
                        env.VAULT_ADDR="https://internal-placeholder.com"
                        env.VAULT_CREDS_ID = "vault-staging-approle"
                    }
                    env.VAULT_PROXY_ADDR="http://internal-placeholder.com:912"
                }
            }
        }
        stage("Generage ArgoCD values") {
            steps {
                script{
                    dir("idc") {
                        git branch: params.BRANCH_TO_DEPLOY, changelog: false, poll: false, credentialsId: 'sys-idccicd-github-token', url: 'https://github.com/intel-innersource/frameworks.cloud.devcloud.services.idc.git'
                        env.VERSION = env.GIT_COMMIT.take(8)
                        withCredentials([usernamePassword(credentialsId: env.VAULT_CREDS_ID, usernameVariable: 'VAULT_APPROLE_ID', passwordVariable: 'VAULT_APPROLE_SECRET')]) {
                            sh """
                                mkdir -p local/secrets/${params.IDC_ENV}
                                vault write -field=token auth/approle/login role_id=$VAULT_APPROLE_ID secret_id=$VAULT_APPROLE_SECRET > local/secrets/${IDC_ENV}/VAULT_TOKEN
                                make ${env.MAKE_COMMAND}
                            """
                        }
                        env.TARGET_DIRECTORIES.split(',').each { targetDirectory ->
                            if (!fileExists(targetDirectory)) {
                                error("Files in target directory ${targetDirectory} have not been generated")
                            }
                            dir(targetDirectory) {
                                String generatedFiles = sh(script: 'find . -type f -print', returnStdout: true).trim()
                                generatedFiles.split('\n').each {
                                    Map<String,String> fileToReplace = [filename: it.substring(2), initialPath: targetDirectory, targetPath: "applications/${targetDirectory.split('/').last()}"]
                                    filesToReplace.add(fileToReplace)
                                }
                            }
                        }
                    }
                }
            }
        }
        stage("Prepare changes in ArgoCD repository") {
            steps {
                script {
                    env.TARGET_BRANCH_NAME = "deploy-${params.IDC_ENV}-${params.LAYER}-${env.VERSION}-${env.BUILD_ID}"
                    dir("idc-argocd"){
                        git branch: 'main', changelog: false, poll: false, credentialsId: 'sys-idccicd-github-token', url: 'https://github.com/intel-innersource/frameworks.cloud.devcloud.services.idc-argocd.git'
                        sh """
                            git checkout -b ${env.TARGET_BRANCH_NAME}
                        """
                    }
                    filesToReplace.each {
                        sh "mkdir -p `dirname idc-argocd/${it.targetPath}/${it.filename}` && cp -f idc/${it.initialPath}/${it.filename} idc-argocd/${it.targetPath}/${it.filename}"
                    }
                    dir("idc-argocd") {
                        withCredentials([gitUsernamePassword(credentialsId: 'sys-idccicd-github-token', gitToolName: 'Default')]) {
                            sh """
                                git add .
                                git commit -m 'Deploy version ${env.VERSION} in ${params.LAYER} ${params.IDC_ENV} environment'
                                git push --set-upstream origin ${env.TARGET_BRANCH_NAME}
                            """
                        }
                    }
                }
            }
        }
        stage("Create Pull Request") {
            steps {
                script {
                    withCredentials([usernamePassword(credentialsId: 'sys-idccicd-github-token', usernameVariable: 'GH_USERNAME', passwordVariable: 'GH_PASSWORD')]) {
                        sh """
                            curl -X POST -x http://internal-placeholder.com:912 \
                            -H "Accept: application/vnd.github+json" -H "Authorization: Bearer $GH_PASSWORD" -H "X-GitHub-Api-Version: 2022-11-28" \
                            https://api.github.com/repos/intel-innersource/frameworks.cloud.devcloud.services.idc-argocd/pulls \
                            -d '{"title":"Deploy version ${env.VERSION} in ${params.LAYER} ${params.IDC_ENV} environment","body":"Automatically created PR for deployment","head":"${env.TARGET_BRANCH_NAME}","base":"main"}'
                        """
                    }
                }
            }
        }
    }
    post {
        always {
            script {
                cleanWs()
                env.GIT_URL = "https://github.com/intel-innersource/frameworks.cloud.devcloud.services.idc"
                postWorkflowRun()
            }
        }
    }
}
