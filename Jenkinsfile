@Library('github.com/fabric8io/fabric8-pipeline-library@master')
def dummy
goTemplate{
  dockerNode{
      goMake{
        githubOrganisation = 'kubernetes'
        dockerOrganisation = 'fabric8'
        project = 'kompose'
        makeCommand = "PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/usr/local/go/bin:/usr/local/glide:/usr/local/:/go/bin:/home/jenkins/go/bin \
                       && bash script/test/cmd/fix_detached_head.sh && make test"
      }
  }
}
