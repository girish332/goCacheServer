package cacheServer

import (
	"cacheServer/appcontext"
	"cacheServer/cache"
	"cacheServer/db"
	"log"
	"os"
	"strconv"
)

func main() {

	driver := os.Getenv("DB_DRIVER")
	postgresURI := os.Getenv("POSTGRES_URI")
	dbTimeout := os.Getenv("DB_TIMEOUT")
	timeout, _ := strconv.Atoi(dbTimeout)
	dbClient, err := db.NewPostgreSQL(driver, postgresURI)
	if err != nil {
		log.Println("error connecting Database")
		// TODO: Write the exit function
	}
	ctx := appcontext.NewContext(dbClient.DB, timeout)

	cacheServer := cache.GetCacheInstance(ctx)
	go cacheServer.Run()
}
