CREATE TABLE compchem_file(
  id SERIAL PRIMARY KEY,
  file_key varchar(255) NOT NULL,
  record_id varchar(20) NOT NULL,
  mimetype varchar(50) NOT NULL,

  CONSTRAINT unique_record_file UNIQUE (record_id, file_key)
);

CREATE TABLE compchem_workflow(
  id SERIAL PRIMARY KEY,
  record_id varchar(20) NOT NULL,
  workflow_name VARCHAR(255) NOT NULL,
  workflow_record_seq_id BIGINT NOT NULL,

  CONSTRAINT unique_workflow_record UNIQUE (record_id, workflow_name, workflow_record_seq_id)
);

CREATE TABLE compchem_workflow_file(
  id SERIAL PRIMARY KEY,
  compchem_file_id BIGINT NOT NULL,
  compchem_workflow_id BIGINT NOT NULL,

  CONSTRAINT unique_workflow_file UNIQUE (compchem_file_id, compchem_workflow_id),
  CONSTRAINT compchem_file_id_fk FOREIGN KEY(compchem_file_id) REFERENCES compchem_file(id),
  CONSTRAINT compchem_workflow_id_fk FOREIGN KEY(compchem_workflow_id) REFERENCES compchem_workflow(id)
);

CREATE INDEX compchem_file_record_idx ON compchem_file(file_key, record_id);
CREATE INDEX compchem_workflow_record_idx ON compchem_workflow(record_id, workflow_name); -- add sequential number?
CREATE INDEX compchem_workflow_file_idx ON compchem_workflow_file(compchem_file_id);
CREATE INDEX compchem_workflow_workflow_idx ON compchem_workflow_file(compchem_workflow_id);
