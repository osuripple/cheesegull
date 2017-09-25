CREATE TABLE beatmaps(
	id INT NOT NULL,
	parent_set_id INT NOT NULL,
	diff_name VARCHAR(1000) NOT NULL,
	file_md5 CHAR(32) NOT NULL,
	mode INT NOT NULL,
	bpm DECIMAL(10, 4) NOT NULL,
	ar DECIMAL(4, 2) NOT NULL,
	od DECIMAL(4, 2) NOT NULL,
	cs DECIMAL(4, 2) NOT NULL,
	hp DECIMAL(4, 2) NOT NULL,
	total_length INT NOT NULL,
	hit_length INT NOT NULL,
	playcount INT NOT NULL,
	passcount INT NOT NULL,
	max_combo INT NOT NULL,
	difficulty_rating INT NOT NULL,
	PRIMARY KEY(id),
	FOREIGN KEY (parent_set_id) REFERENCES sets(id)
		ON DELETE CASCADE
		ON UPDATE CASCADE
);