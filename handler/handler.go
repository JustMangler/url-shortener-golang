package handler

import (
	"context"
	"go-url-short/database"
	"log"
	"math/rand"
	"net/url"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
)

func GetLongURL(id string) (string, error) { // this is the function you want to call
	var long_url string

	err := database.Db.QueryRow(context.Background(), "SELECT long_url FROM free WHERE short_url=$1", id).Scan(&long_url)

	return long_url, err
}

func GetShortByLongURL(long_url string) (string, error) { // this is the function you want to call
	var short_url string

	err := database.Db.QueryRow(context.Background(), "SELECT short_url FROM free WHERE long_url=$1", long_url).Scan(&short_url)

	return short_url, err
}

func GetShortURL(c *fiber.Ctx) error {
	long, err := GetLongURL(c.Params("id"))
	if err != nil {
		return c.Status(200).JSON(&fiber.Map{
			"long_url": long,
		})
	}
	return c.Status(500).JSON(&fiber.Map{
		"success": false,
		"error":   "invalid short URL",
	})
}

func CreateShortURL(c *fiber.Ctx) error {
	payload := struct {
		ShortURL string `json:"short_url"`
		LongURL  string `json:"long_url"`
	}{}

	if err := c.BodyParser(&payload); err != nil {
		return c.Status(500).JSON(&fiber.Map{
			"success": false,
			"error":   "cannot parse JSON",
		})

	}
	_, err := url.ParseRequestURI(payload.LongURL)
	if err != nil {
		return c.Status(500).JSON(&fiber.Map{
			"success": false,
			"error":   "invalid URL",
		})
	}

	if payload.ShortURL == "" {
		payload.ShortURL = RandStringBytesMaskImprSrcSB(8)
	}

	query := `INSERT INTO free (short_url, long_url, visits, delete, created_at, updated_at) VALUES (@short_url, @long_url, @visits, @delete, @created_at, @updated_at)`
	args := pgx.NamedArgs{
		"short_url":  payload.ShortURL,
		"long_url":   payload.LongURL,
		"visits":     0,
		"delete":     false,
		"created_at": time.Now(),
		"updated_at": time.Now(),
	}
	log.Println(GetShortByLongURL(payload.LongURL))
	// Try to insert the row into the database, if it fails, it means the short URL is already in use
	switch _, err = database.Db.Exec(context.Background(), query, args); err {
	// Successfully inserted the row
	case nil:
		return c.Status(200).JSON(&fiber.Map{
			"short_url": payload.ShortURL,
			"long_url":  payload.LongURL,
		})
	default:
		// Check if the error is a duplicate key error on shortURL
		switch _, err := GetLongURL(payload.ShortURL); err {
		case nil:
			log.Println(err)
			return c.Status(500).JSON(&fiber.Map{
				"success": false,
				"error":   "duplicate short URL",
			})
		// Else we know the error is a duplicate key error on longURL
		default:
			short, _ := GetShortByLongURL(payload.LongURL)
			return c.Status(200).JSON(&fiber.Map{
				"short_url": short,
				"long_url":  payload.LongURL,
			})
		}
	}

}

var src = rand.NewSource(time.Now().UnixNano())

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

func RandStringBytesMaskImprSrcSB(n int) string {
	sb := strings.Builder{}
	sb.Grow(n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			sb.WriteByte(letterBytes[idx])
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return sb.String()
}
