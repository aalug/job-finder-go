ALTER TABLE user_skills
    ADD CONSTRAINT unique_user_skill UNIQUE (user_id, skill);

ALTER TABLE job_skills
    ADD CONSTRAINT unique_job_skill UNIQUE (job_id, skill);