package storage

import (
	"context"
	"database/sql"

	"github.com/Masterminds/squirrel"
	"github.com/davudsafarli/twitter/src/auth"
	_ "github.com/lib/pq"
)

type postgres struct {
	db *sql.DB
	qb squirrel.StatementBuilderType
}

func NewPostgres(connstr string) (postgres, error) {
	db, err := sql.Open("postgres", connstr)
	if err != nil {
		return postgres{}, err
	}
	err = db.Ping()
	if err != nil {
		panic(err)
	}
	pg := postgres{
		db: db,
		qb: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
	return pg, nil
}

func (s postgres) CreateUser(ctx context.Context, u auth.User) (auth.User, error) {
	query := s.qb.Insert("users").
		Columns("email", "username", "password").
		Values(u.Email, u.Username, u.Password).
		Suffix("RETURNING id")

	sql, args, err := query.ToSql()
	if err != nil {
		return auth.User{}, err
	}
	row := s.db.QueryRowContext(ctx, sql, args...)
	err = row.Scan(&u.ID)
	if err != nil {
		return auth.User{}, err
	}
	return u, nil
}

func (s postgres) DeleteUser(ctx context.Context, ID int) error {
	query := s.qb.Delete("users").
		Where(squirrel.Eq{"id": ID})

	sql, args, err := query.ToSql()
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return err
	}
	return nil
}

func (s postgres) FindUser(ctx context.Context, usnm string) (auth.User, error) {
	query := s.qb.Select("id", "email", "username", "password").From("users").
		Where(squirrel.Eq{"username": usnm})

	sql, args, err := query.ToSql()
	if err != nil {
		return auth.User{}, err
	}
	row := s.db.QueryRowContext(ctx, sql, args...)

	u := auth.User{}

	if err := row.Scan(&u.ID, &u.Email, &u.Username, &u.Password); err != nil {
		return auth.User{}, err
	}

	return u, nil
}
