package alias

import (
	"database/sql"

	"github.com/MonkaKokosowa/watchalong-server/database"
	_ "modernc.org/sqlite"
)

type Alias struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Alias    string `json:"alias"`
}

func (alias *Alias) AddAlias() error {
	var existingAlias Alias
	row := database.DB.QueryRow(`SELECT * FROM aliases WHERE username = ?`, alias.Username)
	if err := row.Scan(&existingAlias.ID, &existingAlias.Username, &existingAlias.Alias); err != nil {
		if err == sql.ErrNoRows {
			if _, err := database.DB.Exec(`INSERT INTO aliases (username, alias) VALUES (?, ?)`, alias.Username, alias.Alias); err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		if _, err := database.DB.Exec(`UPDATE aliases SET alias = ? WHERE username = ?`, alias.Alias, alias.Username); err != nil {
			return err
		}
	}
	return nil
}

func GetAliases() (map[string]string, error) {
	aliases := make(map[string]string)
	rows, err := database.DB.Query(`SELECT * FROM aliases`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var alias Alias
		if err := rows.Scan(&alias.ID, &alias.Username, &alias.Alias); err != nil {
			return nil, err
		}
		aliases[alias.Username] = alias.Alias
	}

	return aliases, nil
}

func ClearAliases() error {
	if _, err := database.DB.Exec(`DELETE FROM aliases`); err != nil {
		return err
	}
	return nil
}
