CREATE TABLE IF NOT EXISTS GOODS(
Id Int32,
ProjectId Int32,
Name String,
Description String,
Priority Int32,
Removed Bool,
EventTime DateTime
)
ENGINE = MergeTree()
ORDER BY Id;

ALTER TABLE default.GOODS ADD INDEX IF NOT EXISTS index_id(Id) TYPE minmax GRANULARITY 8192;
ALTER TABLE default.GOODS ADD INDEX IF NOT EXISTS index_project_id(ProjectId) TYPE minmax GRANULARITY 8192;
ALTER TABLE default.GOODS ADD INDEX IF NOT EXISTS index_project_name(Name) TYPE bloom_filter GRANULARITY 8192;
