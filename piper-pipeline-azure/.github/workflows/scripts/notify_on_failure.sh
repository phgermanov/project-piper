#!/usr/bin/env bash

gh_token="$ENV_GITHUB_TOKEN"
smtp_user="$ENV_SMTP_USER"
smtp_password="$ENV_SMTP_PASSWORD"

response_output="resp.json"
url="$GITHUB_API_URL/repos/$GITHUB_REPOSITORY/actions/runs/$GITHUB_RUN_ID/jobs"
echo -e "curl -L -s -o $response_output -H \"Authorization: Bearer <token>\" -H \"Accept: application/vnd.github.raw+json\" $url \n"
status_code=$(curl --write-out %{http_code} \
                   -L -s -o "$response_output" \
                   -H "Authorization: Bearer ${gh_token}" \
                   -H "Accept: $accept_header" \
                   $url)
if [[ "$status_code" -ne 200 ]]; then
    echo "request to $url failed with code $status_code:"
    cat "$response_output"
    exit 1
fi

failed_job_obj=$(jq '.jobs[] | select( .conclusion as $c | ["failure", "cancelled"] | index($c) )' $response_output)
rm -f $response_output
if [ -z "${failed_job_obj}" ]; then
    echo "No failed jobs."
    exit 0
fi

job_status=$(jq -r .conclusion <<< "$failed_job_obj")
html_url=$(jq -r .html_url <<< "$failed_job_obj")

smtp_body="From: Piper on GHA <project-piper@sap.com>\n"
smtp_body+="To: DL_6287AE4DEC3CA802990E86E5@global.corp.sap\n"
smtp_body+="Importance: High\n"
smtp_body+="Subject: Workflow failure in $GITHUB_REPOSITORY\n\n"

smtp_body+="Workflow '$GITHUB_WORKFLOW' has a job with status '$job_status'.\n"
smtp_body+="Workflow link: $html_url"

curl --ssl-reqd \
     --url 'smtps://smtpauth.mail.net.sap:465' \
     --mail-from 'project-piper@sap.com' \
     --mail-rcpt 'DL_6287AE4DEC3CA802990E86E5@global.corp.sap' \
     --user "${smtp_user}:${smtp_password}" \
     -T <(echo -e "$smtp_body")
