package postgres

import (
	"context"
	"os"
	"path/filepath"

	"github.com/charlie-wasp/go-masters-2025/beer-reviews/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

type Postgres struct {
	pool *pgxpool.Pool
}

func New(connstr string) (*Postgres, error) {
	pool, err := pgxpool.New(context.Background(), connstr)
	if err != nil {
		return nil, err
	}

	pg := Postgres{pool: pool}

	err = pg.applyMigrations()
	if err != nil {
		return nil, err
	}

	return &pg, nil
}

func (pg *Postgres) applyMigrations() error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	parent := filepath.Dir(wd)
	migrationsFS := os.DirFS(parent)
	goose.SetBaseFS(migrationsFS)
	goose.SetTableName("goose_migrations")

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	db := stdlib.OpenDBFromPool(pg.pool)
	if err := goose.Up(db, "migrations"); err != nil {
		return err
	}
	return db.Close()
}

func (pg *Postgres) AddReview(ctx context.Context, review models.Review) error {
	_, err := pg.pool.Exec(
		ctx,
		"INSERT INTO reviews (content, user_id, beer_id) VALUES ($1, $2, $3)",
		review.Content,
		review.UserID,
		review.BeerID,
	)

	return err
}

func (pg *Postgres) ListReviews(ctx context.Context) ([]models.Review, error) {
	return pg.executeSelectReviewsQuery(ctx, "SELECT * FROM reviews")
}

func (pg *Postgres) ListReviewsWithoutRating(ctx context.Context) ([]models.Review, error) {
	return pg.executeSelectReviewsQuery(ctx, "SELECT * FROM reviews WHERE rating IS NULL")
}

func (pg *Postgres) ListReviewsByBeerID(ctx context.Context, beerID int) ([]models.Review, error) {
	return pg.executeSelectReviewsQuery(ctx, "SELECT * FROM reviews WHERE beer_id = $1", beerID)
}

func (pg *Postgres) ListReviewsByUserID(ctx context.Context, userID int) ([]models.Review, error) {
	return pg.executeSelectReviewsQuery(ctx, "SELECT * FROM reviews WHERE user_id = $1", userID)
}

func (pg *Postgres) UpdateReview(ctx context.Context, review models.Review) error {
	_, err := pg.pool.Exec(
		ctx,
		"UPDATE reviews SET content=$1, user_id=$2, beer_id=$3, rating=$4 WHERE id = $5",
		review.Content,
		review.UserID,
		review.BeerID,
		review.Rating,
		review.ID,
	)

	return err
}

func (pg *Postgres) AvgBeerRating(ctx context.Context, beerID int) (float64, error) {
	var result float64

	err := pg.pool.QueryRow(
		ctx,
		"SELECT ROUND(AVG(rating), 2) FROM reviews WHERE beer_id = $1",
		beerID,
	).Scan(&result)
	if err != nil {
		return result, err
	}

	return result, nil
}

func (pg *Postgres) executeSelectReviewsQuery(ctx context.Context, sql string, args ...any) ([]models.Review, error) {
	rows, err := pg.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByName[models.Review])
}
