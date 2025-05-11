# Compchem file processor

This repository contains code for an API serving as a file processor/workflow manager for compchem invenio.

## Tasks

### Increment 1

Description:

- After this increment is completed the users file from deposition form will be automatically processed by workflows mapped to their specific files.
- The user is able to trigger workflows manually
- The deployer is able to configure what file types are processed by what workflow.

Tasks:

- [x] Finish implementing /workflows/avaiable API
- [x] Move process file to /workflows/{recordId}/start
- [x] Support file lists on /workflows/{recordId}/start
- [] Add UI for starting workflows manually - blocked for now TBD later
- [x] Analyze and add sequential ID generation on workflow based on ${RECORD_ID} and ${WORKFLOW_TYPE}


NFRs: ID generation should be based on argo workflows state, to avoid duplicating the state into the file processor


ADR - Keeping this component stateless is not completely an option, to avoid huge complexity of querying argo about processing a specific file I have for now decided to go for a schema like this:

To make the component stateless I came up with the following approaches:
1. I thought about labeling the workflows with labels such as compchem.files.processed=text1.txt,text2.txt and using argo api with selector, but this would not work well if I want to select a single file to see what is going on
2. Other option would be to assign labels such as record_id=${record_id} ${file_key}=true but I suspect this approach wouldn't scale well since argo is probably not able to index the workflows
First approach is infeasible because it won't allow precise filtering on per file basis, with second one I anticipate a problem if the system is processing large count of files

So instead of going completely stateless I want to introduce a WRITE ONCE, READ ONLY schema to maintain a record of what is being processed. This way the status of processing is in argo so the database is written once and after that is never updated, it only serves for us to do in-memory joins with argos data to present statuses of workflows to the users.
There don't need to be any atomicity guarantees because the component will be write through and if something like starting workflow fails, the user will get this information by fetching the data about their workflow/file and the service won't be able to find anything so it can assume the workflow was never started.

Database schema:
id (surrogate key), record_id, file_key, mimetype - write once, read only table after that - compchem_file
id (surrogate key), file_workflow_seq_id (sequential number for ordering of workflows of file, creation timestamp is also an option), compchem_file_id (foreign key references compchem_file), workflow_name (this serves as a id for argo_workflow uniquely identifies workflow within argo

To make the querying reliable on per workflow basis is easy, to make it reliable on per file basis without keeping state in the database we have two options:
1. File is processed only by one workflow at a time
- as a consequence the user might "lock" his file in a workflow with other files, to unlock it he will need to cancel it thereby cancelling processing about unwanted files
- good thing is that we can warn him of this
- another advantage is that this "locking" won't happen automatically since automatic processing will only ever occur over a single file, so if the files get locked together its the users choice and we warned them

2. File is processed by arbitrary amount of workflows
- we get rid of the disadvantage of locking files together in a workflow, if the user is unsatisfied and wants to reprocess a file which is already being processed in a different way they can do this
- as a consequence we will have more complex queries - we have to examine workflows for file from to the top and stop only when we arrive at the last one/first stopped one

for simplicity I will choose approach #1 it can always be adapted


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
