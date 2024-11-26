# Gitlab MergeRequest Reviewer

`Gitlab MergeRequest Reviewer` would review the file changes in the given merge request, and generate the summary.

## How to use

### CLI

```shell
## run command
make build && build/gitlab-mr-reviewer --project=${GITLAB_PROJECT_ID}\
      --merge-request=${GITLAB_MERGE_REQUEST_IID} \
      --gitlab-token="${YOUR_GITLAB_ACCESS_TOKEN}" \
      --openai-token="${YOUR_OPENAI_API_KEY}"
```

### Example in Gitlab CI Pipeline

`.gitlab-ci.yml`

```yaml
.gitlab-mr-reviewer:
  rules:
    - if: $CI_PIPELINE_SOURCE == "merge_request_event"
  image: gitlab-mr-reviewer
  script:
    - >
      ./gitlab-mr-reviewer 
      --config="./config/config.yaml"
      --project=${CI_PROJECT_ID} 
      --merge-request=${CI_MERGE_REQUEST_IID} 
      --gitlab-token="${YOUR_GITLAB_ACCESS_TOKEN}" 
      --openai-token="${YOUR_OPENAI_API_KEY}"

review-merge-request:
  stage: prepare
  extends:
    - .gitlab-mr-reviewer
  allow_failure: true
```
