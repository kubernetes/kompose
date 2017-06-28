#!/bin/bash

# Copyright 2017 The Kubernetes Authors All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Check if there are nested vendor dirs inside Kompose vendor.
# All dependencies should be flattened and there shouldn't be vendor in inside vendor.  

function check_nested_vendor() {
    echo "Checking for nested vendor dirs"

    # count all vendor directories inside Kompose vendor
    NO_NESTED_VENDORS=$(find vendor/ -type d | sed 's/^[^/]*.//g' | grep -E "vendor$" | grep -v _vendor | wc -l)
    
    if [ $NO_NESTED_VENDORS -ne 0 ]; then
        echo "ERROR"
        echo "  There are $NO_NESTED_VENDORS nested vendors in Kompose vendor directory"
        echo "  Please run 'glide update --strip-vendor'"
        return 1
    else
        echo "OK"
        return 0
    fi
}


# Check if Kompose vendor directory was cleaned by glide-vc  
function check_glide-vc() {
    echo "Checking if vendor was cleaned using glide-vc."

    # dry run glide-vc and count how many could be deleted. 
    NO_DELETED_FILES=$(glide-vc --only-code --no-tests --dryrun | wc -l)
    
    if [ $NO_DELETED_FILES -ne 0 ]; then
        echo "ERROR"
        echo "  There are $NO_DELETED_FILES files that can be deleted by glide-vc."
        echo "  Please run 'glide-vc --only-code --no-tests'" 
        return 1
    else
        echo "OK"
        return 0
    fi
}


# Run both checks and exit report fail exit code if one of them failed.
check_nested_vendor
VENDOR_CHECK=$?

check_glide-vc
VC_CHECK=$? 

if [ $VENDOR_CHECK -eq 0 ] && [ $VC_CHECK -eq 0 ]; then
    exit 0
else
    exit 1
fi

