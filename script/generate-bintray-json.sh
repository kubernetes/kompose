#!/bin/sh

DATE=`date --iso-8601=date`
TIME=`date --iso-8601=seconds`

cat > "./.bintray.json" <<EOF
{
    "package": {
        "name": "kompose",
        "repo": "kompose",
        "subject": "kompose",
        "desc": "Go from Docker Compose to Kubernetes",
        "website_url": "https://github.com/kubernetes/kompose",
        "issue_tracker_url": "https://github.com/kubernetes/komposeissues",
        "vcs_url": "https://github.com/kubernetes/kompose",
        "licenses": ["Apache-2.0"],
        "public_download_numbers": false,
        "public_stats": false
    },

    "version": {
        "name": "latest",
        "desc": "Kompose build from master branch",
        "released": "${DATE}",
        "vcs_tag": "${TRAVIS_COMMIT}",
        "attributes": [{"name": "TRAVIS_JOB_NUMBER", "values" : ["${TRAVIS_JOB_NUMBER}"], "type": "string"},
                       {"name": "TRAVIS_JOB_ID", "values" : ["${TRAVIS_JOB_ID}"], "type": "string"},
                       {"name": "TRAVIS_COMMIT", "values" : ["${TRAVIS_COMMIT}"], "type": "string"},
                       {"name": "TRAVIS_BRANCH", "values" : ["${TRAVIS_BRANCH}"], "type": "string"},
                       {"name": "TRAVIS_PULL_REQUEST", "values" : ["${TRAVIS_PULL_REQUEST}"], "type": "string"},
                       {"name": "date", "values" : ["${TIME}"], "type": "date"}],
        "gpgSign": false
    },

    "files":
        [
            {"includePattern": "bin/(.*)",
             "uploadPattern": "./latest/\$1", 
             "matrixParams": {"override": 1 }
            }
        ],
    "publish": true
}
EOF