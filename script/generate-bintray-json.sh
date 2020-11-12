#!/bin/sh

DATE=`date --iso-8601=date`
TIME=`date --iso-8601=seconds`

# GITHUB_PULL_REQUEST MUST be set by your own hands.
# See: https://github.com/actions/checkout/issues/58

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
        "vcs_tag": "${GITHUB_SHA}",
        "attributes": [{"name": "GITHUB_RUN_NUMBER", "values" : ["${GITHUB_RUN_NUMBER}"], "type": "string"},
                       {"name": "GITHUB_RUN_ID", "values" : ["${GITHUB_RUN_ID}"], "type": "string"},
                       {"name": "GITHUB_SHA", "values" : ["${GITHUB_SHA}"], "type": "string"},
                       {"name": "GITHUB_REF", "values" : ["${GITHUB_REF}"], "type": "string"},
                       {"name": "GITHUB_PULL_REQUEST", "values" : ["${GITHUB_PULL_REQUEST}"], "type": "string"},
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
