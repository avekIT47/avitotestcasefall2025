-- Откатываем добавление статуса CLOSED
ALTER TABLE pull_requests 
DROP CONSTRAINT IF EXISTS pull_requests_status_check;

ALTER TABLE pull_requests 
ADD CONSTRAINT pull_requests_status_check 
CHECK (status IN ('OPEN', 'MERGED'));

