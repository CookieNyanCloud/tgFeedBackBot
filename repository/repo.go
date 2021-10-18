package repository

import (
	"fmt"
	"github.com/jmoiron/sqlx"
)

type Repo struct {
	db *sqlx.DB
}
func NewRepo(db *sqlx.DB) *Repo {
	return &Repo{db: db}
}

type Users struct {
	Id int64
	Date int
}

type Operations interface {
	MakeSms(id int64, date int) error
	GetId(txt string, date int) (int64, error)
}

func (r *Repo) MakeSms(id int64, txt string, date int) error{
	fmt.Println(id,txt,date)
	query := fmt.Sprintf(`INSERT INTO tgusers (tgid, txt, date)
	VALUES($1, $2, $3)
	ON CONFLICT (tgid)
	DO
	UPDATE SET txt = EXCLUDED.txt, date = EXCLUDED.date`)
	_, err := r.db.Exec(query, id,txt,date)
	if err != nil {
		return err
	}
	return nil
}


func (r *Repo) GetId(txt string, date int) (int64, error) {
	fmt.Println(txt,date)
	var id int64
	query := fmt.Sprintf(`SELECT tgid FROM tgusers WHERE txt=$1 and date=$2`)
	err := r.db.Get(&id, query, txt,date)
	if err != nil {
		return 0,err
	}
	return id,nil
}
