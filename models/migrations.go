// THIS FILE HAS BEEN AUTOMATICALLY GENERATED
// To re-generate it, run "go generate" in the models folder.

package models

var migrations = [...]string{
	`CREATE TABLE sets(
	id INT NOT NULL,
	ranked_status TINYINT NOT NULL,
	approved_date DATETIME NOT NULL,
	last_update DATETIME NOT NULL,
	last_checked DATETIME NOT NULL,
	artist VARCHAR(1000) NOT NULL,
	title VARCHAR(1000) NOT NULL,
	creator VARCHAR(1000) NOT NULL,
	source VARCHAR(1000) NOT NULL,
	tags VARCHAR(1000) NOT NULL,
	has_video TINYINT NOT NULL,
	genre TINYINT NOT NULL,
	language TINYINT NOT NULL,
	favourites INT NOT NULL,
	set_modes TINYINT NOT NULL,
	PRIMARY KEY(id)
);
`,
	`CREATE TABLE beatmaps(
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
);`,
	`ALTER TABLE sets ADD FULLTEXT(artist, title, creator, source, tags);`,
	`ALTER TABLE beatmaps MODIFY difficulty_rating DECIMAL(20, 15);
`,
	`ALTER TABLE sets DROP INDEX artist;`,
}
