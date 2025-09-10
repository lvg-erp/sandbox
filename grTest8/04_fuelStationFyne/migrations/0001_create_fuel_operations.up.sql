CREATE TABLE fuel_operations (
     id SERIAL PRIMARY KEY,
     column_id VARCHAR(50),
     fuel_type VARCHAR(50),
     liters FLOAT,
     action VARCHAR(50),
     fill_timestamp BIGINT,
     drain_timestamp BIGINT
);