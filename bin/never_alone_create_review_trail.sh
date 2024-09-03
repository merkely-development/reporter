#!/usr/bin/env bash
set -Eeu

SCRIPT_NAME="never_alone_create_review_trail.sh"
MAIN_BRANCH=""
RELEASE_FLOW=""
PROPOSED_COMMIT=""
TRAIL_NAME=""


function print_help
{
    cat <<EOF
Use: $SCRIPT_NAME [options]

Script to get commit and pull request info for all commits to main/master branch
and report them to Kosli

Options are:
  -h               Print this help menu
  -m <branch>      Name of main/master branch. Required
  -f <flow>        Name of kosli flow to report commit and pull request info to. Required
  -c <commit-sha>  Commit sha for release we are building now. Required
  -t <trail-name>  Name of the trail that the reviews shall be reported to. Required
EOF
}


function die
{
    echo "Error: $1" >&2
    exit 1
}


function repo_root
{
  git rev-parse --show-toplevel
}


function check_arguments
{
    while getopts "hc:m:f:t:" opt; do
        case $opt in
            h)
                print_help
                exit 1
                ;;
            c)
                PROPOSED_COMMIT=${OPTARG}
                ;;
            m)
                MAIN_BRANCH=${OPTARG}
                ;;
            f)
                RELEASE_FLOW=${OPTARG}
                ;;
            t)
                TRAIL_NAME=${OPTARG}
                ;;
            \?)
                echo "Invalid option: -$OPTARG" >&2
                exit 1
                ;;
        esac
    done

    if [ -z "${PROPOSED_COMMIT}" ]; then
        die "option -c <commit-sha> is required"
    fi
    if [ -z "${MAIN_BRANCH}" ]; then
        die "option -m <branch> is required"
    fi
    if [ -z "${RELEASE_FLOW}" ]; then
        die "option -f <commit-prs-filename> is required"
    fi
    if [ -z "${TRAIL_NAME}" ]; then
        die "option -t <trail-name> is required"
    fi
}

function begin_trail_with_template
{
    local release_flow=$1; shift
    local trail_name=$1; shift
    local commits=("$@")
    local trail_template_file_name="review_trail.yaml"
    {
    cat <<EOF
version: 1
trail:
  attestations:
EOF

    for commit in "${commits[@]}"; do
        echo "    - name: sha_${commit}"
        echo "      type: generic"
    done
    } > ${trail_template_file_name}

    kosli begin trail ${trail_name} \
        --flow=${release_flow} \
        --description="$(git log -1 --pretty='%aN - %s')" \
        --template-file=${trail_template_file_name}
}

function get_never_alone_attestation_in_trail
{
    local commit_flow=$1; shift
    local trail_name=$1; shift
    local slot_name="never-alone-data"
    local -r curl_output_file=$(mktemp)


    http_code=$(curl -X 'GET' \
        --user ${KOSLI_API_TOKEN}:unused \
        "${KOSLI_HOST}/api/v2/attestations/${KOSLI_ORG}/${commit_flow}/trail/${trail_name}/${slot_name}" \
        -H 'accept: application/json' \
        --output "${curl_output_file}" \
        --write-out "%{http_code}" \
        --silent)

    if [[ ${http_code} -lt 200 || ${http_code} -gt 299 ]] ; then
        >&2 cat "${curl_output_file}"
        rm "${curl_output_file}"
        echo "[]"
        return
    fi

    cat "$curl_output_file"
    rm "${curl_output_file}"
}

function get_never_alone_compliance
{

    echo false
}

function attest_commit_trail_never_alone
{
    local release_flow=$1; shift
    local trail_name=$1; shift
    local -r commit=$1; shift
    # local link_to_attestation=https://app.kosli.com/${{inputs.kosli_org}}/flows/${CODE_REVIEW_FLOW}/trails/${{inputs.trail_name}}
    local link_to_attestation=http://localhost/${KOSLI_ORG}/flows/cli/trails/${commit}

    never_alone_data=$(get_never_alone_attestation_in_trail cli ${commit})
    if [ "${never_alone_data}" != "[]" ]; then

        compliant=$(get_never_alone_compliance "${never_alone_data}")

        kosli attest generic \
            --flow ${release_flow} \
            --trail ${trail_name} \
            --name="sha_${commit}" \
            --compliant=${compliant} \
            --external-url never-alone-data=${link_to_attestation}
    fi
}

function main
{
    check_arguments "$@"
    # base_commit: the commit of latest release
    # local -r base_commit=$($(repo_root)/bin/never_alone_get_commit_of_latest_release.sh)
    # base_commit="ad4500e73dcb6fb980bcc2b12f44f0750a4adfcc"
    base_commit="d9a332df12ec3883f48b0d79858be5ef9c2bed45"
    base_commit="4d6ccf339e627ea850071e859f93a34b53284512"

    # Use gh instead of git so we can keep the commit depth of 1. The order of the response for gh is reversed
    # so I do a tac at the end to get it the same order.
    commits=($(gh api repos/:owner/:repo/compare/${base_commit}...${PROPOSED_COMMIT} -q '.commits[].sha' | tac))

    begin_trail_with_template ${RELEASE_FLOW} ${TRAIL_NAME} "${commits[@]}"
    
    for commit in "${commits[@]}"; do
        set +e
        attest_commit_trail_never_alone ${RELEASE_FLOW} ${TRAIL_NAME} $commit
        set -e
    done
}

main "$@"
