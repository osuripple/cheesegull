package sql

import "github.com/jmoiron/sqlx"

// This file provides SQL migrations for MySQL.

var migrations = []func(db *sqlx.DB) error{
	migrateInitialise,
	setModes,
	// add foreign key to beatmaps, HUGE speedup
	func(db *sqlx.DB) error {
		_, err := db.Exec("ALTER TABLE beatmaps ADD FOREIGN KEY (parent_id) REFERENCES sets(set_id);")
		return err
	},
}

func migrateInitialise(db *sqlx.DB) error {
	_, err := db.Exec(`
		CREATE TABLE db_migrations(version INT NOT NULL);
		INSERT INTO db_migrations(version) VALUES (0);
		CREATE TABLE beatmaps(
			id             int NOT NULL,
			parent_id      int NOT NULL,
			diff_name      VARCHAR(255) NOT NULL,
			file_md5       VARCHAR(255) NOT NULL,
			mode           INT NOT NULL,
			bpm            FLOAT NOT NULL,
			ar             FLOAT NOT NULL,
			od             FLOAT NOT NULL,
			cs             FLOAT NOT NULL,
			hp             FLOAT NOT NULL,
			total_length   int NOT NULL,
			hit_length     int NOT NULL,
			playcount      int NOT NULL,
			passcount      int NOT NULL,
			max_combo      int NOT NULL,
			difficulty_rating FLOAT NOT NULL,
			PRIMARY KEY(id)
		) ENGINE=InnoDB;
		CREATE TABLE sets(
			set_id            int NOT NULL,
			ranked_status     int NOT NULL,
			approved_date     DATETIME NOT NULL,
			last_update       DATETIME NOT NULL,
			last_checked      DATETIME NOT NULL,
			artist            VARCHAR(1000) NOT NULL,
			title             VARCHAR(1000) NOT NULL,
			creator           VARCHAR(1000) NOT NULL,
			source            VARCHAR(1000) NOT NULL,
			tags              VARCHAR(1000) NOT NULL,
			has_video         TINYINT(1) NOT NULL,
			genre             INT NOT NULL,
			language          INT NOT NULL,
			favourites        INT NOT NULL,
			PRIMARY KEY(set_id),
			FULLTEXT(artist, title, creator, source, tags)
		) ENGINE=InnoDB;
	`)
	return err
}

func setModes(db *sqlx.DB) error {
	_, err := db.Exec(`
		ALTER TABLE sets ADD COLUMN set_modes INT NOT NULL;
	`)
	if err != nil {
		return err
	}

	var bms []struct {
		ParentID int
		Mode     int
	}
	err = db.Select(&bms, "SELECT parent_id, mode FROM beatmaps")
	if err != nil {
		return err
	}

	modes := make(map[int]int)
	for _, b := range bms {
		modes[b.ParentID] |= 1 << uint(b.Mode)
	}

	for k, v := range modes {
		_, err := db.Exec("UPDATE sets SET set_modes = ? WHERE set_id = ?", v, k)
		if err != nil {
			return err
		}
	}

	return nil
}

// AutoMigrateDB automatically updates a database to the latest version.
func AutoMigrateDB(db *sqlx.DB) error {
	var i int
	db.Get(&i, "SELECT version FROM db_migrations LIMIT 1")
	if i < len(migrations) {
		m := migrations[i:]
		for _, m := range m {
			err := m(db)
			if err != nil {
				return err
			}
		}
	}
	db.Exec("UPDATE db_migrations SET version = ?", len(migrations))
	return nil
}
