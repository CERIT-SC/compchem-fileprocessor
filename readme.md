# Compchem file processor

This repository contains code for an API serving as a file processor/workflow manager for compchem invenio.


## Tasks

### Increment 1
Description:
- After this increment is completed the users file from deposition form will be automatically processed by workflows mapped to their specific files.
- The user is able to trigger workflows manually
- The deployer is able to configure what file types are processed by what workflow.

Tasks:
- [] Finish implementing /workflows/avaiable API
- [] Move process file to /workflows/{recordId}/start
- [] Support file lists on /workflows/{recordId}/start
- [] Add UI for starting workflows manually
- [] Analyze and add sequential ID generation on workflow based on ${RECORD_ID} and ${WORKFLOW_TYPE}

NFRs: ID generation should be based on argo workflows state, to avoid duplicating the state into the file processor

### Increment 2
Description:
- After this increment is complete the user should be able to see from the deposit form what is happenning in the compchem-file processor.
- They should be able to inspect logs of tasks to see their progress.
- They should be notified about successful/unsuccessful completion of their task.

Tasks:
- [] Investigate how argo event streams can be used to display workflow informating in the UI
- [] Establish some sort of standard for workflows to report their status e.g. every step should end with SUCCESS/FAILURE/PARTIAL?

### Increment 3
Description:
- After this increment is complete the user has fine control over his workflows, this means that apart from starting new workflows, they are able to stop/restart workflows as well

Tasks:
- [] FIgure out which Argo server apis to use

### Increment 4
Description
- Upon finishing this increment the API will be accessible only if the user has right permission within invenio RDM
- Argo workflows will use short lived tokens that are automatically invalidated after finishing a task

Tasks:
- [] Analyze how fine-grained the access in invenio RDM is, per file or per record?
- [] Analyze how tokens can be issued to workflows and how they can be automatically revoked
- [] Analyze if feasible to secure file processor to respect invenio access policy
    - [] SSE need to be secure, one possibility is to have invenio generate a SSE url on authenticated request
    - [] REST API needs to respect policies - user needs proper permissions to record/file to start/monitor/manage processes, workflow definitions cannot be accessed anonymously
